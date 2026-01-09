using System.Data;
using System.Text;

namespace BotWorker.Infrastructure.Persistence.ORM
{
    public abstract partial class MetaData<TDerived> where TDerived : MetaData<TDerived>, new()
    {
        public static (string, IDataParameter[]) SqlSelect(string columns, object id, object? id2 = null)
        {
            return SqlSelectDict(columns, ToDict(id, id2));
        }

        public static (string, IDataParameter[]) SqlSelectDict(string columns = "*", Dictionary<string, object?>? keys = null)
        {
            keys ??= [];
            var (where, parameters) = SqlWhere(keys, allowEmpty: true);
            return ($"SELECT {columns} FROM {FullName} {where}", parameters);
        }

        public static (string sql, IDataParameter[] parameters) SqlSelect(Dictionary<string, object?> keyValues, string? orderBy = null, int? top = null)
        {   
            var (where, parameters) = SqlWhere(keyValues);
            string topClause = top.HasValue ? SqlTop(top.Value) : "";
            string limitClause = top.HasValue ? SqlLimit(top.Value) : "";
            
            var sql = $"SELECT {topClause}* FROM {FullName} {where}";
            if (!string.IsNullOrWhiteSpace(orderBy)) sql += $" ORDER BY {orderBy}";
            sql += limitClause;

            return (sql, parameters ?? []);
        }

        public static (string, IDataParameter[]) SqlSelectWhere(Dictionary<string, object?> conditions, IEnumerable<string>? selectFields = null,
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
                if (IsPostgreSql)
                {
                    paginationSql = $"LIMIT {limit.GetValueOrDefault(int.MaxValue)} OFFSET {offset.GetValueOrDefault(0)}";
                }
                else
                {
                    // 使用 OFFSET/FETCH 模式（需要 ORDER BY）
                    if (string.IsNullOrEmpty(orderSql))
                        throw new InvalidOperationException("使用分页（OFFSET/FETCH）时必须指定 orderBy 字段。");

                    paginationSql = $"OFFSET {offset.GetValueOrDefault(0)} ROWS";
                    if (limit.HasValue)
                        paginationSql += $" FETCH NEXT {limit.Value} ROWS ONLY";
                }
            }
            else if (limit.HasValue)
            {
                topClause = SqlTop(limit.Value);
                paginationSql = SqlLimit(limit.Value);
            }

            string sql = $"SELECT {topClause}{selectClause} FROM {FullName} {where} {orderSql} {paginationSql}".Trim();

            return (sql, parameters);
        } 

        public static (string sql, IDataParameter[] parameters) SqlSelectWhere(object conditionObj, IEnumerable<string>? selectFields = null, int? limit = null, int? offset = null)
        {
            return SqlSelectWhere(conditionObj.ToDictionary(), selectFields, limit, offset);
        }
    }
}

