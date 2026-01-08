using Microsoft.Data.SqlClient;
using BotWorker.Infrastructure.Extensions;
using BotWorker.Infrastructure.Persistence.Database.Mapping;

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
        {
            return CountWhere("");
        }

        public static async Task<long> CountAsync()
        {
            var (sql, parameters) = SqlCount(FullName);
            return await ExecScalarAsync<long>(sql, parameters);
        }

        public static long CountWhere(string where)
        {
            return Query($"SELECT COUNT({Key}) FROM {FullName} {where.EnsureStartsWith("WHERE")}").AsInt();
        }

        public static long CountByKeyValue(string field, string key, string id)
        {
            return Query($"SELECT COUNT({field}) FROM {FullName} WHERE {key} = {id.Quotes()}").AsInt();
        }

        public static long CountField(string fieldName, string KeyField, long FieldValue)
        {
            return CountByKeyValue(fieldName, KeyField, FieldValue.ToString());
        }

        public static long CountKey(string id)
        {
            return CountByKeyValue(Key, Key2, id);
        }

        public static long CountKey2(string id)
        {
            return CountByKeyValue(Key2, Key, id);
        }
    }
}

