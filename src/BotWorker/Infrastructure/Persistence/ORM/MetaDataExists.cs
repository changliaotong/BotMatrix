using Microsoft.Data.SqlClient;
using BotWorker.Infrastructure.Extensions;

namespace BotWorker.Infrastructure.Persistence.ORM
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
            => ExistsAsync(id, id2).GetAwaiter().GetResult();

        public static async Task<bool> ExistsAsync(object id, object? id2 = null)
        {
            var (sql, parameters) = SqlExists(id, id2);
            var result = await QueryScalarAsync<object>(sql, null, parameters);
            return result != null && result != DBNull.Value;
        }

        public static (string, SqlParameter[]) SqlExistsAandB(string fieldName, object value, string fieldName2, object? value2)
        {
            return ($"SELECT TOP 1 1 FROM {FullName} WHERE {fieldName} = @p1 AND {fieldName2} = @p2", SqlParams(("@p1", value), ("@p2", value2)));
        }

        public static bool ExistsAandB(string fieldName, object value, string fieldName2, object? value2)
            => ExistsAsync(fieldName, value, fieldName2, value2).GetAwaiter().GetResult();

        public static async Task<bool> ExistsAsync(string fieldName, object value, string fieldName2, object? value2)
        {
            var (sql, parameters) = SqlExists(fieldName, value, fieldName2, value2);
            var result = await QueryScalarAsync<object>(sql, null, parameters);
            return result != null && result != DBNull.Value;
        }

        public static bool ExistsField(string fieldName, object value)
            => ExistsFieldAsync(fieldName, value).GetAwaiter().GetResult();

        public static async Task<bool> ExistsFieldAsync(string fieldName, object value)
        {
            return (await GetWhereAsync(Key, $"{fieldName} = {value.AsString().Quotes()}")).AsBool();
        }

        public static bool ExistsWhere(string sWhere)
            => ExistsWhereAsync(sWhere).GetAwaiter().GetResult();

        public static async Task<bool> ExistsWhereAsync(string sWhere)
        {
            var result = await QueryScalarAsync<object>($"SELECT TOP 1 1 FROM {FullName} {sWhere.EnsureStartsWith("WHERE")}", null);
            return result != null && result != DBNull.Value;
        }

    }
}

