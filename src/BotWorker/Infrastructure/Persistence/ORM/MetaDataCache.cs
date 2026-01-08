using Microsoft.Data.SqlClient;
using BotWorker.Infrastructure.Persistence.Database;
using BotWorker.Infrastructure.Caching;

namespace BotWorker.Infrastructure.Persistence.ORM
{
    public abstract partial class MetaData<TDerived> where TDerived : MetaData<TDerived>, new()
    {
        public static ICacheService? CacheService { get; set; } // 由外部初始化注入

        protected static string GetCacheKey(params object[] keys)
        {
            string keyPart = string.Join("_", keys);
            return $"MetaData:{FullName}:Id:{keyPart}";
        }

        public static async Task<TDerived?> GetByKeyAsync(object key1, object? key2 = null)
        {
            var cacheKey = key2 == null ? GetCacheKey(key1) : GetCacheKey(key1, key2);

            if (CacheService != null)
            {
                var cached = await CacheService.GetAsync<TDerived>(cacheKey);
                if (cached != null)
                    return cached;
            }

            var (sql, parameters) = SqlSelect("*", key1, key2);
            var result = await QuerySingleAsync<TDerived>(sql, null, parameters);

            if (result != null && CacheService != null)
                await CacheService.SetAsync(cacheKey, result, TimeSpan.FromMinutes(5));

            return result;
        }

        public static T GetCached<T>(string fieldName, object id, object? id2 = null, T? defaultValue = default)
        {
            T LoadFromDb()
            {
                var (sql, paras) = SqlGet(fieldName, id, id2);
                return QueryScalar<T>(sql, paras)!;               
            }

            if (CacheService == null)
            {
                return LoadFromDb();
            }

            var cacheKey = id2 == null ? GetCacheKey(fieldName, id) : GetCacheKey(fieldName, id, id2);

            return CacheService.GetOrAdd(cacheKey, LoadFromDb, TimeSpan.FromMinutes(5));
        }

        public static async Task<T> GetAsync<T>(string fieldName, object id, object? id2 = null, T defaultValue = default!) where T : struct
        {
            async Task<T> LoadFromDbAsync()
            {
                var (sql, paras) = SqlGet(fieldName, id, id2);
                var raw = await ExecScalarAsync<T>(sql, paras);
                return SqlHelper.ConvertValue<T>(raw, defaultValue);
            }

            if (CacheService == null)
            {
                return await LoadFromDbAsync();
            }

            var cacheKey = id2 == null ? GetCacheKey(fieldName, id) : GetCacheKey(fieldName, id, id2);

            return await CacheService.GetOrAddAsync(cacheKey, LoadFromDbAsync, TimeSpan.FromMinutes(5));
        }


        public static async Task InvalidateCacheAsync(object key1, object? key2 = null)
        {
            if (CacheService == null) return;
            var cacheKey = key2 == null ? GetCacheKey(key1) : GetCacheKey(key1, key2);
            await CacheService.RemoveAsync(cacheKey);

            // TODO: 同时删除关联的列表缓存，如果有缓存列表Key管理，需要调用
        }

        /// <summary>
        /// 失效特定字段的缓存
        /// </summary>
        public static void InvalidateFieldCache(string fieldName, object key1, object? key2 = null)
        {
            if (CacheService == null) return;
            var cacheKey = key2 == null ? GetCacheKey(fieldName, key1) : GetCacheKey(fieldName, key1, key2);
            CacheService.Remove(cacheKey);
        }

        /// <summary>
        /// 异步失效特定字段的缓存
        /// </summary>
        public static async Task InvalidateFieldCacheAsync(string fieldName, object key1, object? key2 = null)
        {
            if (CacheService == null) return;
            var cacheKey = key2 == null ? GetCacheKey(fieldName, key1) : GetCacheKey(fieldName, key1, key2);
            await CacheService.RemoveAsync(cacheKey);
        }

        // 写操作示例（新增/更新）：
        public async Task SaveAsync(ICacheService cacheService)
        {
            // 调用你的数据库保存逻辑

            // 操作数据库成功后，更新缓存
            var keys = GetKeyValues().Select(k => k.Value).ToArray();
            var cacheKey = GetCacheKey(keys);
            await cacheService.SetAsync(cacheKey, this, TimeSpan.FromMinutes(5));

            // 删除相关列表缓存
        }
    }
}

