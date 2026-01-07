using Microsoft.Data.SqlClient;
using BotWorker.Common.Exts;

namespace sz84.Core.MetaDatas
{
    public abstract partial class MetaData<TDerived> where TDerived : MetaData<TDerived>, new()
    {
        public static (string, SqlParameter[]) SqlExists(object id, object? id2 = null)
        {
            var (where, parameters) = SqlWhere(id, id2);
            return ($"SELECT TOP 1 1 FROM {FullName} {where}", parameters);
        }

        public static (string, SqlParameter[]) SqlExists(params (string, object?)[] keys)
            => SqlExists(ToDict(keys));

        public static (string, SqlParameter[]) SqlExists(string fieldName, object value, string fieldName2, object? value2)
        {
            var keys = new Dictionary<string, object?>
            {
                { fieldName, value ?? DBNull.Value },
                { fieldName2, value2 ?? DBNull.Value}
            };
            return SqlExists(FullName, keys);
        }

        public static bool Exists(object id, object? id2 = null)
        {
            var (sql, parameters) = SqlExists(id, id2);
            var result = QueryScalar<object>(sql, parameters);
            return result != null && result != DBNull.Value;
        }

        public static async Task<bool> ExistsAsync(object id, object? id2 = null)
        {
            var (sql, parameters) = SqlExists(id, id2);
            return await ExecScalarAsync<bool>(sql, parameters);
        }

        public static (string, SqlParameter[]) SqlExistsAandB(string fieldName, object value, string fieldName2, object? value2)
        {
            return ($"SELECT TOP 1 1 FROM {FullName} WHERE {fieldName} = @p1 AND {fieldName2} = @p2", SqlParams(("@p1", value), ("@p2", value2)));
        }

        public static bool ExistsAandB(string fieldName, object value, string fieldName2, object? value2)
        {
            var (sql, parameters) = SqlExistsAandB(fieldName, value, fieldName2, value2);
            var result = QueryScalar<object>(sql, parameters);
            return result != null && result != DBNull.Value;
        }

        public static async Task<bool> ExistsAsync(string fieldName, object value, string fieldName2, object? value2)
        {
            var (sql, parameters) = SqlExists(fieldName, value, fieldName2, value2);
            return await ExecScalarAsync<bool>(sql, parameters);            
        }

        public static bool ExistsField(string fieldName, object value)
        {
            return GetWhere(Key, $"{fieldName} = {value.AsString().Quotes()}").AsBool();
        }

        public static bool ExistsWhere(string sWhere)
        {
            var result = QueryScalar<object>($"SELECT TOP 1 1 FROM {FullName} {sWhere.EnsureStartsWith("WHERE")}");
            return result != null && result != DBNull.Value;
        }

    }
}
