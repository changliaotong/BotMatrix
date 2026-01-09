using System.Data;

namespace BotWorker.Infrastructure.Persistence.ORM
{
    public abstract partial class MetaData<TDerived> where TDerived : MetaData<TDerived>, new()
    {
        /// <summary>
        /// 自动判断新增或更新并保存
        /// </summary>
        /// <param name="trans">事务对象</param>
        /// <returns>受影响行数</returns>
        public virtual async Task<int> SaveAsync(IDbTransaction? trans = null)
        {
            var keyValues = GetKeyValues();
            bool isNew = true;

            // 检查主键是否有有效值
            bool hasValidKey = false;
            foreach (var kv in keyValues)
            {
                if (kv.Value != null && !kv.Value.Equals(0) && !kv.Value.Equals(Guid.Empty) && !kv.Value.Equals(string.Empty) && !kv.Value.Equals(DBNull.Value))
                {
                    hasValidKey = true;
                    break;
                }
            }

            if (hasValidKey)
            {
                // 如果主键有值，尝试检查数据库中是否存在
                // 注意：这里需要支持复合主键
                var keys = new Dictionary<string, object?>();
                foreach (var kv in keyValues)
                {
                    keys.Add(kv.Name, kv.Value);
                }
                
                var (sql, parameters) = SqlExists(FullName, keys);
                var result = await QueryScalarAsync<object>(sql, trans, parameters);
                if (result != null && result != DBNull.Value)
                {
                    isNew = false;
                }
            }

            if (isNew)
            {
                return await InsertAsync(trans);
            }
            else
            {
                return await UpdateAsync(trans);
            }
        }
    }
}
