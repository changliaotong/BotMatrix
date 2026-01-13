namespace BotWorker.Infrastructure.Caching
{
    public class GenericCacheRepository<T> : ICacheRepository<T>
    {
        private readonly EntityCacheHelper _cacheHelper;
        private readonly string _keyPrefix;

        public GenericCacheRepository(EntityCacheHelper cacheHelper, string keyPrefix)
        {
            _cacheHelper = cacheHelper;
            _keyPrefix = keyPrefix;
        }

        private string GetKey(string id) => $"{_keyPrefix}:{id}";

        public Task<T?> GetAsync(string id) => _cacheHelper.GetAsync<T>(GetKey(id));
        public Task SetAsync(string id, T value, TimeSpan? expiry = null) => _cacheHelper.SetAsync(GetKey(id), value, expiry);
        public Task RemoveAsync(string id) => _cacheHelper.RemoveAsync(GetKey(id));
        public Task<T?> GetOrSetAsync(string id, Func<Task<T?>> dbFetcher, TimeSpan? expiry = null) =>
            _cacheHelper.GetOrSetAsync(GetKey(id), dbFetcher, expiry);
    }

}
