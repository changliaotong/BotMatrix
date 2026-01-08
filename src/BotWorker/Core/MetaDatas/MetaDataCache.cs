using Microsoft.Data.SqlClient;
using BotWorker.Core.Database;
using BotWorker.Infrastructure.Caching;

namespace BotWorker.Core.MetaDatas
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

            var result = await GetByKeyNoCacheAsync(key1, key2);

            if (result != null && CacheService != null)
                await CacheService.SetAsync(cacheKey, result, TimeSpan.FromMinutes(5));

            return result;
        }

        /// <summary>
        /// 直接从数据库读取，绕过缓存。
        /// 适用于积分、金币等对实时性要求极高的资产数据。
        /// </summary>
        public static async Task<TDerived?> GetByKeyNoCacheAsync(object key1, object? key2 = null)
        {
            var (sql, parameters) = SqlSelect("*", key1, key2);
            return await QuerySingleAsync<TDerived>(sql, parameters);
        }

        /// <summary>
        /// 直接从数据库读取某个字段的值，绕过缓存。
        /// 适用于扣费前的余额精准检查。
        /// </summary>
        public static T GetFieldNoCache<T>(string fieldName, object id, object? id2 = null, T defaultValue = default)
        {
            var (sql, parameters) = SqlGetField<TDerived>(fieldName, id, id2);
            var result = QueryFirst<T>(sql, parameters);
            return result ?? defaultValue;
        }

        public static T Get<T>(string fieldName, object id, object? id2 = null, T? defaultValue = default)
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

            return CacheService.GetOrAdd(cacheKey, LoadFromDb, _instance.CacheTime);
        }

        /// <summary>
        /// 强制清除指定主键的缓存
        /// </summary>
        public static void RemoveCache(object id, object? id2 = null)
        {
            if (CacheService == null) return;
            var cacheKey = id2 == null ? GetCacheKey(Key, id) : GetCacheKey(Key, id, id2);
            CacheService.Remove(cacheKey);
        }

        public static async Task RemoveCacheAsync(object id, object? id2 = null)
        {
            if (CacheService == null) return;
            var cacheKey = id2 == null ? GetCacheKey(Key, id) : GetCacheKey(Key, id, id2);
            await CacheService.RemoveAsync(cacheKey);
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
        /// 终极方案：同步更新缓存中的特定字段，而不删除整个缓存对象。
        /// </summary>
        public static async Task SyncCacheFieldAsync(object id, object? id2, string fieldName, object? value)
        {
            if (CacheService == null) return;
            var cacheKey = id2 == null ? GetCacheKey(id) : GetCacheKey(id, id2);

            // 获取现有缓存
            var cached = await CacheService.GetAsync<TDerived>(cacheKey);
            if (cached != null)
            {
                // 反射设置属性值
                var prop = typeof(TDerived).GetProperty(fieldName, System.Reflection.BindingFlags.Public | System.Reflection.BindingFlags.Instance | System.Reflection.BindingFlags.IgnoreCase);
                if (prop != null && prop.CanWrite)
                {
                    try
                    {
                        var convertedValue = Convert.ChangeType(value, prop.PropertyType);
                        prop.SetValue(cached, convertedValue);
                        // 写回缓存，保持原有过期时间
                        await CacheService.SetAsync(cacheKey, cached, _instance.CacheTime);
                    }
                    catch { /* 转换失败则放弃同步，下次读取会自动从DB加载 */ }
                }
            }
        }

        public static void SyncCacheField(object id, object? id2, string fieldName, object? value)
        {
            if (CacheService == null) return;
            var cacheKey = id2 == null ? GetCacheKey(id) : GetCacheKey(id, id2);
            var cached = CacheService.Get<TDerived>(cacheKey);
            if (cached != null)
            {
                var prop = typeof(TDerived).GetProperty(fieldName, System.Reflection.BindingFlags.Public | System.Reflection.BindingFlags.Instance | System.Reflection.BindingFlags.IgnoreCase);
                if (prop != null && prop.CanWrite)
                {
                    try
                    {
                        var convertedValue = Convert.ChangeType(value, prop.PropertyType);
                        prop.SetValue(cached, convertedValue);
                        CacheService.Set(cacheKey, cached, _instance.CacheTime);
                    }
                    catch { }
                }
            }
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
