using System;
using System.Collections.Concurrent;
using System.Threading.Tasks;

namespace BotWorker.Services
{
    public interface IGlobalCacheService
    {
        Task<T?> GetAsync<T>(string key);
        Task SetAsync<T>(string key, T value, TimeSpan? expiry = null);
        Task RemoveAsync(string key);
    }

    public class GlobalCacheService : IGlobalCacheService
    {
        private readonly ConcurrentDictionary<string, object> _cache = new();

        public Task<T?> GetAsync<T>(string key)
        {
            if (_cache.TryGetValue(key, out var value))
            {
                return Task.FromResult((T?)value);
            }
            return Task.FromResult(default(T));
        }

        public Task SetAsync<T>(string key, T value, TimeSpan? expiry = null)
        {
            if (value != null)
            {
                _cache[key] = value;
            }
            return Task.CompletedTask;
        }

        public Task RemoveAsync(string key)
        {
            _cache.TryRemove(key, out _);
            return Task.CompletedTask;
        }
    }
}


