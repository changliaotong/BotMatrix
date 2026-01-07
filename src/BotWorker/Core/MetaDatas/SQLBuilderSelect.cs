using System.Text;
using Microsoft.Data.SqlClient;
using BotWorker.Core.Database.Mapping;

namespace BotWorker.Core.MetaDatas
{
    public abstract partial class MetaData<TDerived> where TDerived : MetaData<TDerived>, new()
    {
        public static (string, SqlParameter[]) SqlSelect(string columns, object id, object? id2 = null)
        {
            return SqlSelectDict(columns, ToDict(id, id2));
        }

        public static (string, SqlParameter[]) SqlSelectDict(string columns = "*", Dictionary<string, object?>? keys = null)
        {
            keys ??= [];
            var (where, parameters) = SqlWhere(keys, allowEmpty: true);
            return ($"SELECT {columns} FROM {FullName} {where}", parameters);
        }

        public static (string sql, SqlParameter[] parameters) SqlSelect(Dictionary<string, object?> keyValues, string? orderBy = null, int? top = null)
        {   
            var (where, parameters) = SqlWhere(keyValues);
            var sql = new StringBuilder($"SELECT {(top != null ? $"TOP {top} " : "")}* FROM {FullName} {where}");
            
            if (!string.IsNullOrWhiteSpace(orderBy)) sql.Append($" ORDER BY {orderBy}");
            return (sql.ToString(), parameters ?? []);
        }

        public static (string, SqlParameter[]) SqlSelectWhere(Dictionary<string, object?> conditions, IEnumerable<string>? selectFields = null,
            string? orderBy = null, int? limit = null, int? offset = null)
        {
            var (where, parameters) = SqlWhere(conditions);
            string orderSql = orderBy != null ? $"ORDER BY {orderBy}" : "";

            // 主体字段
            string selectClause = selectFields != null && selectFields.Any()
                ? string.Join(", ", selectFields)
                : "*";

            // 分页逻辑
            string topClause = "";
            string paginationSql = "";

            if (offset.HasValue || limit.HasValue)
            {
                // 使用 OFFSET/FETCH 模式（需要 ORDER BY）
                if (string.IsNullOrEmpty(orderSql))
                    throw new InvalidOperationException("使用分页（OFFSET/FETCH）时必须指定 orderBy 字段。");

                paginationSql = $"OFFSET {offset.GetValueOrDefault(0)} ROWS";
                if (limit.HasValue)
                    paginationSql += $" FETCH NEXT {limit.Value} ROWS ONLY";
            }
            else if (limit.HasValue)
            {
                // 没有 offset，只使用 TOP N
                topClause = $"TOP {limit.Value} ";
            }

            string sql = $"SELECT {topClause}{selectClause} FROM {FullName} {where} {orderSql} {paginationSql}".Trim();

            return (sql, parameters);
        } 

        public static (string sql, SqlParameter[] parameters) SqlSelectWhere(object conditionObj, IEnumerable<string>? selectFields = null, int? limit = null, int? offset = null)
        {
            return SqlSelectWhere(conditionObj.ToDictionary(), selectFields, limit, offset);
        }

        //// 根据字段名和值生成查询SQL
        //public static (string, SqlParameter[]) SqlSelectById(IReadOnlyList<string> selectFields, Dictionary<string, object?> conditions)
        //{
        //    var (where, parameters) = SqlWhere(conditions);
        //    var selectPart = selectFields.Any() ? string.Join(", ", selectFields) : "*";            

        //    string sql = $"SELECT {selectPart} FROM {FullName} {where}";

        //    return (sql, parameters);
        //}

        //public static (string, SqlParameter[]) BuildPagedQuery(string[] selectedFields, Dictionary<string, object?> conditions, string orderBy, int pageIndex, int pageSize)
        //{
        //   var (where, parameters) = SqlWhere(conditions);

        //    int offset = (pageIndex - 1) * pageSize;

        //    string sql = $"SELECT {string.Join(", ", selectedFields)} FROM {FullName} {where} ORDER BY {orderBy} OFFSET {offset} ROWS FETCH NEXT {pageSize} ROWS ONLY;";

        //    return (sql, parameters);
        //}
    }
}
