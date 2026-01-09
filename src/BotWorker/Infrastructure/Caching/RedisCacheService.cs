using StackExchange.Redis;
using System.Text.Json;
using Microsoft.Extensions.Logging;

namespace BotWorker.Infrastructure.Caching
{
    public class RedisCacheService : ICacheService
    {
        private readonly IConnectionMultiplexer _redis;
        private readonly IDatabase _db;
        private readonly ILogger<RedisCacheService>? _logger;

        public RedisCacheService(IConnectionMultiplexer redis, ILogger<RedisCacheService>? logger = null)
        {
            _redis = redis;
            _db = _redis.GetDatabase();
            _logger = logger;
        }

        public T? Get<T>(string key)
        {
            if (_redis == null || !_redis.IsConnected) return default;
            var value = _db.StringGet(key);
            if (value.IsNullOrEmpty) return default;
            try
            {
                return JsonSerializer.Deserialize<T>((string)value!);
            }
            catch (Exception ex)
            {
                _logger?.LogError(ex, "Redis Get error for key {Key}", key);
                return default;
            }
        }

        public bool TryGetValue<T>(string key, out T value)
        {
            if (_redis == null || !_redis.IsConnected)
            {
                value = default!;
                return false;
            }
            var val = _db.StringGet(key);
            if (val.IsNullOrEmpty)
            {
                value = default!;
                return false;
            }
            try
            {
                value = JsonSerializer.Deserialize<T>((string)val!)!;
                return true;
            }
            catch (Exception ex)
            {
                _logger?.LogError(ex, "Redis TryGetValue error for key {Key}", key);
                value = default!;
                return false;
            }
        }

        public void Set<T>(string key, T value, TimeSpan? expiration = null)
        {
            if (_redis == null || !_redis.IsConnected) return;
            var json = JsonSerializer.Serialize(value);
            _db.StringSet(key, json, expiration, When.Always, CommandFlags.None);
        }

        public void Remove(string key)
        {
            if (_redis == null || !_redis.IsConnected) return;
            _db.KeyDelete(key);
        }

        public T GetOrAdd<T>(string key, Func<T> factory, TimeSpan? expiration = null)
        {
            if (TryGetValue<T>(key, out var value))
                return value;

            value = factory();
            Set(key, value, expiration);
            return value;
        }

        public async Task SetAsync<T>(string key, T value, TimeSpan? expiration = null)
        {
            if (_redis == null || !_redis.IsConnected) return;
            var json = JsonSerializer.Serialize(value);
            await _db.StringSetAsync(key, json, expiration, When.Always, CommandFlags.None);
        }

        public async Task<T?> GetAsync<T>(string key)
        {
            if (_redis == null || !_redis.IsConnected) return default;
            try
            {
                var value = await _db.StringGetAsync(key);
                if (value.IsNullOrEmpty) return default;
                return JsonSerializer.Deserialize<T>((string)value!);
            }
            catch (Exception ex)
            {
                _logger?.LogError(ex, "Redis GetAsync error for key {Key}", key);
                return default;
            }
        }

        public async Task RemoveAsync(string key)
        {
            if (_redis == null || !_redis.IsConnected) return;
            await _db.KeyDeleteAsync(key);
        }

        public async Task<bool> ExistsAsync(string key)
        {
            if (_redis == null || !_redis.IsConnected) return false;
            return await _db.KeyExistsAsync(key);
        }

        public async Task<bool> SetIfNotExistsAsync(string key, string value, TimeSpan? expiration = null)
        {
            if (_redis == null || !_redis.IsConnected) return false;
            return await _db.StringSetAsync(key, value, expiration, When.NotExists);
        }

        public async Task<T> GetOrAddAsync<T>(string key, Func<Task<T>> factory, TimeSpan? expiration = null)
        {
            var cached = await GetAsync<T>(key);
            if (cached != null) return cached;

            var value = await factory();
            await SetAsync(key, value, expiration);
            return value;
        }

        public async Task SetStringAsync(string key, string value, TimeSpan? expiration = null)
        {
            await _db.StringSetAsync(key, value, expiration, When.Always, CommandFlags.None);
        }

        public async Task<string?> GetStringAsync(string key)
        {
            return await _db.StringGetAsync(key);
        }
    }
}
