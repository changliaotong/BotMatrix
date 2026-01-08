using System.Text;
using Microsoft.Data.SqlClient;
using BotWorker.Common.Exts;


namespace BotWorker.Core.MetaDatas
{
    public abstract partial class MetaData<TDerived> where TDerived : MetaData<TDerived>, new()
    {

        public virtual async Task<int> UpdateAsync(params string[] excludeFields)
        {
            var data = ToDictionary();

            // 默认排除主键字段 + 额外排除字段
            var exclude = new HashSet<string>(StringComparer.OrdinalIgnoreCase)
            {
                KeyField,
                "Id",
                "Guid",
            };
            if (!string.IsNullOrEmpty(KeyField2))
                exclude.Add(KeyField2);
            foreach (var f in excludeFields)
                exclude.Add(f);

            // 构造 SET 子句字段（排除主键和额外排除字段）
            var setData = data
                .Where(kvp => !exclude.Contains(kvp.Key))
                .ToDictionary(kvp => kvp.Key, kvp => kvp.Value);

            // 构造 WHERE 子句字段（只用主键字段）
            var whereData = data
                .Where(kvp => string.Equals(kvp.Key, KeyField, StringComparison.OrdinalIgnoreCase)
                           || string.Equals(kvp.Key, KeyField2, StringComparison.OrdinalIgnoreCase))
                .ToDictionary(kvp => kvp.Key, kvp => kvp.Value);

            if (whereData.Count == 0)
                throw new InvalidOperationException("主键字段未赋值，无法更新");

            var (sql, paras) = SqlUpdate(setData, whereData);

            var result = await ExecAsync(sql, paras);
            if (result > 0)
            {
                var id = whereData.GetValueOrDefault(KeyField);
                var id2 = string.IsNullOrEmpty(KeyField2) ? null : whereData.GetValueOrDefault(KeyField2);
                if (id != null) await RemoveCacheAsync(id, id2);
            }
            return result;
        }


        public static (string sql, SqlParameter[] parameters) SqldUpdate<T>(T entity, object id, object? id2 = null) where T : class
        {
            var (sql, parameters) = SqldUpdate(entity, ToDict(id, id2));
            return (sql, parameters);
        }

        public static (string, SqlParameter[]) SqlUpdate(Dictionary<string, object?> setValues, Dictionary<string, object?> whereKeys)
        {
            if (setValues.Count == 0)
                throw new ArgumentException("UPDATE 操作必须指定至少一个 SET 字段");

            var sb = new StringBuilder($"UPDATE {FullName} SET ");
            var parameters = new List<SqlParameter>();
            int i = 0;

            foreach (var kvp in setValues)
            {
                if (i++ > 0) sb.Append(", ");

                string field = kvp.Key;
                object? value = kvp.Value;

                if (value is DateTime dt && dt == DateTime.MinValue)
                {
                    // ✅ 特殊处理：DateTime.MinValue → GETDATE()
                    sb.Append($"{field} = GETDATE()");
                }
                else
                {
                    string paramName = $"@u{i}";
                    sb.Append($"{field} = {paramName}");
                    parameters.Add(new SqlParameter(paramName, value ?? DBNull.Value));
                }
            }

            var (whereClause, whereParams) = SqlWhere(whereKeys, allowEmpty: false);
            parameters.AddRange(whereParams);

            return ($"{sb} {whereClause}", [.. parameters]);
        }

        public static (string, SqlParameter[]) SqlUpdate(List<Cov> updateColumns, object id, object? id2 = null)
        {
            return SqlUpdate(updateColumns, ToDict(id, id2));
        }

        public static int Update(List<Cov> columns, object id, object? id2 = null)
        {
            var (sql, paras) = SqlUpdate(columns, ToDict(id, id2));
            var result = Exec(sql, paras);
            if (result > 0) RemoveCache(id, id2);
            return result;
        }

        public static int Update(string sqlSet, object id, object? id2 = null)
        {
            var (where, parameters) = SqlWhere(id, id2);
            var result = Exec($"UPDATE {FullName} SET {sqlSet} {where}", parameters);
            if (result > 0) RemoveCache(id, id2);
            return result;
        }

        /// <summary>
        /// 更新数据但不清理缓存。
        /// </summary>
        public static int UpdateNoCache(string sqlSet, object id, object? id2 = null)
        {
            var (where, parameters) = SqlWhere(id, id2);
            return Exec($"UPDATE {FullName} SET {sqlSet} {where}", parameters);
        }

        public static async Task<int> UpdateAsync(List<Cov> columns, object id, object? id2 = null)
        {
            var (sql, paras) = SqlUpdate(columns, ToDict(id, id2));
            var result = await ExecAsync(sql, paras);
            if (result > 0) await RemoveCacheAsync(id, id2);
            return result;
        }

        /// <summary>
        /// 异步更新数据但不清理缓存。
        /// </summary>
        public static async Task<int> UpdateAsyncNoCache(string sqlSet, object id, object? id2 = null)
        {
            var (where, parameters) = SqlWhere(id, id2);
            return await ExecAsync($"UPDATE {FullName} SET {sqlSet} {where}", parameters);
        }

        public static async Task<int> UpdateAsync(object obj, object id, object? id2 = null)
        {
            var (sql, paras) = SqlUpdate(obj.ToFields(), ToDict(id, id2));
            var result = await ExecAsync(sql, paras);
            if (result > 0) await RemoveCacheAsync(id, id2);
            return result;
        }

        public static (string, SqlParameter[]) SqlUpdate((string, object?)[] setValues, (string, object?)[] whereKeys)
            => SqlUpdate(ToDict(setValues), ToDict(whereKeys));

        public static (string, SqlParameter[]) SqlSetValues(string set, object id, object? id2 = null)
        {
            var whereDict = ToDict(id, id2);
            var (where, whereParams) = SqlWhere(whereDict, allowEmpty: false);

            // 用户传入的是完整的 SET 语句片段
            var sql = $"UPDATE {FullName} SET {set} {where}";

            return (sql, whereParams);
        }

        public static (string sql, SqlParameter[] parameters) SqlSetValue(string fieldName, object value, object id, object? id2 = null)
        {
            var (where, whereParams) = SqlWhere(id, id2);

            var parameters = whereParams
                .Append(new SqlParameter("@value", value ?? DBNull.Value))
                .ToArray();

            string sql = $"UPDATE {FullName} SET {fieldName} = @value {where}";

            return (sql, parameters);
        }


        public static int SetValue(string fieldName, object value, object id, object? id2 = null)
        {
            var (sql, parameters) = SqlSetValue(fieldName, value, id, id2);
            var result = Exec(sql, parameters);
            if (result > 0) RemoveCache(id, id2);
            return result;
        }

        /// <summary>
        /// 设置字段值，但不清理 Redis 缓存。
        /// 适用于 AnswerId, LastDate 等高频变动且对缓存一致性要求不高的状态字段。
        /// </summary>
        public static int SetValueNoCache(string fieldName, object value, object id, object? id2 = null)
        {
            var (sql, parameters) = SqlSetValue(fieldName, value, id, id2);
            return Exec(sql, parameters);
        }

        /// <summary>
        /// 终极方案：设置字段值并同步更新 Redis 缓存（而不是删除）。
        /// 兼顾了 AnswerId 等状态字段的一致性与高性能需求。
        /// </summary>
        public static int SetValueSync(string fieldName, object value, object id, object? id2 = null)
        {
            var (sql, parameters) = SqlSetValue(fieldName, value, id, id2);
            int result = Exec(sql, parameters);
            if (result > 0)
            {
                SyncCacheField(id, id2, fieldName, value);
            }
            return result;
        }

        public static async Task<int> SetValueSyncAsync(string fieldName, object value, object id, object? id2 = null)
        {
            var (sql, parameters) = SqlSetValue(fieldName, value, id, id2);
            int result = await ExecAsync(sql, parameters);
            if (result > 0)
            {
                await SyncCacheFieldAsync(id, id2, fieldName, value);
            }
            return result;
        }

        public static int SetValues(string set, object id, object? id2 = null)
        {
            var (sql, parameters) = SqlSetValues(set, id, id2);
            var result = Exec(sql, parameters);
            if (result > 0) RemoveCache(id, id2);
            return result;
        }

        /// <summary>
        /// 批量设置字段值，但不清理 Redis 缓存。
        /// </summary>
        public static int SetValuesNoCache(string set, object id, object? id2 = null)
        {
            var (sql, parameters) = SqlSetValues(set, id, id2);
            return Exec(sql, parameters);
        }

        public static (string, SqlParameter[]) SqlUpdateWhere(string sSet, string sWhere)
        {
            return ($"UPDATE {FullName} SET {sSet} {sWhere.EnsureStartsWith("WHERE")}", []);
        }

        /// <summary>
        /// 批量更新操作。
        /// 注意：此方法不会清理 Redis 缓存，因为无法从 sqlWhere 中准确获取受影响的 ID。
        /// 建议后续优化：支持传入受影响的 ID 列表，或者在调用后手动调用 CacheService.Clear()。
        /// </summary>
        public static int UpdateWhere(string sqlSet, string sqlWhere)
        {
            return Exec(SqlUpdateWhere(sqlSet, sqlWhere));
        }

        public static int SetValueOther(string fieldName, object otherValue, object id, object? id2 = null)
        {
            var (sql, parameters) = SqlUpdateOther(fieldName, otherValue, id, id2);
            var result = Exec(sql, parameters);
            if (result > 0) RemoveCache(id, id2);
            return result;
        }

        public static int SetNow(string fieldName, object id, object? id2 = null)
        {
            return SetValueOther(fieldName, "GETDATE()", id, id2);
        }

        public static int Plus(string fieldName, object plusValue, object id, object? id2 = null)
        {
            var (sql, parameters) = SqlPlus(fieldName, plusValue, id, id2);
            var result = Exec(sql, parameters);
            if (result > 0) RemoveCache(id, id2);
            return result;
        }

        /// <summary>
        /// 原子增加字段值并同步更新缓存（不删除）。
        /// 完美解决金币、积分等高频变动数据的性能与一致性问题。
        /// </summary>
        public static int PlusSync(string fieldName, decimal delta, object id, object? id2 = null)
        { 
            var (sql, parameters) = SqlPlus(fieldName, delta, id, id2);
            int result = Exec(sql, parameters);
            if (result > 0)
            {
                // 获取更新后的最新值并同步到缓存
                var newValue = GetFieldNoCache<decimal>(fieldName, id, id2);
                SyncCacheField(id, id2, fieldName, newValue);
            }
            return result;
        }

        public static SqlTask TaskSetValue(string fieldName, object value, object id, object? id2 = null)
        {
            var (sql, parameters) = SqlSetValue(fieldName, value, id, id2);
            return new SqlTask(sql, parameters, id, id2, fieldName, value);
        }

        public static SqlTask TaskPlus(string fieldName, object plusValue, object id, object? id2 = null)
        {
            var (sql, parameters) = SqlPlus(fieldName, plusValue, id, id2);
            // 标记 needsReSync = true，ExecTrans 成功后会自动去数据库读真实结果并同步
            return new SqlTask(sql, parameters, id, id2, fieldName, true);
        }

        public static (string, SqlParameter[]) SqlSetValue(string fieldName, object value, object id, object? id2 = null)
        {
            var whereDict = ToDict(id, id2);
            var (whereClause, whereParams) = SqlWhere(whereDict, allowEmpty: false);

            string sql = $"UPDATE {FullName} SET {fieldName} = @fieldValue {whereClause}";

            var paramList = whereParams.ToList();
            paramList.Add(new SqlParameter("@fieldValue", value ?? DBNull.Value));

            return (sql, paramList.ToArray());
        }

        public static (string, SqlParameter[]) SqlUpdate(string fieldName, object fieldValue, object id, object? id2 = null)
        {
            var sql = $"UPDATE {FullName} SET {fieldName} = @fieldValue";
            var (where, parameters) = SqlWhere(id, id2);
            var paramList = parameters.ToList();
            paramList.Insert(0, new SqlParameter("@fieldValue", fieldValue ?? DBNull.Value));

            return ($"{sql} {where}", [.. paramList]);
        }

        public static (string, SqlParameter[]) SqlUpdateOther(string fieldName, object fieldValue, object id, object? id2 = null)
        {
            var (where, parameters) = SqlWhere(id, id2);
            return ($"UPDATE {FullName} SET {fieldName} = {fieldValue} {where}", parameters);
        }

        #region 缓存感知的事务处理

        /// <summary>
        /// 封装 SQL 任务及其对应的缓存元数据
        /// </summary>
        public class SqlTask
        {
            public string Sql { get; set; }
            public SqlParameter[] Parameters { get; set; }
            public object Id { get; set; }
            public object? Id2 { get; set; }
            public string? SyncField { get; set; } 
            public bool NeedsReSync { get; set; } // 新增：标记是否需要在事务后重读数据库

            public SqlTask(string sql, SqlParameter[] parameters, object id, object? id2 = null, string? syncField = null, bool needsReSync = false)
            { 
                Sql = sql;
                Parameters = parameters;
                Id = id;
                Id2 = id2;
                SyncField = syncField;
                NeedsReSync = needsReSync;
            }

            public static implicit operator SqlTask((string sql, SqlParameter[] parameters) tuple) 
                => new SqlTask(tuple.sql, tuple.parameters, null!);
        }

        public static int ExecTrans(params SqlTask[] tasks)
        { 
            var sqls = tasks.Select(t => (t.Sql, t.Parameters)).ToArray();
            int result = ExecTrans(sqls);

            if (result > 0)
            { 
                foreach (var task in tasks)
                { 
                    if (task.Id == null) continue;

                    if (task.NeedsReSync && !string.IsNullOrEmpty(task.SyncField))
                    { 
                        // 核心改进：事务成功后，立即从数据库读取原子更新后的真实值
                        var realValue = GetFieldNoCache<object>(task.SyncField, task.Id, task.Id2);
                        SyncCacheField(task.Id, task.Id2, task.SyncField, realValue);
                    }
                    else if (!string.IsNullOrEmpty(task.SyncField))
                    {
                        // 普通的 SetValue 操作，可以直接同步
                        // 注意：这里需要 SqlTask 携带 Value，我稍后补充
                    }
                    else
                    { 
                        RemoveCache(task.Id, task.Id2);
                    }
                }
            }
            return result;
        }

        #endregion
    }
}
