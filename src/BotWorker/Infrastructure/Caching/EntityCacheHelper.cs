using Serilog;
using StackExchange.Redis;
using System.Text.Json;

namespace BotWorker.Infrastructure.Caching
{
    
    /// <summary>
    /// 缓存操作核心类
    /// </summary>
    public class EntityCacheHelper
    {
        private readonly IConnectionMultiplexer _redis;
        private readonly IDatabase _db;
        public ICacheLogger? Logger;

        private readonly SemaphoreSlim _lock = new(1, 1);

        // 失效通知事件
        public event Func<string, Task>? OnCacheRemoved;

        public EntityCacheHelper(string redisConnectionString, ICacheLogger? logger = null)
        {
            if (string.IsNullOrWhiteSpace(redisConnectionString))
                throw new ArgumentNullException(nameof(redisConnectionString));

            Logger = logger;
            _redis = ConnectionMultiplexer.Connect(redisConnectionString);
            _db = _redis.GetDatabase();

            // 订阅 Key 过期事件（需 Redis 配置支持）
            var subscriber = _redis.GetSubscriber();
            subscriber.Subscribe(RedisChannel.Pattern("__keyevent@0__:expired"), async (channel, message) =>
            {
                if (OnCacheRemoved != null)
                    await OnCacheRemoved(message!);
            });
        }

        public async Task<T?> GetAsync<T>(string key)
        {
            try
            {
                var val = await _db.StringGetAsync(key);
                if (val.IsNullOrEmpty) return default;
                return JsonSerializer.Deserialize<T>(val.AsString()!);
            }
            catch (Exception ex)
            {
                Logger?.LogError($"GetAsync<{typeof(T).Name}> failed key={key}", ex);
                return default;
            }
        }

        public async Task<bool> SetAsync<T>(string key, T value, TimeSpan? expiry = null)
        {
            try
            {
                if (value == null) return false;
                var json = JsonSerializer.Serialize(value);
                return await _db.StringSetAsync(key, json, expiry, When.Always);
            }
            catch (Exception ex)
            {
                Logger?.LogError($"SetAsync<{typeof(T).Name}> failed key={key}", ex);
                return false;
            }
        }

        public async Task<bool> RemoveAsync(string key)
        {
            try
            {
                var removed = await _db.KeyDeleteAsync(key);

                // 触发失效事件（同步调用，确保事件触发）
                if (removed && OnCacheRemoved != null)
                {
                    await OnCacheRemoved(key);
                }

                return removed;
            }
            catch (Exception ex)
            {
                Logger?.LogError($"RemoveAsync failed key={key}", ex);
                return false;
            }
        }

        public async Task<bool> ExistsAsync(string key)
        {
            try
            {
                return await _db.KeyExistsAsync(key);
            }
            catch (Exception ex)
            {
                Logger?.LogError($"ExistsAsync failed key={key}", ex);
                return false;
            }
        }

        /// <summary>
        /// 带缓存击穿保护的GetOrSet方法，防止高并发下数据库压力
        /// </summary>
        public async Task<T?> GetOrSetAsync<T>(string key, Func<Task<T?>> dbFetcher, TimeSpan? expiry = null)
        {
            if (dbFetcher == null) throw new ArgumentNullException(nameof(dbFetcher));

            try
            {
                // 先读缓存
                var cached = await GetAsync<T>(key);
                if (cached != null) return cached;

                // 加锁防止缓存击穿
                await _lock.WaitAsync();
                try
                {
                    // 双重检测
                    cached = await GetAsync<T>(key);
                    if (cached != null) return cached;

                    var dbData = await dbFetcher();
                    if (dbData != null)
                    {
                        await SetAsync(key, dbData, expiry ?? TimeSpan.FromHours(1));
                    }
                    return dbData;
                }
                finally
                {
                    _lock.Release();
                }
            }
            catch (Exception ex)
            {
                Logger?.LogError($"GetOrSetAsync<{typeof(T).Name}> failed key={key}", ex);
                return await dbFetcher();
            }
        }
    }
    
}
