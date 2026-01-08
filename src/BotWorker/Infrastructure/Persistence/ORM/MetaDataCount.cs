using Microsoft.Data.SqlClient;

namespace BotWorker.Infrastructure.Persistence.ORM
{
    public abstract partial class MetaData<TDerived> where TDerived : MetaData<TDerived>, new()
    {
        public static (string, SqlParameter[]) SqlCount(string tableFullName, Dictionary<string, object?> keys)
        {
            var (where, parameters) = SqlWhere(keys, allowEmpty: true);
            return ($"SELECT COUNT(*) FROM {tableFullName} {where}", parameters);
        }

        public static (string, SqlParameter[]) SqlCount(string tableFullName, params (string, object?)[] keys)
            => SqlCount(tableFullName, ToDict(keys));

        public static long Count()
            => CountAsync().GetAwaiter().GetResult();

        public static async Task<long> CountAsync()
        {
            var (sql, parameters) = SqlCount(FullName);
            return (await QueryScalarAsync<long>(sql, null, parameters));
        }

        public static long CountWhere(string where)
            => CountWhereAsync(where).GetAwaiter().GetResult();

        public static async Task<long> CountWhereAsync(string where)
        {
            return (await QueryScalarAsync<long>($"SELECT COUNT({Key}) FROM {FullName} {where.EnsureStartsWith("WHERE")}", null)).AsLong();
        }

        public static long CountByKeyValue(string field, string key, string id)
            => CountByKeyValueAsync(field, key, id).GetAwaiter().GetResult();

        public static async Task<long> CountByKeyValueAsync(string field, string key, string id)
        {
            return (await QueryScalarAsync<long>($"SELECT COUNT({field}) FROM {FullName} WHERE {key} = {id.Quotes()}", null)).AsLong();
        }

        public static long CountField(string fieldName, string KeyField, long FieldValue)
            => CountFieldAsync(fieldName, KeyField, FieldValue).GetAwaiter().GetResult();

        public static async Task<long> CountFieldAsync(string fieldName, string KeyField, long FieldValue)
        {
            return await CountByKeyValueAsync(fieldName, KeyField, FieldValue.ToString());
        }

        public static long CountKey(string id)
            => CountKeyAsync(id).GetAwaiter().GetResult();

        public static async Task<long> CountKeyAsync(string id)
        {
            return await CountByKeyValueAsync(Key, Key2, id);
        }

        public static long CountKey2(string id)
            => CountKey2Async(id).GetAwaiter().GetResult();

        public static async Task<long> CountKey2Async(string id)
        {
            return await CountByKeyValueAsync(Key2, Key, id);
        }
    }
}

