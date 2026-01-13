using Microsoft.Extensions.Hosting;
using Microsoft.Extensions.Logging;
using Microsoft.Extensions.Configuration;
using StackExchange.Redis;
using BotWorker.Application.Messaging.Pipeline;
using BotWorker.Infrastructure.Communication.OneBot;
using System.Text.Json;

namespace BotWorker.Infrastructure.Messaging
{
    public class RedisStreamConsumer : BackgroundService
    {
        private readonly ILogger<RedisStreamConsumer> _logger;
        private readonly IConnectionMultiplexer _redis;
        private readonly MessagePipeline _pipeline;
        private readonly IOneBotApiClient _apiClient;
        private readonly string _streamName;
        private readonly string _groupName;
        private readonly string _consumerName;
        private readonly int _batchSize;
        private readonly int _blockTimeMs;
        private readonly long _startTime = DateTimeOffset.Now.ToUnixTimeSeconds();

        public RedisStreamConsumer(
            ILogger<RedisStreamConsumer> logger,
            IConnectionMultiplexer redis,
            MessagePipeline pipeline,
            IOneBotApiClient apiClient,
            IConfiguration configuration)
        {
            _logger = logger;
            _redis = redis;
            _pipeline = pipeline;
            _apiClient = apiClient;

            var workerId = configuration["BotWorker:WorkerID"] ?? "csharp-worker";
            _streamName = configuration["Redis:Streams:Default"] ?? "botmatrix:queue:default";
            _groupName = configuration["Redis:Streams:Group"] ?? "botmatrix-group";
            _consumerName = $"{workerId}-{Guid.NewGuid().ToString().Substring(0, 8)}";
            _batchSize = configuration.GetValue<int>("Redis:Streams:BatchSize", 10);
            _blockTimeMs = configuration.GetValue<int>("Redis:Streams:BlockTimeMs", 2000);
        }

        protected override async Task ExecuteAsync(CancellationToken stoppingToken)
        {
            var db = _redis.GetDatabase();

            // Try to fix MISCONF error on Redis side (self-healing)
            try
            {
                var server = _redis.GetServer(_redis.GetEndPoints()[0]);
                if (server != null)
                {
                    await server.ConfigSetAsync("stop-writes-on-bgsave-error", "no");
                    _logger.LogInformation("Attempted to set stop-writes-on-bgsave-error to no for self-healing");
                }
            }
            catch (Exception ex)
            {
                _logger.LogWarning("Failed to set stop-writes-on-bgsave-error (self-healing): {Message}", ex.Message);
            }

            // Ensure consumer group exists
            try
            {
                await db.StreamCreateConsumerGroupAsync(_streamName, _groupName, "0", createStream: true);
                _logger.LogInformation("Created consumer group {GroupName} for stream {StreamName}", _groupName, _streamName);
            }
            catch (RedisServerException ex) when (ex.Message.Contains("BUSYGROUP"))
            {
                _logger.LogInformation("Consumer group {GroupName} already exists", _groupName);
            }
            catch (Exception ex)
            {
                _logger.LogError(ex, "Failed to create consumer group {GroupName}", _groupName);
            }

            _logger.LogInformation("Started Redis Stream consumer {ConsumerName} on stream {StreamName}", _consumerName, _streamName);

            while (!stoppingToken.IsCancellationRequested)
            {
                try
                {
                    if (!_redis.IsConnected)
                    {
                        _logger.LogWarning("Redis is not connected, waiting for connection...");
                        await Task.Delay(5000, stoppingToken);
                        continue;
                    }

                    // Read from group
                    // Using ">" to read new messages
                    var messages = await db.StreamReadGroupAsync(_streamName, _groupName, _consumerName, ">", _batchSize);

                    if (messages == null || messages.Length == 0)
                    {
                        await Task.Delay(_blockTimeMs, stoppingToken);
                        continue;
                    }

                    foreach (var msg in messages)
                    {
                        var payload = msg.Values.FirstOrDefault(v => v.Name == "payload").Value;
                        if (payload.IsNull)
                        {
                            payload = msg.Values.FirstOrDefault(v => v.Name == "data").Value;
                        }

                        if (!payload.IsNull)
                        {
                            try
                            {
                                string payloadStr = payload.ToString();
                                // _logger.LogDebug("Processing message from stream {StreamName}, ID: {MsgId}. Payload: {Payload}", _streamName, msg.Id, payloadStr);

                                var botMessage = await BotMessageMapper.MapToOneBotEventAsync(payloadStr, _apiClient);
                                if (botMessage != null)
                                {
                                    // 检查消息是否过旧（例如超过 60 秒）或是在程序启动前的缓存消息
                                    var now = DateTimeOffset.Now.ToUnixTimeSeconds();
                                    if (botMessage.Time > 0)
                                    {
                                        if (botMessage.Time < _startTime)
                                        {
                                            /*
                                            _logger.LogWarning("Skipping cached message sent before startup: {MsgId}, Time: {Time}, StartTime: {StartTime}", 
                                                botMessage.MsgId, botMessage.Time, _startTime);
                                            */
                                            await db.StreamAcknowledgeAsync(_streamName, _groupName, msg.Id);
                                            continue;
                                        }
                                        
                                        if (now - botMessage.Time > 60) // 超过 60 秒的消息不再处理
                                        {
                                            /*
                                            _logger.LogWarning("Skipping expired message: {MsgId}, Time: {Time}, Age: {Age}s", 
                                                botMessage.MsgId, botMessage.Time, now - botMessage.Time);
                                            */
                                            await db.StreamAcknowledgeAsync(_streamName, _groupName, msg.Id);
                                            continue;
                                        }
                                    }

                                    /*
                                    if (botMessage.EventType != "meta_event")
                                    {
                                        _logger.LogInformation("Mapped message: {MsgId}, Type: {EventType}, From: {UserId}, Group: {GroupId}, Content: {Message}", 
                                            botMessage.MsgId, botMessage.EventType, botMessage.UserId, botMessage.GroupId, botMessage.Message);
                                    }
                                    else
                                    {
                                        _logger.LogDebug("Mapped meta event: {MsgId}", botMessage.MsgId);
                                    }
                                    */
                                    await _pipeline.ExecuteAsync(botMessage);
                                }
                                else
                                {
                                    // 检查是否是元事件或其它可以忽略的事件
                                    if (!payloadStr.Contains("\"post_type\":\"meta_event\""))
                                    {
                                        _logger.LogWarning("Failed to map payload to BotMessage: {Payload}", payloadStr);
                                    }
                                }

                                // Acknowledge
                                await db.StreamAcknowledgeAsync(_streamName, _groupName, msg.Id);
                            }
                            catch (Exception ex)
                            {
                                _logger.LogError(ex, "Error processing stream message {MsgId}", msg.Id);
                            }
                        }
                        else
                        {
                            _logger.LogWarning("Received stream message {MsgId} without payload", msg.Id);
                            await db.StreamAcknowledgeAsync(_streamName, _groupName, msg.Id);
                        }
                    }
                }
                catch (Exception ex)
                {
                    _logger.LogError(ex, "Error in Redis Stream consumer loop");
                    await Task.Delay(5000, stoppingToken);
                }
            }
        }
    }
}
