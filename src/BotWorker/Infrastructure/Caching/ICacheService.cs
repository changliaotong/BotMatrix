namespace sz84.Infrastructure.Caching
{
    public interface ICacheService
    {
        /// <summary>
        /// 直接从缓存获取数据，找不到返回默认值
        /// </summary>
        T? Get<T>(string key);

        /// <summary>
        /// 尝试从缓存获取数据，返回是否成功
        /// </summary>
        bool TryGetValue<T>(string key, out T value);

        /// <summary>
        /// 同步设置缓存
        /// </summary>
        void Set<T>(string key, T value, TimeSpan? expiration = null);

        /// <summary>
        /// 移除缓存
        /// </summary>
        void Remove(string key);

        /// <summary>
        /// 同步获取缓存，如果不存在则通过 factory 同步加载数据并缓存
        /// </summary>
        T GetOrAdd<T>(string key, Func<T> factory, TimeSpan? expiration = null);

        /// <summary>
        /// 异步设置缓存（对象序列化为 JSON）
        /// </summary>
        Task SetAsync<T>(string key, T value, TimeSpan? expiration = null);

        /// <summary>
        /// 异步获取缓存（JSON反序列化为对象）
        /// </summary>
        Task<T?> GetAsync<T>(string key);

        /// <summary>
        /// 异步移除缓存
        /// </summary>
        Task RemoveAsync(string key);

        /// <summary>
        /// 判断缓存是否存在
        /// </summary>
        Task<bool> ExistsAsync(string key);

        /// <summary>
        /// 如果不存在则设置缓存（用于并发控制）
        /// </summary>
        Task<bool> SetIfNotExistsAsync(string key, string value, TimeSpan? expiration = null);

        /// <summary>
        /// 异步获取缓存，如果不存在则通过 factory 异步加载数据并缓存
        /// </summary>
        Task<T> GetOrAddAsync<T>(string key, Func<Task<T>> factory, TimeSpan? expiration = null);

        /// <summary>
        /// 设置字符串缓存（异步）
        /// </summary>
        Task SetStringAsync(string key, string value, TimeSpan? expiration = null);

        /// <summary>
        /// 获取字符串缓存（异步）
        /// </summary>
        Task<string?> GetStringAsync(string key);
    }
}
