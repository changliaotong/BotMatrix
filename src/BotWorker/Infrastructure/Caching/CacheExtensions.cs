namespace BotWorker.Infrastructure.Caching
{
    public static class CacheExtensions
    {
        public static T? GetValue<T>(this ICacheService? cache, string key)
        {
            if (cache == null) return default;
            return cache.TryGetValue(key, out T val) ? val : default;
        }

        public static void SafeSet<T>(this ICacheService? cache, string key, T value, TimeSpan? expire = null)
        {
            cache?.Set(key, value, expire);
        }

        public static Task<T> GetScopedAsync<T>(
            this ICacheService cache,
            string scope,
            object id,
            string key,
            Func<Task<T>> factory,
            TimeSpan? expiration = null)
        {
            string cacheKey = $"{scope}:{id}:{key}";
            return cache.GetOrAddAsync(cacheKey, factory, expiration);
        }

        public static T GetScoped<T>(
            this ICacheService cache,
            string scope,
            object id,
            string key,
            Func<T> factory,
            TimeSpan? expiration = null)
        {
            string cacheKey = $"{scope}:{id}:{key}";
            if (cache.TryGetValue<T>(cacheKey, out var value))
                return value;

            value = factory();
            cache.Set(cacheKey, value, expiration);
            return value;
        }
    }
}
