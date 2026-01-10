namespace BotWorker.Infrastructure.Caching
{
    public abstract class CacheRepositoryBase<T>(EntityCacheHelper cacheHelper) : ICacheRepository<T>
    {
        protected readonly EntityCacheHelper _cacheHelper = cacheHelper;

        // 子类必须实现，定义缓存 Key 规则
        protected abstract string GetCacheKey(string id);

        public Task<T?> GetAsync(string id) =>
            _cacheHelper.GetAsync<T>(GetCacheKey(id));

        public Task SetAsync(string id, T value, TimeSpan? expiry = null) =>
            _cacheHelper.SetAsync(GetCacheKey(id), value, expiry);

        public Task RemoveAsync(string id) =>
            _cacheHelper.RemoveAsync(GetCacheKey(id));

        public Task<T?> GetOrSetAsync(string id, Func<Task<T?>> dbFetcher, TimeSpan? expiry = null) =>
            _cacheHelper.GetOrSetAsync(GetCacheKey(id), dbFetcher, expiry);
    }

}
