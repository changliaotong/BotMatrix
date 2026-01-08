using System.Text;
using Microsoft.Data.SqlClient;
using BotWorker.Infrastructure.Extensions;


namespace BotWorker.Infrastructure.Persistence.ORM
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

            return await ExecAsync(sql, paras);
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
            return Exec(sql, paras);
        }

        public static int Update(string sqlSet, object id, object? id2 = null)
        {
            var (where, parameters) = SqlWhere(id, id2);
            return Exec($"UPDATE {FullName} SET {sqlSet} {where}", parameters);
        }

        public static async Task<int> UpdateAsync(List<Cov> columns, object id, object? id2 = null)
        {
            var (sql, paras) = SqlUpdate(columns, ToDict(id, id2));
            return await ExecAsync(sql, paras);
        }

        public static async Task<int> UpdateAsync(object obj, object id, object? id2 = null)
        {
            var (sql, paras) = SqlUpdate(obj.ToFields(), ToDict(id, id2));
            return await ExecAsync(sql, paras);
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
            return Exec(sql, parameters);
        }

        public static int SetValues(string set, object id, object? id2 = null)
        {
            var (sql, parameters) = SqlSetValues(set, id, id2);
            return Exec(sql, parameters);
        }

        public static (string, SqlParameter[]) SqlUpdateWhere(string sSet, string sWhere)
        {
            return ($"UPDATE {FullName} SET {sSet} {sWhere.EnsureStartsWith("WHERE")}", []);
        }

        public static int UpdateWhere(string sqlSet, string sqlWhere)
        {
            return Exec(SqlUpdateWhere(sqlSet, sqlWhere));
        }

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
        {
            var (sql, parameters) = SqlPlus(fieldName, plusValue, id, id2);
            return Exec(sql, parameters);
        }

        public static (string, SqlParameter[]) SqlPlus(string fieldName, object plusValue, object id, object? id2 = null)
        {
            var whereDict = ToDict(id, id2);
            var (whereClause, whereParams) = SqlWhere(whereDict, allowEmpty: false);

            string sql = $"UPDATE {FullName} SET {fieldName} = ISNULL({fieldName}, 0) + @plusValue {whereClause}";

            var paramList = whereParams.ToList();
            paramList.Add(new SqlParameter("@plusValue", plusValue ?? 0));

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
    }
}

