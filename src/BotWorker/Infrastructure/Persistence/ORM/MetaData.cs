using System.Data;
using System.Data.Common;
using System.Reflection;
using Newtonsoft.Json;
using BotWorker.Infrastructure.Extensions;

namespace BotWorker.Infrastructure.Persistence.ORM
{
    public static class MetaData
    {
        public static IDataParameter CreateParameter(string parameterName, object? value)
        {
            if (value is byte[] byteValue)
            {
                var param = Persistence.Database.DbProviderFactory.CreateParameter(parameterName, byteValue);
                if (param is DbParameter dbParam) dbParam.DbType = DbType.Binary;
                return param;
            }
            else if (value is bool boolValue)
            {
                return Persistence.Database.DbProviderFactory.CreateParameter(parameterName, boolValue ? 1 : 0);
            }
            else if (value is Enum enumValue)
            {
                return Persistence.Database.DbProviderFactory.CreateParameter(parameterName, Convert.ToInt32(enumValue));
            }
            else if (value is string strValue)
            {
                return Persistence.Database.DbProviderFactory.CreateParameter(parameterName, string.IsNullOrEmpty(strValue) ? "" : strValue);
            }
            else
            {
                return Persistence.Database.DbProviderFactory.CreateParameter(parameterName, value ?? DBNull.Value);
            }
        }
    }

    public abstract partial class MetaData<TDerived> where TDerived : MetaData<TDerived>, new()
    {
        [JsonIgnore]
        public virtual string DataBase { get; } = "sz84_robot";
        [JsonIgnore]
        public abstract string TableName { get; }
        [JsonIgnore]
        public abstract string KeyField { get; }
        [JsonIgnore]
        public virtual string KeyField2 { get; } = string.Empty;
        [JsonIgnore]
        public static string Key { get; set; } = string.Empty;
        [JsonIgnore]
        public static string Key2 { get; set; } = string.Empty;
        [JsonIgnore]
        public static string DbName { get; set; } = string.Empty;
        [JsonIgnore]
        public virtual IReadOnlyList<string> KeyFields =>
            string.IsNullOrEmpty(KeyField2) ? [KeyField] : new[] { KeyField, KeyField2 };

        // 静态缓存子类的主键信息和完整表名
        [JsonIgnore]
        public static readonly string[] Keys;
        [JsonIgnore]
        public static readonly string FullName;

        // 实例方法访问静态缓存字段，方便服务层通过实例获取信息        
        public IReadOnlyList<string> GetKeys() => Keys;        
        public string GetFullName() => FullName;
        private static readonly TDerived _instance = new();

        static MetaData()
        {
            var instance = _instance;
            DbName = instance.DataBase;
            Keys = [instance.KeyField, instance.KeyField2];
            Key = instance.KeyField;
            Key2 = instance.KeyField2;
            FullName = $"[{instance.DataBase}].[dbo].[{instance.TableName}]";
        }

        // 查询：直接静态调用，内部用单例实例处理
        public static Task<List<TDerived>> QueryListAsync(QueryOptions? options = null)
            => _instance.GetListAsync(options);

        public Dictionary<string, object?> ToDictionary()
        {
            var dict = new Dictionary<string, object?>();
            var props = GetType().GetProperties();

            foreach (var prop in props)
            {
                // 1. 跳过索引器属性（带参数的属性，不能直接获取）
                if (prop.GetIndexParameters().Length > 0)
                    continue;

                // 2. 跳过标记了 [DbIgnore] 的属性（自定义显式排除）
                if (prop.GetCustomAttribute<DbIgnoreAttribute>() != null)
                    continue;

                // 3. 跳过只读属性（没有 setter，通常是计算属性，不存数据库）
                if (!prop.CanWrite)
                    continue;

                // 4. 跳过非公共读写属性（一般不存数据库）
                if (!prop.CanRead || !prop.GetMethod!.IsPublic || !prop.SetMethod!.IsPublic)
                    continue;

                // 5. 可选：跳过静态属性（静态字段不属于实例，不存数据库）
                if (prop.GetMethod!.IsStatic)
                    continue;

                // 6. 可选：跳过索引器或特殊属性名，比如以 "_" 或 "$" 开头的（业务需求）
                if (prop.Name.StartsWith("_") || prop.Name.StartsWith("$"))
                    continue;

                // 7. 这里可加你业务特殊判断，比如排除某些字段名等

                var value = prop.GetValue(this);

                dict[prop.Name] = value;
            }

            return dict;
        }

        public static async Task<TDerived> LoadAsync(object id, object? id2 = null)
        {
            return await GetSingleAsync(id, id2) ?? throw new Exception($"主键属性 {id} {id2}不存在");
        }

        public static void SyncCacheField(long qq, long groupId, string field, object value)
        {
            // TODO: 实现缓存同步逻辑（如 Redis 对应字段更新）
            Console.WriteLine($"[CacheSync] {FullName}: {qq}-{groupId} {field} = {value}");
        }

        public static void SyncCacheField(long qq, string field, object value)
        {
            SyncCacheField(qq, 0, field, value);
        }

        public static string QueryRes(string sql, string format)
        {
            return SQLConn.QueryRes(sql, format);
        }

        public static List<T> Query<T>(string sql, params IDataParameter[] parameters)
        {
            return SQLConn.QueryAsync<T>(sql, parameters).GetAwaiter().GetResult();
        }

        public static T? QueryScalar<T>(string sql, params IDataParameter[] parameters)
        {
            return SQLConn.QueryScalar<T>(sql, parameters);
        }

        public static DataSet QueryDataset(string sql, params IDataParameter[] parameters)
        {
            return SQLConn.QueryDataset(sql, parameters);
        }

        public static DataSet QueryDataset(string sql, IDbTransaction? trans, params IDataParameter[] parameters)
        {
            return SQLConn.QueryDataset(sql, trans, parameters);
        }

        public static async Task<T?> QueryScalarAsync<T>(string sql, IDbTransaction? trans = null, params IDataParameter[] parameters)
        {
            return await SQLConn.QueryScalarAsync<T>(sql, true, trans, parameters);
        }

        public static async Task<List<T>> QueryAsync<T>(string sql, IDbTransaction? trans = null, params IDataParameter[] parameters)
        {
            return await SQLConn.QueryAsync<T>(sql, parameters);
        }

        public static async Task<T?> QuerySingleAsync<T>(string sql, IDbTransaction? trans = null, params IDataParameter[] parameters) where T : class, new()
        {
            return await SQLConn.QuerySingleAsync<T>(sql, parameters);
        }

        public static async Task<List<T>> QueryListAsync<T>(string sql, IDbTransaction? trans = null, params IDataParameter[] parameters) where T : new()
        {
            return await SQLConn.QueryListAsync<T>(sql, parameters, trans);
        }

        public static async Task<DataSet> QueryDatasetAsync(string sql, params IDataParameter[] parameters)
        {
            return await SQLConn.QueryDatasetAsync(sql, parameters);
        }

        public static async Task<IDbTransaction> BeginTransactionAsync()
        {
            var conn = Persistence.Database.DbProviderFactory.CreateConnection();
            if (conn is DbConnection dbConn) await dbConn.OpenAsync(); else conn.Open();
            return conn.BeginTransaction();
        }

        public static IDbTransaction BeginTransaction()
        {
            var conn = Persistence.Database.DbProviderFactory.CreateConnection();
            conn.Open();
            return conn.BeginTransaction();
        }

        public static async Task<int> ExecAsync(string sql, params IDataParameter[] parameters)
        {
            return await ExecAsync(sql, null, parameters);
        }

        public static async Task<int> ExecAsync((string sql, IDataParameter[] parameters) sqlInfo, IDbTransaction? trans = null)
        {
            return await ExecAsync(sqlInfo.sql, trans, sqlInfo.parameters);
        }

        public static async Task<int> ExecAsync(string sql, IDbTransaction? trans = null, params IDataParameter[] parameters)
        {
            return await SQLConn.ExecAsync(sql, false, trans, parameters);
        }

        public static async Task<T?> ExecScalarAsync<T>(string sql, params IDataParameter[] parameters) where T : struct
        {
            return await ExecScalarAsync<T>(sql, null, parameters);
        }

        public static async Task<T?> ExecScalarAsync<T>(string sql, IDbTransaction? trans = null, params IDataParameter[] parameters) where T : struct
        {
            return await SQLConn.ExecScalarAsync<T>(sql, false, trans, parameters);
        }

        public static async Task<Dictionary<string, object>> ExecWithOutputAsync(string sql, IDataParameter[] parameters, string[] outputFields, IDbTransaction? trans = null)
        {
            return await SQLConn.ExecWithOutputAsync(sql, parameters, outputFields, trans);
        }

        public static int Exec(string sql, params IDataParameter[] parameters)
        {
            return Exec(sql, null, parameters);
        }

        public static int Exec(string sql, IDbTransaction? trans = null, params IDataParameter[] parameters)
        {
            return SQLConn.Exec(sql, false, trans, parameters);
        }

        public static int Exec((string sql, IDataParameter[] parameters) sqlInfo, IDbTransaction? trans = null)
        {
            return SQLConn.Exec(sqlInfo.sql, false, trans, sqlInfo.parameters);
        }

        // 返回主键列表，保持顺序，方便生成SQL和参数绑定
        public List<(string Name, object Value)> GetKeyValues()
        {
            var list = new List<(string, object)>();
            foreach (var key in Keys)
            {
                var prop = typeof(TDerived).GetProperty(key) ?? throw new Exception($"主键属性 {key} 不存在");
                list.Add((key, prop.GetValue(this) ?? DBNull.Value));
            }
            return list;
        }

        protected virtual Dictionary<string, object> GetInsertFields()
        {
            return PropertyHelper.GetAll(GetType())
                        .Where(p => p.IncludeInInsert)
                        .ToDictionary(p => p.ColumnName, p => p.GetValue(this) ?? DBNull.Value);
        }

        protected virtual Dictionary<string, object> GetUpdateFields()
        {
            return PropertyHelper.GetAll(GetType())
                .Where(p => p.IncludeInUpdate)
                .ToDictionary(p => p.ColumnName, p => p.GetValue(this) ?? DBNull.Value);
        }

        public static string GetSqlValue(object value, string parameterName)
        {
            if (value is DateTime dateTimeValue && dateTimeValue == DateTime.MinValue)
            {
                return $"ISNULL({parameterName}, GETDATE())";
            }
            else if (value is Guid guidValue && guidValue == Guid.Empty)
            {
                return $"ISNULL({parameterName}, NEWID())";
            }
            else
            {
                return parameterName;
            }
        }

        public static IDataParameter CreateParameter(string parameterName, object? value) => MetaData.CreateParameter(parameterName, value);

        public static string FormatValue(object value)
        {
            if (value is null)
            {
                return "NULL";
            }
            else if (value is string)
            {
                return $"N'{EscapeSqlString(value.AsString())}'";
            }
            else if (value is DateTime dateTime)
            {
                if (dateTime == DateTime.MinValue)
                {
                    return "GETDATE()";
                }
                else
                {
                    return $"'{dateTime:yyyy-MM-dd HH:mm:ss}'";
                }
            }
            else
            {
                return value.AsString();
            }
        }

        private static string EscapeSqlString(string value)
        {
            // 在需要的情况下对字符串中的特殊字符进行转义，以防止SQL注入攻击
            // 这里只是简单地对单引号进行替换，更复杂的情况可能需要更多的处理
            return value.Replace("'", "''");
        }
    }
}

