using System.Data;
using System.Reflection;
using Microsoft.Data.SqlClient;

namespace sz84.Core.MetaDatas
{
    public abstract partial class MetaData<TDerived> where TDerived : MetaData<TDerived>, new()
    {
        public static T GetDef<T>(string fieldName, object id, T def)
            => Get(fieldName, id, null, def);

        public static T GetDef<T>(string fieldName, object id, object? id2, T def)
            => Get(fieldName, id, id2, def);

        public static int GetInt(string fieldName, object id, object? id2 = null)
            => Get<int>(fieldName, id, id2);

        public static short GetShort(string fieldName, object id, object? id2 = null)
            => Get<short>(fieldName, id, id2);

        public static long GetLong(string fieldName, object id, object? id2 = null)
            => Get<long>(fieldName, id, id2);

        public static float GetFloat(string fieldName, object id, object? id2 = null)
            => Get<float>(fieldName, id, id2);

        public static double GetDouble(string fieldName, object id, object? id2 = null)
            => Get<double>(fieldName, id, id2);

        public static byte GetByte(string fieldName, object id, object? id2 = null)
            => Get<byte>(fieldName, id, id2);

        public static decimal GetDecimal(string fieldName, object id, object? id2 = null)
            => Get<decimal>(fieldName, id, id2);

        public static bool GetBool(string fieldName, object id, object? id2 = null)
            => Get<bool>(fieldName, id, id2);

        public static string GetValue(string fieldName, object id, object? id2 = null)
            => Get<string>(fieldName, id, id2) ?? string.Empty;

        //public static DateTime GetDateTime(string fieldName, object id, object? id2 = null)
        //    => Get<DateTime>(fieldName, id, id2);

        public static Guid GetGuid(string fieldName, object id, object? id2 = null)
            => Get<Guid>(fieldName, id, id2);

        public static Dictionary<string, object?>? GetDict(object id, object? id2 = null, params string[] fieldNames)
        {
            return GetDicts(id, id2, fieldNames).FirstOrDefault();
        }

        public static Dictionary<string, object?>? GetDict(object id, params string[] fieldNames)
        {
            return GetDicts(id, fieldNames).FirstOrDefault();
        }

        public static Dictionary<string, object?>? GetDict(string where, SqlParameter[]? parameters, params string[] fieldNames)
        {
            return GetDicts(where, parameters, fieldNames).FirstOrDefault();
        }

        // get bytes
        public static async Task<byte[]?> GetBytes(string fieldName, object id, object? id2 = null)
        {
            var (where, paras) = SqlWhere(id, id2);
            var result = await ExecuteScalarAsync($"{$"SELECT {fieldName} FROM {FullName}"} {where}", paras);
            return result;
        }

        public static T GetValueAandB<T>(string select, string key, object value, string key2, object? value2)
        {
            return QueryScalar<T>($"SELECT {select} FROM {FullName} WHERE {key} = @p1 AND {key2} = @p2", SqlParams(("@p1", value), ("@p2", value2))) ?? default!;
        }

        public static async Task<TDerived?> GetSingleAsync(string columns, object id, object? id2 = null)
        {
            var (sql, parameters) = SqlSelect(columns, id, id2);
            return await QuerySingleAsync<TDerived>($"{sql}", parameters);
        }

        public static async Task<TDerived> GetSingleAsync(object id, object? id2 = null)
        {
            return await GetSingleAsync("*", id, id2) ?? throw new InvalidOperationException($"Failed to load entity by id {id} {id2}");
        }

        public static T? GetSingle<T>(object id, object? id2 = null, params string[] fieldNames) where T : new()
        {
            return GetList<T>(id, id2, fieldNames).FirstOrDefault();
        }

        public static IEnumerable<string> GetRandom(string[] fieldNames, int top = 1)
        {
            string fields = string.Join(", ", fieldNames.Select(fieldName => fieldName.IsNull() ? Key : fieldName));
            string sql = $"SELECT TOP {top} {fields} FROM {FullName} ORDER BY NEWID()";
            DataSet ds = QueryDataset(sql);
            IEnumerable<string> res = ds.Tables[0].AsEnumerable().Select(dr => string.Join(", ", fieldNames.Select(fieldName => dr[fieldName].ToString())));
            return res;
        }

        public static IEnumerable<string> GetRandom(string fieldName, int top = 1)
        {
            return GetRandom([fieldName], top);
        }

        public static string GetRandom(string fieldName)
        {
            return GetRandom(fieldName, 1).FirstOrDefault().AsString();
        }

        private static List<Dictionary<string, object?>> GetDictsInternal(string whereClause, SqlParameter[] parameters, string[] fieldNames)
        {
            if (fieldNames == null || fieldNames.Length == 0)
                throw new ArgumentException("必须指定要查询的字段", nameof(fieldNames));

            var sql = $"SELECT {string.Join(", ", fieldNames)} FROM {FullName} {whereClause}";

            var result = new List<Dictionary<string, object?>>();
            var ds = QueryDataset(sql, parameters);

            if (ds != null && ds.Tables.Count > 0)
            {
                foreach (DataRow row in ds.Tables[0].Rows)
                {
                    var dict = new Dictionary<string, object?>(StringComparer.OrdinalIgnoreCase);
                    foreach (var field in fieldNames)
                    {
                        dict[field] = row[field] is DBNull ? null : row[field];
                    }
                    result.Add(dict);
                }
            }

            return result;
        }

        public static (string, SqlParameter[]) SqlGetStr(string fieldName, object id, object? id2 = null)
        {
            var (where, parameters) = SqlWhere(id, id2);
            return ($"SELECT ISNULL(CONVERT(NVARCHAR(MAX), {fieldName}), '') as res FROM {FullName} {where}", parameters);
        }

        public static (string, SqlParameter[]) SqlGet(string fieldName, object id, object? id2 = null)
        {
            var (where, parameters) = SqlWhere(id, id2);
            string sql = $"SELECT {fieldName} AS res FROM {FullName} {where}";
            return (sql, parameters);
        }

        //public static string GetValue(string fieldName, object id, object? id2 = null)
        //{
        //    var (sql, parameters) = SqlGetStr(fieldName, id, id2);
        //    return QueryScalar<string>(sql, parameters) ?? "";
        //}

        public static object GetObject<T>(string fieldName, object id, object? id2 = null)
        {
            var (sql, parameters) = SqlGet(fieldName, id, id2);
            var result = Query<T>(sql, parameters);
            return result ?? default!;
        }

        public static string SqlGetWhere(string fieldName, string where, string sOrderby = "")
        {
            return $"SELECT TOP 1 {fieldName} FROM {FullName} {where.EnsureStartsWith("WHERE")} {sOrderby.EnsureStartsWith("ORDER BY")}";
        }

        public static string GetWhere(string fieldName, string where, string sOrderby = "")
        {
            var sql = SqlGetWhere(fieldName, where, sOrderby);
            return Query<string>(sql);
        }

        public static T? GetWhere<T>(string fieldName, string where, string sOrderby = "")
        {
            var res = QueryScalar<T>(SqlGetWhere(fieldName, where, sOrderby));
            if (res == null)
            {
                return default;
            }
            return res;
        }

        public static string SqlDate()
        {
            return "SELECT CONVERT(VARCHAR(19), GETDATE(), 120)";
        }

        public static string GetDate(string format = "yyyy-MM-dd HH:mm:ss")
        {
            return Query(SqlDate()).AsDateTimeFormat(format);
        }

        public static DateTime GetDateTime()
        {
            return Convert.ToDateTime(GetDate());
        }

        public static string GetDateTime(string fieldName, object value, object? value2 = null, string format = "yyyy-MM-dd HH:mm:ss")
        {
            return GetValue($"CONVERT(VARCHAR(50), {fieldName}, 120)", value, value2).AsDateTimeFormat(format);
        }

        public static string GetNewId()
        {
            return Query("SELECT NEWID()");
        }

        public static string MaxId()
        {
            return Query($"SELECT MAX({Key}) FROM {FullName}");
        }
    }
}
