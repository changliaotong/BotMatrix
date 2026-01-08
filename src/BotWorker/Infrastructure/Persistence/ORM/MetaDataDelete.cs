using Microsoft.Data.SqlClient;
using BotWorker.Infrastructure.Extensions;

namespace BotWorker.Infrastructure.Persistence.ORM
{
    public abstract partial class MetaData<TDerived> where TDerived : MetaData<TDerived>, new()
    {
        public virtual async Task<int> DeleteAsync(SqlTransaction? trans = null)
        {
            var where = new Dictionary<string, object?>();

            var type = GetType();

            var key1Value = (type.GetProperty(KeyField)?.GetValue(this)) ?? throw new InvalidOperationException($"主键字段 {KeyField} 未赋值");
            where[KeyField] = key1Value;

            if (!string.IsNullOrEmpty(KeyField2))
            {
                var key2Value = (type.GetProperty(KeyField2)?.GetValue(this)) ?? throw new InvalidOperationException($"主键字段 {KeyField2} 未赋值");
                where[KeyField2] = key2Value;
            }

            var (whereSql, parameters) = SqlWhere(where, allowEmpty: false);

            var sql = $"DELETE FROM {GetFullName()} {whereSql}";

            return await ExecAsync(sql, trans, parameters);
        }

        public static (string, SqlParameter[]) SqlDelete(object id, object? id2 = null)
        {
            var dict = ToDict(id, id2);
            return SqlDelete(FullName, dict);
        }

        public static (string, SqlParameter[]) SqlDelete(string tableFullName, Dictionary<string, object?> keys)
        {
            var (where, parameters) = SqlWhere(keys, allowEmpty: false);
            return ($"DELETE FROM {tableFullName} {where}", parameters);
        }

        public static (string, SqlParameter[]) SqlDelete(string tableFullName, params (string, object?)[] keys)
            => SqlDelete(tableFullName, ToDict(keys));


        public static int Delete(object id, object? id2 = null, SqlTransaction? trans = null)
            => DeleteAsync(id, id2, trans).GetAwaiter().GetResult();

        //delete async
        public static async Task<int> DeleteAsync(object id, object? id2 = null, SqlTransaction? trans = null)
        {
            var (sql, paras) = SqlDelete(id, id2);
            return await ExecAsync(sql, trans, paras);
        }

        public static string SqlDeleteAll(object value)
        {
            return $"DELETE FROM {FullName} WHERE {Key} = {value.AsString().Quotes()}";
        }

        public static int DeleteAll(object value, SqlTransaction? trans = null)
            => DeleteAllAsync(value, trans).GetAwaiter().GetResult();

        public static async Task<int> DeleteAllAsync(object value, SqlTransaction? trans = null)
        {
            return await ExecAsync(SqlDeleteAll(value), trans);
        }

        public static string SqlDeleteAll2(object value)
        {
            return $"DELETE FROM {FullName} WHERE {Key2} = {value.AsString().Quotes()}";
        }

        public static int DeleteAll2(object value, SqlTransaction? trans = null)
            => DeleteAll2Async(value, trans).GetAwaiter().GetResult();

        public static async Task<int> DeleteAll2Async(object value, SqlTransaction? trans = null)
        {
            return await ExecAsync(SqlDeleteAll2(value), trans);
        }

        public static int DeleteWhere(string sWhere, SqlTransaction? trans = null)
            => DeleteWhereAsync(sWhere, trans).GetAwaiter().GetResult();

        public static async Task<int> DeleteWhereAsync(string sWhere, SqlTransaction? trans = null)
        {
            return await ExecAsync($"DELETE FROM {FullName} {sWhere.EnsureStartsWith("WHERE")}", trans);
        }
    }
}

