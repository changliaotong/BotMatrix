using System.Data;
using System.Reflection;
using Microsoft.Data.SqlClient;
using BotWorker.Common.Exts;
using sz84.Core.Database;
using sz84.Core.Database.Mapping;

namespace sz84.Core.MetaDatas
{
    public abstract partial class MetaData<TDerived> where TDerived : MetaData<TDerived>, new()
    {
        public virtual async Task<List<TDerived>> GetListAsync(QueryOptions? options = null)
        {
            options ??= new QueryOptions { GetAll = true };

            string where = string.IsNullOrWhiteSpace(options.FilterSql) ? "" : $"WHERE {options.FilterSql}";
            string orderBy = string.IsNullOrWhiteSpace(options.OrderBy) ? $"ORDER BY {KeyField} DESC" : $"ORDER BY {options.OrderBy}";

            string topClause = "";
            string pagingClause = "";

            if (options.GetAll == true)
            {
                // 全部获取，不分页不限制
            }
            else if (options.Top.HasValue)
            {
                topClause = $"TOP {options.Top.Value}";
            }
            else
            {
                // 分页（SQL Server 2012+）
                int page = Math.Max(options.PageIndex ?? 1, 1);
                int size = Math.Max(options.PageSize ?? 20, 1);
                int offset = (page - 1) * size;

                pagingClause = $"OFFSET {offset} ROWS FETCH NEXT {size} ROWS ONLY";
            }

            string sql = $"SELECT {topClause} * FROM {GetFullName()} {where} {orderBy} {pagingClause}";

            return await QueryListAsync<TDerived>(sql, options.Parameters);
        }

        public static List<Dictionary<string, object?>> GetDicts(string whereClause, SqlParameter[]? parameters, params string[] fieldNames)
        {
            return GetDictsInternal(whereClause, parameters ?? [], fieldNames);
        }

        public static List<Dictionary<string, object?>> GetDicts(object id, params string[] fieldNames)
        {
            return GetDicts(id, null, fieldNames);
        }

        public static List<Dictionary<string, object?>> GetDicts(object id, object? id2 = null, params string[] fieldNames)
        {
            var (where, paras) = SqlWhere(id, id2);
            var sql = $"SELECT {string.Join(", ", fieldNames)} FROM {FullName} {where}";

            var result = new List<Dictionary<string, object?>>();
            var ds = QueryDataset(sql, paras);

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

        public static async Task<List<TDerived>> GetListAsync(string sql, params SqlParameter[] parameters) 
        {
            return await QueryListAsync<TDerived>(sql, parameters);
        }            

        public static List<T> GetList<T>(object id, object? id2 = null, params string[] fieldNames) where T : new()
        {
            var (where, paras) = SqlWhere(id, id2);
            var sql = $"SELECT {string.Join(", ", fieldNames)} FROM {FullName} {where}";

            var list = new List<T>();
            var ds = QueryDataset(sql, paras);

            if (ds != null && ds.Tables.Count > 0)
            {
                var table = ds.Tables[0];
                var props = typeof(T).GetProperties(BindingFlags.Public | BindingFlags.Instance);

                foreach (DataRow row in table.Rows)
                {
                    var obj = new T();
                    foreach (var prop in props)
                    {
                        if (fieldNames.Contains(prop.Name, StringComparer.OrdinalIgnoreCase) &&
                            table.Columns.Contains(prop.Name) &&
                            row[prop.Name] is not DBNull)
                        {
                            try
                            {
                                var targetType = Nullable.GetUnderlyingType(prop.PropertyType) ?? prop.PropertyType;
                                var value = Convert.ChangeType(row[prop.Name], targetType);
                                prop.SetValue(obj, value);
                            }
                            catch
                            {
                                // 转换失败就跳过
                            }
                        }
                    }
                    list.Add(obj);
                }
            }
            return list;
        }

        public static async Task<TDerived?> GetByKeysAsync(object id, object? id2 = null)
        {
            var (sql, parameters) = SqlWhere(id, id2);
            var entity = await QuerySingleAsync<TDerived>(sql, parameters);
            return entity;
        }

        public static async Task<T?> GetByKeysAsync<T>(object id, object? id2 = null) where T : new()
        {
            var (sql, parameters) = SqlWhere(id, id2);
            try
            {
                await using var reader = await ExecuteReaderAsync(sql, parameters);
                if (await reader.ReadAsync())
                    return EntityMapper.MapDataReaderToEntity<T>(reader);
                return default;
            }
            catch (Exception ex)
            {
                Error($"查询 {typeof(T).Name} 失败", ex, new { id });
                throw;
            }
        }

        public static string Query(string sql)
        {
            return SQLConn.Query(sql);
        }

        public static DataSet QueryPage(int page, int pageSize, string sWhere)
        {
            return SQLConn.QueryPage(FullName, Key, $"{Key} DESC", page, pageSize, sWhere);
        }

        public static string SqlQueryWhere(string sSelect, string where, string sOrderby)
        {
            return $"SELECT {sSelect} FROM {FullName} {where.EnsureStartsWith("WHERE")} {sOrderby.EnsureStartsWith("ORDER BY")}";
        }

        public static string QueryWhere(string select, string where, string orderby, string format, string countFormat = "")
        {
            return QueryRes(SqlQueryWhere(select, where, orderby), format, countFormat);
        }
    }
}
