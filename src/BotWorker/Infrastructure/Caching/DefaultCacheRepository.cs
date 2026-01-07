namespace sz84.Infrastructure.Caching
{
    public class DefaultCacheRepository<T>(EntityCacheHelper cache) : ICacheRepository<T> where T : class, new()
    {
        private readonly EntityCacheHelper _cache = cache;

        private static string GetCacheKey(string id) => $"entity:{typeof(T).Name}:{id}";

        public Task<T?> GetAsync(string id) => _cache.GetAsync<T>(GetCacheKey(id));

        public Task SetAsync(string id, T entity, TimeSpan? expire = null)
            => _cache.SetAsync(GetCacheKey(id), entity, expire);

        public Task RemoveAsync(string id)
            => _cache.RemoveAsync(GetCacheKey(id));

        public async Task<T?> GetOrSetAsync(string id, Func<Task<T?>> dbFetcher, TimeSpan? expiry = null)
        {
            var key = GetCacheKey(id);
            var cached = await _cache.GetAsync<T>(key);
            if (cached != null)
            {
                _cache.Logger?.LogInformation($"[Cache Hit] {key}");
                return cached;
            }

            _cache.Logger?.LogInformation($"[Cache Miss] {key}, loading from DB...");
            var fetched = await dbFetcher();
            if (fetched != null)
            {
                await _cache.SetAsync(key, fetched, expiry);
                _cache.Logger?.LogInformation($"[Cache Set] {key}");
            }
            return fetched;
        }
    }
}
