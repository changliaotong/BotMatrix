using System;
using System.Collections.Concurrent;
using System.Text.Json;
using System.Threading.Tasks;

namespace BotWorker.Core.OneBot
{
    public interface IOneBotApiClient
    {
        Task<T?> SendActionAsync<T>(string action, object? @params = null) where T : class;
        Task<object?> SendActionAsync(string action, object? @params = null);
    }

    public class OneBotApiClient : IOneBotApiClient
    {
        private readonly ConcurrentDictionary<string, TaskCompletionSource<string>> _pendingRequests = new();

        public async Task<T?> SendActionAsync<T>(string action, object? @params = null) where T : class
        {
            var response = await SendActionAsync(action, @params);
            if (response == null) return null;
            return JsonSerializer.Deserialize<T>(response.ToString() ?? "");
        }

        public Task<object?> SendActionAsync(string action, object? @params = null)
        {
            // TODO: 实现通过 WebSocket 或 HTTP 发送
            return Task.FromResult<object?>(null);
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
