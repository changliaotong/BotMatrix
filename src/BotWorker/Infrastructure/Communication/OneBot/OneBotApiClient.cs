using System;
using System.Collections.Concurrent;
using System.Text.Json;
using System.Threading.Tasks;
using StackExchange.Redis;
using Microsoft.Extensions.Logging;

namespace BotWorker.Infrastructure.Communication.OneBot
{
    public interface IOneBotApiClient
    {
        Task<T?> SendActionAsync<T>(string platform, string selfId, string action, object? @params = null) where T : class;
        Task<object?> SendActionAsync(string platform, string selfId, string action, object? @params = null);
    }

    public class OneBotApiClient : IOneBotApiClient
    {
        private readonly IConnectionMultiplexer _redis;
        private readonly ILogger<OneBotApiClient> _logger;
        private readonly ConcurrentDictionary<string, TaskCompletionSource<string>> _pendingRequests = new();

        public OneBotApiClient(IConnectionMultiplexer redis, ILogger<OneBotApiClient> logger)
        {
            _redis = redis;
            _logger = logger;
        }

        public async Task<T?> SendActionAsync<T>(string platform, string selfId, string action, object? @params = null) where T : class
        {
            var response = await ((IOneBotApiClient)this).SendActionAsync(platform, selfId, action, @params);
            if (response == null) return null;
            return JsonSerializer.Deserialize<T>(response.ToString() ?? "");
        }

        public async Task<object?> SendActionAsync(string platform, string selfId, string action, object? @params = null)
        {
            var echo = Guid.NewGuid().ToString();
            var payload = new
            {
                type = "action",
                platform = platform,
                self_id = selfId,
                action = action,
                @params = @params,
                echo = echo
            };

            var json = JsonSerializer.Serialize(payload);
            var db = _redis.GetDatabase();
            
            _logger.LogInformation("[OneBot] Publishing action {Action} for {Platform}:{SelfId} with echo {Echo}", action, platform, selfId, echo);
            
            // Publish to botmatrix:actions
            await db.PublishAsync(RedisChannel.Literal("botmatrix:actions"), json);

            return await Task.FromResult<object?>(new { status = "sent", echo = echo });
        }

        public void HandleApiResponse(string rawJson)
        {
            try
            {
                using var doc = JsonDocument.Parse(rawJson);
                if (doc.RootElement.TryGetProperty("echo", out var echoProp))
                {
                    var echo = echoProp.GetString();
                    if (!string.IsNullOrEmpty(echo) && _pendingRequests.TryRemove(echo, out var tcs))
                    {
                        tcs.SetResult(rawJson);
                    }
                }
            }
            catch
            {
                // 忽略解析错误
            }
        }
    }
}


