using Microsoft.Data.SqlClient;
using BotWorker.Common.Exts;

namespace sz84.Core.MetaDatas
{
    public abstract partial class MetaData<TDerived> where TDerived : MetaData<TDerived>, new()
    {
        public virtual async Task<int> InsertAsync()
        {
            var (sql, paras) = SqlInsertDict(ToDictionary(), [], ["Id", "Guid"]);
            return await ExecAsync(sql, false, paras);
        }

        public virtual async Task<T?> InsertAsync<T>(string field) where T : struct
        {
            var (sql, paras) = SqlInsertDict(ToDictionary(), [field], ["Id", "Guid"]);
            var dict = await ExecWithOutputAsync(sql, paras, [field]);
            return dict.TryGetValue(field, out var val) ? (T?)Convert.ChangeType(val, typeof(T)) : null;
        }

        public virtual async Task<(T1?, T2?)> InsertAsync<T1, T2>(string field1, string field2) where T1 : struct where T2 : struct
        {
            var (sql, paras) = SqlInsertDict(ToDictionary(), [field1, field2], ["Id", "Guid"]);
            var dict = await ExecWithOutputAsync(sql, paras, [field1, field2]);

            var val1 = dict.TryGetValue(field1, out var v1) ? (T1?)Convert.ChangeType(v1, typeof(T1)) : null;
            var val2 = dict.TryGetValue(field2, out var v2) ? (T2?)Convert.ChangeType(v2, typeof(T2)) : null;

            return (val1, val2);
        }

        public static async Task<int> InsertObjectAsync(object obj)
        {
            var (sql, paras) = SqlInsertDict(obj.ToFields());
            return await ExecScalarAsync<int>(sql, paras);
        }

        public static async Task<Dictionary<string, object>> InsertReturnFieldsAsync(object obj, params string[] outputFields)
        {
            var (sql, paras) = SqlInsertDict(obj.ToFields(), outputFields);
            return await ExecWithOutputAsync(sql, paras, outputFields);
        }     

        public static int Insert(List<Cov> columns)
        {
            var data = CovToParams(columns);
            var (sql, paras) = SqlInsertDict(data);
            return Exec(sql, paras);
        }

        public static (string, SqlParameter[]) SqlInsert(List<Cov> columns)
        {
            var data = CovToParams(columns);
            return SqlInsertDict(data);
        }

        public static (string sql, SqlParameter[] parameters) SqlInsertDict(
            Dictionary<string, object?> data,
            string[]? outputFields = null,
            string[]? excludeFields = null)
        {
            if (data == null || data.Count == 0)
                throw new ArgumentException("INSERT 操作必须包含字段和值");

            // ✅ 防止 null 异常
            excludeFields ??= [];
            outputFields ??= [];

            var fields = new List<string>();
            var values = new List<string>();
            var parameters = new List<SqlParameter>();
            int i = 0;

            foreach (var kvp in data)
            {
                if (excludeFields.Contains(kvp.Key, StringComparer.OrdinalIgnoreCase))
                    continue;

                i++;
                string field = kvp.Key;
                object? value = kvp.Value;

                if (value is DateTime dt && dt == DateTime.MinValue)
                {
                    fields.Add(field);
                    values.Add("GETDATE()");
                }
                else
                {
                    string param = $"@p{i}";
                    fields.Add(field);
                    values.Add(param);
                    parameters.Add(new SqlParameter(param, value ?? DBNull.Value));
                }
            }

            string outputClause = "";
            if (outputFields.Length > 0)
            {
                var insertedFields = string.Join(", ", outputFields.Select(f => $"INSERTED.{f}"));
                outputClause = $"OUTPUT {insertedFields} ";
            }
           
            var sql = $"INSERT INTO {FullName} ({string.Join(", ", fields)}) {outputClause}VALUES ({string.Join(", ", values)})";
            return (sql, parameters.ToArray());
        }
    }
}
