using System.Data;
using System.Reflection;
using BotWorker.Common.Extensions;
using BotWorker.Infrastructure.Persistence.Database;
using BotWorker.Infrastructure.Persistence.ORM;

namespace BotWorker.Infrastructure.Persistence.ORM
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

            return await QueryListAsync<TDerived>(sql, null, options.Parameters);
        }

        public static List<Dictionary<string, object?>> GetDicts(string whereClause, IDataParameter[]? parameters, params string[] fieldNames)
        {
            return GetDictsInternal(whereClause, parameters ?? [], fieldNames);
        }

        public static List<Dictionary<string, object?>> GetDicts(object id, object? id2, IDbTransaction? trans, params string[] fieldNames)
        {
            var (where, paras) = SqlWhere(id, id2);
            var sql = $"SELECT {string.Join(", ", fieldNames)} FROM {FullName} {where}";

            var result = new List<Dictionary<string, object?>>();
            var ds = QueryDataset(sql, trans, paras);

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

        public static List<Dictionary<string, object?>> GetDicts(object id, params string[] fieldNames)
        {
            return GetDicts(id, null, null, fieldNames);
        }

        public static List<Dictionary<string, object?>> GetDicts(object id, object? id2, params string[] fieldNames)
        {
            return GetDicts(id, id2, null, fieldNames);
        }

        public static async Task<List<TDerived>> GetListAsync(string sql, IDbTransaction? trans = null, params IDataParameter[] parameters) 
        {
            return await QueryListAsync<TDerived>(sql, trans, parameters);
        }            

        // 核心 QueryWhere 实现 - 异步返回实体列表
        public static async Task<List<TDerived>> QueryWhere(string where, IDbTransaction? trans = null, params IDataParameter[] parameters)
        {
            return await QueryListAsync<TDerived>($"SELECT * FROM {FullName} WHERE {where}", trans, parameters);
        }

        // 兼容旧的 QueryWhere(where, params parameters)
        public static async Task<List<TDerived>> QueryWhere(string where, params IDataParameter[] parameters)
        {
            return await QueryWhere(where, null, parameters);
        }

        // 返回字符串的 QueryWhere，用于排行榜等场景
        public static string QueryWhere(string select, string where, string orderBy, string format)
        {
            string sql = $"SELECT {select} FROM {FullName} WHERE {where} ORDER BY {orderBy}";
            return QueryRes(sql, format);
        }

        // 返回字符串的 QueryWhere，包含总数格式化
        public static string QueryWhere(string select, string where, string orderBy, string format, string totalFormat)
        {
            string sql = $"SELECT {select} FROM {FullName} WHERE {where} ORDER BY {orderBy}";
            string res = QueryRes(sql, format);
            if (!string.IsNullOrEmpty(res) && !string.IsNullOrEmpty(totalFormat))
            {
                long count = CountWhere(where);
                res += totalFormat.Replace("{c}", count.ToString());
            }
            return res;
        }

        public static async Task<string> QueryResAsync(string sql, string format)
        {
            return await SQLConn.QueryResAsync(sql, format);
        }

        public static async Task<List<TDerived>> QueryAsync(string where, object? param = null)
        {
            if (param is IDataParameter[] paras)
                return await QueryWhere(where.Replace("WHERE ", ""), null, paras);
            
            return await QueryWhere(where.Replace("WHERE ", ""), null, param.ToParameters());
        }

        public static async Task<string> QueryScalarAsync(string sql, params IDataParameter[] parameters)
        {
            return await QueryScalarAsync<string>(sql, null, parameters) ?? "";
        }

        // 核心泛型 QueryWhere 实现 - 返回指定字段的实体列表
        public static List<T> QueryWhere<T>(string where, IDataParameter[]? parameters, IDbTransaction? trans, params string[] fieldNames) where T : new()
        {
            string fields = fieldNames == null || fieldNames.Length == 0 ? "*" : string.Join(", ", fieldNames);
            string sql = $"SELECT {fields} FROM {FullName} WHERE {where}";

            var list = new List<T>();
            var ds = QueryDataset(sql, trans, parameters ?? []);

            if (ds != null && ds.Tables.Count > 0)
            {
                var table = ds.Tables[0];
                var props = typeof(T).GetProperties(BindingFlags.Public | BindingFlags.Instance);

                foreach (DataRow row in table.Rows)
                {
                    var obj = new T();
                    foreach (var prop in props)
                    {
                        if ((fieldNames == null || fieldNames.Length == 0 || fieldNames.Contains(prop.Name, StringComparer.OrdinalIgnoreCase)) &&
                            table.Columns.Contains(prop.Name) &&
                            row[prop.Name] is not DBNull)
                        {
                            try
                            {
                                var targetType = Nullable.GetUnderlyingType(prop.PropertyType) ?? prop.PropertyType;
                                var value = Convert.ChangeType(row[prop.Name], targetType);
                                prop.SetValue(obj, value);
                            }
                            catch { }
                        }
                    }
                    list.Add(obj);
                }
            }
            return list;
        }

        // 兼容 QueryWhere<T>(where, params fieldNames) 
        public static List<T> QueryWhere<T>(string where, params string[] fieldNames) where T : new()
        {
            return QueryWhere<T>(where, null, null, fieldNames);
        }

        // 兼容 QueryWhere<T>(where, parameters, params fieldNames)
        public static List<T> QueryWhere<T>(string where, IDataParameter[]? parameters, params string[] fieldNames) where T : new()
        {
            return QueryWhere<T>(where, parameters, null, fieldNames);
        }

        public static List<T> GetList<T>(object id, object? id2, IDbTransaction? trans, params string[] fieldNames) where T : new()
        {
            var (where, paras) = SqlWhere(id, id2);
            return QueryWhere<T>(where.Replace("WHERE ", ""), paras, trans, fieldNames);
        }

        public static List<T> GetList<T>(object id, object? id2, params string[] fieldNames) where T : new()
        {
            return GetList<T>(id, id2, null, fieldNames);
        }

        public static List<T> GetList<T>(object id, params string[] fieldNames) where T : new()
        {
            return GetList<T>(id, null, null, fieldNames);
        }

        public static async Task<TDerived?> GetByKeysAsync(object id, object? id2 = null)
        {
            var (sql, parameters) = SqlWhere(id, id2);
            var entity = await QuerySingleAsync<TDerived>(sql, null, parameters);
            return entity;
        }

        public static async Task<T?> GetByKeysAsync<T>(object id, object? id2 = null) where T : new()
        {
            var (sql, parameters) = SqlWhere(id, id2);
            try
            {
                var list = await QueryListAsync<T>(sql, null, parameters);
                return list.FirstOrDefault();
            }
            catch
            {
                return default;
            }
        }
    }
}
