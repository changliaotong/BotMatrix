using Microsoft.Extensions.Hosting;
using Microsoft.Extensions.Logging;
using Microsoft.Extensions.Configuration;
using StackExchange.Redis;
using BotWorker.Application.Messaging.Pipeline;
using BotWorker.Infrastructure.Communication.OneBot;
using System.Text.Json;
using Serilog;

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
            Console.WriteLine("DEBUG: RedisStreamConsumer constructor started");
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
            
            Log.Information("RedisStreamConsumer initialized. Stream: {StreamName}, Group: {GroupName}", _streamName, _groupName);
            Console.WriteLine($"DEBUG: RedisStreamConsumer initialized. Stream: {_streamName}");
        }

        protected override async Task ExecuteAsync(CancellationToken stoppingToken)
        {
            Console.WriteLine("DEBUG: RedisStreamConsumer.ExecuteAsync starting");
            Log.Information("RedisStreamConsumer.ExecuteAsync started");
            
            IDatabase db;
            try {
                db = _redis.GetDatabase();
                Log.Information("Successfully got Redis database instance");
            } catch (Exception ex) {
                Log.Error(ex, "Failed to get Redis database instance");
                return;
            }

            // Ensure consumer group exists
            try
            {
                Log.Information("Ensuring consumer group {GroupName} exists for stream {StreamName}", _groupName, _streamName);
                await db.StreamCreateConsumerGroupAsync(_streamName, _groupName, "0", createStream: true);
                Log.Information("Created consumer group {GroupName} for stream {StreamName}", _groupName, _streamName);
            }
            catch (RedisServerException ex) when (ex.Message.Contains("BUSYGROUP"))
            {
                Log.Information("Consumer group {GroupName} already exists", _groupName);
            }
            catch (Exception ex)
            {
                Log.Error(ex, "Failed to create consumer group {GroupName}", _groupName);
            }

            Log.Information("Started Redis Stream consumer {ConsumerName} on stream {StreamName}", _consumerName, _streamName);
            Console.WriteLine($"DEBUG: Redis Stream consumer {_consumerName} started");

            while (!stoppingToken.IsCancellationRequested)
            {
                try
                {
                    if (!_redis.IsConnected)
                    {
                        Log.Warning("Redis is not connected, waiting for connection...");
                        await Task.Delay(5000, stoppingToken);
                        continue;
                    }

                    // Read from group
                    // 1. First, check for pending messages that weren't ACKed (e.g. if the worker crashed)
                    // We use "0" as the ID to get pending messages for THIS consumer
                    var pendingMessages = await db.StreamReadGroupAsync(_streamName, _groupName, _consumerName, "0", _batchSize);
                    
                    // 2. Then, read NEW messages
                    var newMessages = await db.StreamReadGroupAsync(_streamName, _groupName, _consumerName, ">", _batchSize);

                    var messages = (pendingMessages ?? Array.Empty<StreamEntry>())
                        .Concat(newMessages ?? Array.Empty<StreamEntry>())
                        .ToArray();

                    if (messages.Length == 0)
                    {
                        await Task.Delay(_blockTimeMs, stoppingToken);
                        continue;
                    }

                    Log.Information("[RedisStream] Received {Count} messages (Pending: {Pending}, New: {New})", 
                        messages.Length, pendingMessages?.Length ?? 0, newMessages?.Length ?? 0);
                    Console.WriteLine($"DEBUG: [RedisStream] Received {messages.Length} messages");

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
                                Log.Information("Processing message {MsgId}. Payload length: {Length}", msg.Id, payloadStr.Length);

                                var botMessage = await BotMessageMapper.MapToOneBotEventAsync(payloadStr, _apiClient);
                                if (botMessage != null)
                                {
                                    Log.Information("Mapped message: {MsgId}, Type: {EventType}, From: {UserId}, Group: {GroupId}", 
                                        botMessage.MsgId, botMessage.EventType, botMessage.UserId, botMessage.GroupId);
                                    
                                    await _pipeline.ExecuteAsync(botMessage);
                                }
                                else
                                {
                                    // 检查是否是元事件或其它可以忽略的事件
                                    if (!payloadStr.Contains("\"post_type\":\"meta_event\""))
                                    {
                                        Log.Warning("Failed to map payload to BotMessage: {Payload}", payloadStr);
                                    }
                                }

                                // Acknowledge
                                await db.StreamAcknowledgeAsync(_streamName, _groupName, msg.Id);
                            }
                            catch (Exception ex)
                            {
                                Log.Error(ex, "Error processing stream message {MsgId}", msg.Id);
                            }
                        }
                        else
                        {
                            Log.Warning("Received stream message {MsgId} without payload", msg.Id);
                            await db.StreamAcknowledgeAsync(_streamName, _groupName, msg.Id);
                        }
                    }
                }
                catch (Exception ex)
                {
                    Log.Error(ex, "Error in Redis Stream consumer loop");
                    await Task.Delay(5000, stoppingToken);
                }
            }
        }
    }
}
