using System.Data;
using System.Text;
using BotWorker.Infrastructure.Extensions;
using BotWorker.Infrastructure.Persistence.Database;

namespace BotWorker.Infrastructure.Persistence.ORM
{
    public abstract partial class MetaData<TDerived> where TDerived : MetaData<TDerived>, new()
    {

        public virtual async Task<int> UpdateAsync(IDbTransaction? trans = null, params string[] excludeFields)
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

            return await ExecAsync(sql, trans, paras);
        }


        public static (string sql, IDataParameter[] paras) SqlUpdate<T>(T entity, object id, object? id2 = null) where T : class
        {
            var (sql, paras) = SqlUpdate(entity, ToDict(id, id2));
            return (sql, paras);
        }

        public static (string sql, IDataParameter[] paras) SqlUpdate(string fieldName, object value, object id, object? id2 = null)
        {
            var setValues = new Dictionary<string, object?> { { fieldName, value } };
            return SqlUpdate(setValues, ToDict(id, id2));
        }

        public static (string sql, IDataParameter[] paras) SqlUpdate(Dictionary<string, object?> setValues, Dictionary<string, object?> whereKeys)
        {
            if (setValues.Count == 0)
                throw new ArgumentException("UPDATE 操作必须指定至少一个 SET 字段");

            var sb = new StringBuilder($"UPDATE {FullName} SET ");
            var parameters = new List<IDataParameter>();
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
                    parameters.Add(CreateParameter(paramName, value));
                }
            }

            var (whereClause, whereParams) = SqlWhere(whereKeys, allowEmpty: false);
            parameters.AddRange(whereParams);

            return ($"{sb} {whereClause}", [.. parameters]);
        }

        public static (string sql, IDataParameter[] paras) SqlUpdate(List<Cov> updateColumns, object id, object? id2 = null)
        {
            return SqlUpdate(updateColumns, ToDict(id, id2));
        }

        public static int Update(List<Cov> columns, object id, object? id2 = null)
            => Update(columns, id, id2, null);

        public static int Update(List<Cov> columns, object id, object? id2, IDbTransaction? trans)
            => UpdateAsync(columns, id, id2, trans).GetAwaiter().GetResult();

        public static async Task<int> UpdateAsync(List<Cov> columns, object id, object? id2 = null, IDbTransaction? trans = null)
        {
            var (sql, paras) = SqlUpdate(columns, ToDict(id, id2));
            return await ExecAsync(sql, trans, paras);
        }

        public static async Task<int> UpdateObjectAsync(object obj, object id, object? id2 = null, IDbTransaction? trans = null)
        {
            var (sql, paras) = SqlUpdate(obj.ToFields(), ToDict(id, id2));
            return await ExecAsync(sql, trans, paras);
        }

        public static int Update(string sSet, object id, object? id2 = null, IDbTransaction? trans = null)
            => UpdateAsync(sSet, id, id2, trans).GetAwaiter().GetResult();

        public static async Task<int> UpdateAsync(string sSet, object id, object? id2 = null, IDbTransaction? trans = null)
        {
            var (sqlWhere, parameters) = SqlWhere(id, id2);
            var sql = $"UPDATE {FullName} SET {sSet} {sqlWhere}";
            return await ExecAsync(sql, trans, parameters);
        }

        public static int UpdateWhere(string sSet, string sWhere, IDbTransaction? trans = null)
            => UpdateWhereAsync(sSet, sWhere, trans).GetAwaiter().GetResult();

        public static async Task<int> UpdateWhereAsync(string sSet, string sWhere, IDbTransaction? trans = null)
        {
            var sql = $"UPDATE {FullName} SET {sSet} {sWhere.EnsureStartsWith("WHERE")}";
            return await ExecAsync(sql, trans);
        }

        public static (string sql, IDataParameter[] paras) SqlUpdateWhere(string sSet, string sWhere)
        {
            var sql = $"UPDATE {FullName} SET {sSet} {sWhere.EnsureStartsWith("WHERE")}";
            return (sql, Array.Empty<IDataParameter>());
        }

        public static (string sql, IDataParameter[] paras) SqlUpdate((string, object?)[] setValues, (string, object?)[] whereKeys)
            => SqlUpdate(ToDict(setValues), ToDict(whereKeys));

        public static (string sql, IDataParameter[] paras) SqlSetValues(string set, object id, object? id2 = null)
        {
            var whereDict = ToDict(id, id2);
            var (where, paras) = SqlWhere(whereDict, allowEmpty: false);

            // 用户传入的是完整的 SET 语句片段
            var sql = $"UPDATE {FullName} SET {set} {where}";

            return (sql, paras);
        }

        public static (string sql, IDataParameter[] paras) SqlSetValue(string fieldName, object value, object id, object? id2 = null)
        {
            var (where, whereParams) = SqlWhere(id, id2);

            var paras = whereParams
                .Append(CreateParameter("@value", value))
                .ToArray();

            string sql = $"UPDATE {FullName} SET {fieldName} = @value {where}";

            return (sql, paras);
        }


        public static int SetValue(string fieldName, object value, object id, object? id2 = null)
            => SetValueAsync(fieldName, value, id, id2).GetAwaiter().GetResult();

        public static async Task<int> SetValueAsync(string fieldName, object value, object id, object? id2 = null, IDbTransaction? trans = null)
        {
            var (sql, parameters) = SqlSetValue(fieldName, value, id, id2);
            return await ExecAsync(sql, trans, parameters);
        }

        public static int SetValue(string fieldName, object value, object id, object? id2, IDbTransaction? trans)
            => SetValueAsync(fieldName, value, id, id2, trans).GetAwaiter().GetResult();

        public static int SetValues(string set, object id, object? id2 = null)
            => SetValuesAsync(set, id, id2).GetAwaiter().GetResult();

        public static async Task<int> SetValuesAsync(string set, object id, object? id2 = null, IDbTransaction? trans = null)
        {
            var (sql, parameters) = SqlSetValues(set, id, id2);
            return await ExecAsync(sql, trans, parameters);
        }

        public static int SetValues(string set, object id, object? id2, IDbTransaction? trans)
            => SetValuesAsync(set, id, id2, trans).GetAwaiter().GetResult();

        public static int SetValueOther(string fieldName, object otherValue, object id, object? id2 = null)
        {
            var (sql, parameters) = SqlUpdateOther(fieldName, otherValue, id, id2);
            return Exec(sql, parameters);
        }

        public static int SetNow(string fieldName, object id, object? id2 = null)
        {
            return SetValueOther(fieldName, "GETDATE()", id, id2);
        }

        public static int Plus(string fieldName, object plusValue, object id, object? id2 = null)
            => PlusAsync(fieldName, plusValue, id, id2).GetAwaiter().GetResult();

        public static async Task<int> PlusAsync(string fieldName, object plusValue, object id, object? id2 = null, IDbTransaction? trans = null)
        {
            var (sql, parameters) = SqlPlus(fieldName, plusValue, id, id2);
            return await ExecAsync(sql, trans, parameters);
        }

        public static (string sql, IDataParameter[] paras) SqlPlus(string fieldName, object plusValue, object id, object? id2 = null)
        {
            var whereDict = ToDict(id, id2);
            var (whereClause, whereParams) = SqlWhere(whereDict, allowEmpty: false);

            string sql = $"UPDATE {FullName} SET {fieldName} = ISNULL({fieldName}, 0) + @plusValue {whereClause}";

            var paramList = whereParams.ToList<IDataParameter>();
            paramList.Add(CreateParameter("@plusValue", plusValue ?? 0));

            return (sql, paramList.ToArray());
        }

        public static (string sql, IDataParameter[] paras) SqlUpdateOther(string fieldName, object fieldValue, object id, object? id2 = null)
        {
            var (where, paras) = SqlWhere(id, id2);
            return ($"UPDATE {FullName} SET {fieldName} = {fieldValue} {where}", paras);
        }
    }
}

