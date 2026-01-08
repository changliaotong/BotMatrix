using System.Text;
using Microsoft.Data.SqlClient;
using BotWorker.Infrastructure.Extensions;

namespace BotWorker.Infrastructure.Persistence.ORM
{
    public abstract partial class MetaData<TDerived> where TDerived : MetaData<TDerived>, new()
    {    
        // ----------- 通用构造 Dictionary -----------

        public static Dictionary<string, object?> ToDict(object id, object? id2 = null)
        {
            var dict = new Dictionary<string, object?> { [Key] = id };
            if (!string.IsNullOrEmpty(Key2) && id2 != null)
                dict[Key2!] = id2;
            return dict;
        }

        public static Dictionary<string, object?> ToDict(params (string, object?)[] items)
            => items.ToDictionary(t => t.Item1, t => t.Item2);

        // ----------- 通用构造 SqlParameter[] -----------

        public static SqlParameter[] SqlParams(params (string Name, object? Value)[] pairs)
            => [.. pairs.Select(p => new SqlParameter(p.Name, p.Value ?? DBNull.Value))];

        // ----------- WHERE 构建 -----------

        public static (string whereClause, SqlParameter[] parameters) SqlWhere(Dictionary<string, object?> keys, bool allowEmpty = true)
        {
            if (keys == null || keys.Count == 0)
            {
                if (!allowEmpty)
                    throw new InvalidOperationException("SQL WHERE 条件不能为空。可能导致误操作（如全表删除）");
                return ("", []);
            }

            var sb = new StringBuilder("WHERE ");
            var parameters = new List<SqlParameter>();
            int i = 0;

            foreach (var kvp in keys)
            {
                if (i++ > 0) sb.Append(" AND ");
                var paramName = $"@k{i}";
                sb.Append($"[{kvp.Key}] = {paramName}");
                parameters.Add(new SqlParameter(paramName, kvp.Value ?? DBNull.Value));
            }

            return (sb.ToString(), [.. parameters]);
        }

        protected static (string, SqlParameter[]) SqlWhere(object id, object? id2 = null, bool allowEmpty = false)
            => SqlWhere(ToDict(id, id2), allowEmpty);        

  
        public static Dictionary<string, object?> CovToParams(List<Cov> columns)
        {
            return columns.Select(c => (c.Name, c.Value)).ToDictionary();
        }
    }

}

