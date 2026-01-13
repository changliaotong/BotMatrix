using LazyCache;

namespace BotWorker.Infrastructure.Caching
{
    public static class LazyCacheHelper
    {
        private static readonly CachingService _cache = new CachingService();

        /// <summary>
        /// 从缓存获取数据，如果不存在则调用 factory 生成并缓存
        /// </summary>
        public static T GetOrAdd<T>(string key, Func<T> factory, TimeSpan? expire = null)
        {
            if (expire.HasValue)
                return _cache.GetOrAdd(key, factory, expire.Value);
            else
                return _cache.GetOrAdd(key, factory);
        }

        /// <summary>
        /// 直接设置缓存值，带可选过期时间
        /// </summary>
        public static void Set<T>(string key, T value, TimeSpan? expire = null)
        {
            if (expire.HasValue)
                _cache.Add(key, value, expire.Value);
            else
                _cache.Add(key, value);
        }

        /// <summary>
        /// 尝试从缓存取值，没有返回 false
        /// </summary>
        public static bool TryGet<T>(string key, out T? value)
        {
            // LazyCache 没有 TryGet 直接接口，这里用 Get 来模拟
            value = _cache.Get<T>(key);

            // 如果是引用类型，null 就说明没取到
            if (value == null)
                return false;

            // 如果是值类型，判断是否等于默认值
            // 这里假设默认值表示没缓存，某些场景可能需调整判断逻辑
            if (Equals(value, default(T)))
                return false;

            return true;
        }

        /// <summary>
        /// 删除缓存
        /// </summary>
        public static void Remove(string key)
        {
            _cache.Remove(key);
        }
    }

}
