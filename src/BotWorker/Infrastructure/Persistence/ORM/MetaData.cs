using System.Data;
using System.Data.Common;
using System.Reflection;
using System.Collections.Concurrent;
using System.Text;
using System.Threading;
using Newtonsoft.Json;
using BotWorker.Infrastructure.Extensions;


namespace BotWorker.Infrastructure.Persistence.ORM
{
    public static class MetaData
    {
        public static ICacheService? CacheService { get; set; } // 由外部初始化注入

        public static bool UseCache { get; set; } = false;

        public static bool IsPostgreSql => GlobalConfig.DbType == DatabaseType.PostgreSql;

        public static string SqlTop(long n) => IsPostgreSql ? "" : $"TOP {n} ";
        public static string SqlLimit(long n) => IsPostgreSql ? $" LIMIT {n}" : "";

        public static string SqlIsNull(string field, string def) => IsPostgreSql ? $"COALESCE({field}, {def})" : $"ISNULL({field}, {def})";

        public static string SqlRandomOrder => IsPostgreSql ? "RANDOM()" : "NEWID()";

        public static string SqlRandomId => IsPostgreSql ? "gen_random_uuid()" : "NEWID()";

        public static string SqlDateTime => IsPostgreSql ? "CURRENT_TIMESTAMP" : "GETDATE()";
        public static string SqlLen(string field) => IsPostgreSql ? $"LENGTH({field})" : $"LEN({field})";

        public static string SqlLockHint => IsPostgreSql ? "" : " WITH (UPDLOCK, ROWLOCK) ";
        public static string SqlForUpdate => IsPostgreSql ? " FOR UPDATE" : "";

        public static string SqlQuote(string identifier)
        {
            if (string.IsNullOrWhiteSpace(identifier)) return identifier;
            if (identifier.Contains('.') || identifier.Contains('"') || identifier.Contains('[') || identifier.Contains(' '))
                return identifier; // 已经引用过或者是复杂表达式，跳过

            return IsPostgreSql ? $"\"{identifier.ToLower()}\"" : $"[{identifier}]";
        }

        /// <summary>
        /// 智能解析 SQL：
        /// 1. 将 [Column] 转换为符合当前数据库方言的引用（SqlServer -> [Column], PgSql -> "Column"）
        /// 2. 支持将 {0}, {1} 占位符转换为 @p1, @p2 参数名
        /// </summary>
        public static (string Sql, IDataParameter[] Parameters) ResolveSql(string sql, params object?[] args)
        {
            return ResolveSql(sql, null, args);
        }

        private static (IDbTransaction? trans, object?[] actualArgs) ParseArgs(object?[] args)
        {
            if (args.Length > 0 && args[0] is IDbTransaction trans)
            {
                return (trans, args.Skip(1).ToArray());
            }
            return (null, args);
        }

        public static (string Sql, IDataParameter[] Parameters) ResolveSql(string sql, IEnumerable<string>? columns, params object?[] args)
        {
            if (string.IsNullOrWhiteSpace(sql)) return (sql, []);

            // 1. 自动为已知列加引用标识 []
            if (columns != null)
            {
                foreach (var col in columns)
                {
                    if (string.IsNullOrEmpty(col)) continue;
                    // 匹配单词边界的列名，且前后没有引用符号 [ ] " "
                    var pattern = @"(?<![\[""])\b" + Regex.Escape(col) + @"\b(?![\]""])";
                    sql = Regex.Replace(sql, pattern, "[" + col + "]");
                }
            }

            // 2. 处理引用符号 [Column] -> "column" (if pgsql)
            if (IsPostgreSql)
            {
                // 仅替换被方括号包裹的内容，并转为小写，以适配全小写迁移方案
                sql = Regex.Replace(sql, @"\[([^\]]+)\]", m => "\"" + m.Groups[1].Value.ToLower() + "\"");
            }

            // 3. 处理参数占位符 {0} -> @p1
            var parameters = new List<IDataParameter>();
            if (args != null && args.Length > 0)
            {
                for (int i = 0; i < args.Length; i++)
                {
                    var placeholder = "{" + i + "}";
                    var paramName = "@p" + (i + 1);
                    if (sql.Contains(placeholder))
                    {
                        sql = sql.Replace(placeholder, paramName);
                        parameters.Add(CreateParameter(paramName, args[i]));
                    }
                }
            }

            return (sql, [.. parameters]);
        }

        public static string SqlDateAdd(string unit, object number, string start)
        {
            if (IsPostgreSql)
            {
                string intervalUnit = unit.ToLower() switch
                {
                    "day" => "day",
                    "hour" => "hour",
                    "minute" => "minute",
                    "second" => "second",
                    "month" => "month",
                    "year" => "year",
                    _ => unit
                };
                return $"(({start})::timestamp + ({number}::text || ' {intervalUnit}')::interval)";
            }
            return $"DATEADD({unit}, {number}, {start})";
        }

        public static string SqlDateDiff(string unit, string start, string end)
        {
            if (IsPostgreSql)
            {
                if (unit.ToLower() == "day")
                    return $"(({end})::date - ({start})::date)";
                if (unit.ToLower() == "hour")
                    return $"(EXTRACT(EPOCH FROM ({end}) - ({start})) / 3600)";
                if (unit.ToLower() == "minute")
                    return $"(EXTRACT(EPOCH FROM ({end}) - ({start})) / 60)";
                if (unit.ToLower() == "second")
                    return $"EXTRACT(EPOCH FROM ({end}) - ({start}))";
                if (unit.ToLower() == "month")
                    return $"(EXTRACT(YEAR FROM age({end}, {start})) * 12 + EXTRACT(MONTH FROM age({end}, {start})))";
            }
            return $"DATEDIFF({unit}, {start}, {end})";
        }

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

        public static async Task<TransactionWrapper> BeginTransactionAsync(IDbTransaction? existingTrans = null, IsolationLevel level = IsolationLevel.ReadCommitted)
        {
            if (existingTrans != null) 
            {
                var trans = existingTrans;
                // 如果已经是包装器，解包以避免多层包装
                if (trans is TransactionWrapper wrapper)
                    trans = wrapper.Transaction;

                return new TransactionWrapper(trans!, false);
            }
            
            var conn = Persistence.Database.DbProviderFactory.CreateConnection();
            if (conn is DbConnection dbConn) await dbConn.OpenAsync(); else conn.Open();
            var newTrans = conn.BeginTransaction(level);
            return new TransactionWrapper(newTrans, true);
        }

        public static TransactionWrapper BeginTransaction(IDbTransaction? existingTrans = null, IsolationLevel level = IsolationLevel.ReadCommitted)
        {
            if (existingTrans != null) 
            {
                var trans = existingTrans;
                if (trans is TransactionWrapper wrapper)
                    trans = wrapper.Transaction;

                return new TransactionWrapper(trans!, false);
            }

            var conn = Persistence.Database.DbProviderFactory.CreateConnection();
            conn.Open();
            var newTrans = conn.BeginTransaction(level);
            return new TransactionWrapper(newTrans, true);
        }

        public static async Task CommitTransactionAsync(IDbTransaction trans)
        {
            if (trans is TransactionWrapper wrapper)
            {
                await wrapper.CommitAsync();
            }
            else
            {
                if (trans is DbTransaction dbTrans)
                    await dbTrans.CommitAsync();
                else
                    trans.Commit();
            }
        }

        public static async Task RollbackTransactionAsync(IDbTransaction trans)
        {
            if (trans is TransactionWrapper wrapper)
            {
                await wrapper.RollbackAsync();
            }
            else
            {
                if (trans is DbTransaction dbTrans)
                    await dbTrans.RollbackAsync();
                else
                    trans.Rollback();
            }
        }

        public static void CommitTransaction(IDbTransaction trans)
        {
            if (trans is TransactionWrapper wrapper)
            {
                wrapper.Commit();
            }
            else
            {
                trans.Commit();
            }
        }

        public static void RollbackTransaction(IDbTransaction trans)
        {
            if (trans is TransactionWrapper wrapper)
            {
                wrapper.Rollback();
            }
            else
            {
                trans.Rollback();
            }
        }

        public static void ClearTransaction(IDbTransaction? trans = null)
        {
            // 在显式传递模式下，不再需要清理全局状态
        }

        /// <summary>
        /// 将匿名对象或字典转换为 Cov 列表，简化 Insert/Update 调用
        /// </summary>
        public static List<Cov> ToCovList(object obj)
        {
            if (obj is List<Cov> list) return list;
            if (obj is Dictionary<string, object?> dict) return dict.Select(kv => new Cov(kv.Key, kv.Value)).ToList();
            
            var result = new List<Cov>();
            var props = obj.GetType().GetProperties();
            foreach (var prop in props)
            {
                if (prop.GetIndexParameters().Length == 0)
                {
                    result.Add(new Cov(prop.Name, prop.GetValue(obj)));
                }
            }
            return result;
        }

        /// <summary>
        /// 事务包装器，确保即使忘记调用 Commit/Rollback，连接也能通过 Dispose 正确释放，并清理 CurrentTransaction
        /// </summary>
        public class TransactionWrapper : IDbTransaction, IAsyncDisposable
        {
            private readonly IDbTransaction _inner;
            private readonly bool _ownConnection;
            private bool _disposed;
            private bool _committed;
            private bool _rolledBack;

            public TransactionWrapper(IDbTransaction inner, bool ownConnection = true)
            {
                _inner = inner;
                _ownConnection = ownConnection;
            }

            public IDbTransaction Transaction => _inner;
            public IDbConnection? Connection => _inner.Connection;
            public IsolationLevel IsolationLevel => _inner.IsolationLevel;

            public static async Task<T> ExecuteAsync<T>(Func<TransactionWrapper, Task<T>> action, IDbTransaction? trans = null, IsolationLevel level = IsolationLevel.ReadCommitted)
            {
                await using var wrapper = await BeginTransactionAsync(trans, level);
                try
                {
                    var result = await action(wrapper);
                    await wrapper.CommitAsync();
                    return result;
                }
                catch
                {
                    await wrapper.RollbackAsync();
                    throw;
                }
            }

            public static async Task ExecuteAsync(Func<TransactionWrapper, Task> action, IDbTransaction? trans = null, IsolationLevel level = IsolationLevel.ReadCommitted)
            {
                await using var wrapper = await BeginTransactionAsync(trans, level);
                try
                {
                    await action(wrapper);
                    await wrapper.CommitAsync();
                }
                catch
                {
                    await wrapper.RollbackAsync();
                    throw;
                }
            }

            public void Commit()
            {
                if (_disposed || _committed || _rolledBack) return;
                if (_ownConnection) _inner.Commit();
                _committed = true;
            }

            public async Task CommitAsync()
            {
                if (_disposed || _committed || _rolledBack) return;
                if (_ownConnection) 
                {
                    if (_inner is DbTransaction dbTrans) await dbTrans.CommitAsync();
                    else _inner.Commit();
                }
                _committed = true;
            }

            public void Rollback()
            {
                if (_disposed || _committed || _rolledBack) return;
                _inner.Rollback();
                _rolledBack = true;
            }

            public async Task RollbackAsync()
            {
                if (_disposed || _committed || _rolledBack) return;
                if (_inner is DbTransaction dbTrans) await dbTrans.RollbackAsync();
                else _inner.Rollback();
                _rolledBack = true;
            }

            public void Dispose()
            {
                if (!_disposed)
                {
                    if (!_committed && !_rolledBack && _ownConnection)
                    {
                        try { Rollback(); } catch { /* Ignore */ }
                    }

                    if (_ownConnection)
                    {
                        var conn = _inner.Connection;
                        _inner.Dispose();
                        conn?.Dispose();
                    }
                    _disposed = true;
                }
            }

            public async ValueTask DisposeAsync()
            {
                if (!_disposed)
                {
                    if (!_committed && !_rolledBack && _ownConnection)
                    {
                        try { await RollbackAsync(); } catch { /* Ignore */ }
                    }

                    if (_ownConnection)
                    {
                        var conn = _inner.Connection;
                        if (_inner is DbTransaction dbTrans)
                            await dbTrans.DisposeAsync();
                        else
                            _inner.Dispose();

                        if (conn is DbConnection dbConn)
                            await dbConn.DisposeAsync();
                        else
                            conn?.Dispose();
                    }
                    _disposed = true;
                }
            }
        }

        public static IDbTransaction? Unwrap(IDbTransaction? trans)
        {
            if (trans is TransactionWrapper wrapper) return wrapper.Transaction;
            return trans;
        }

        /// <summary>
        /// 数据库连接作用域，用于动态切换连接字符串和数据库类型
        /// </summary>
        public class ConnectionScope : IDisposable
        {
            private readonly bool _switched;

            public ConnectionScope(bool usePostgreSql)
            {
                if (usePostgreSql && Persistence.Database.DbProviderFactory.CurrentDbType != DatabaseType.PostgreSql)
                {
                    if (string.IsNullOrEmpty(GlobalConfig.KnowledgeBaseConnection))
                    {
                        // 如果没有配置 PG 连接，则不切换，保持默认
                        return;
                    }
                    Persistence.Database.DbProviderFactory.SetContext(GlobalConfig.KnowledgeBaseConnection, DatabaseType.PostgreSql);
                    _switched = true;
                }
            }

            public void Dispose()
            {
                if (_switched)
                {
                    Persistence.Database.DbProviderFactory.ClearContext();
                }
            }
        }
    }

    /// <summary>
    /// 标记高频更新字段。
    /// 当更新标记了此属性的字段时，ORM 不会自动失效行级缓存，仅失效该字段的列级缓存。
    /// 建议同时配合 [JsonIgnore] 使用，使行级对象不包含该高频变动字段。
    /// </summary>
    [AttributeUsage(AttributeTargets.Property)]
    public class HighFrequencyAttribute : Attribute { }

    public abstract partial class MetaData<TDerived> where TDerived : MetaData<TDerived>, new()
    {
        private static bool _isTableChecked = false;
        private static readonly ConcurrentDictionary<Type, PropertyInfo[]> _propertyCache = new();
        private static readonly bool _usePostgreSql;

        [JsonIgnore]
        public static string Limit1 => IsPostgreSql ? "" : "TOP 1";

        [JsonIgnore]
        public static string SqlDate => IsPostgreSql ? "CURRENT_DATE" : "CONVERT(DATE, GETDATE())";

        [JsonIgnore]
        public static string SqlDateTime => MetaData.SqlDateTime;

        [JsonIgnore]
        public static string SqlYesterday => IsPostgreSql ? "CURRENT_DATE - INTERVAL '1 day'" : "CONVERT(DATE, GETDATE()-1)";

        public static string SqlTop(long n) => MetaData.SqlTop(n);
        public static string SqlLimit(long n) => MetaData.SqlLimit(n);

        public static string SqlIsNull(string field, string def) => MetaData.SqlIsNull(field, def);

        public static string SqlRandomOrder => MetaData.SqlRandomOrder;

        public static string SqlRandomId => MetaData.SqlRandomId;

        public static string SqlLen(string field) => MetaData.SqlLen(field);

        public static string SqlLockHint => MetaData.SqlLockHint;
        public static string SqlForUpdate => MetaData.SqlForUpdate;

        public static string SqlDateAdd(string unit, object number, string start) => MetaData.SqlDateAdd(unit, number, start);

        public static string SqlDateAdd(string unit, object number, DateTime start) => MetaData.SqlDateAdd(unit, number, $"'{start:yyyy-MM-dd HH:mm:ss}'");

        public static string SqlDateAdd(string unit, object number, DateTime? start) => MetaData.SqlDateAdd(unit, number, start?.ToString("yyyy-MM-dd HH:mm:ss") ?? (IsPostgreSql ? "CURRENT_TIMESTAMP" : "GETDATE()"));

        public static string SqlDateDiff(string unit, string start, string end) => MetaData.SqlDateDiff(unit, start, end);

        [JsonIgnore]
        public static bool IsPostgreSql => _usePostgreSql || MetaData.IsPostgreSql;

        public static string Quote(string identifier) => MetaData.SqlQuote(identifier);

        public static (string Sql, IDataParameter[] Parameters) ResolveSql(string sql, params object?[] args)
        {
            var columns = GetProperties().Select(p => p.Name);
            return MetaData.ResolveSql(sql, columns, args);
        }

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

            // 检查是否标记了 [UsePostgreSql] 特性
            _usePostgreSql = typeof(TDerived).GetCustomAttribute<UsePostgreSqlAttribute>() != null;
            
            if (IsPostgreSql)
                FullName = $"\"{instance.DataBase.ToLower()}\".\"public\".\"{instance.TableName.ToLower()}\"";
            else
                FullName = $"[{instance.DataBase}].[dbo].[{instance.TableName}]";
        }

        public static long GetAutoId(string tableName) => SQLConn.GetAutoId(tableName);
        public static async Task<long> GetAutoIdAsync(string tableName) => await SQLConn.GetAutoIdAsync(tableName);

        // ----------- 通用构造 Dictionary & Parameters -----------

        public static Dictionary<string, object?> ToDict(object id, object? id2 = null)
        {
            var dict = new Dictionary<string, object?> { [Key] = id };
            if (!string.IsNullOrEmpty(Key2) && id2 != null)
                dict[Key2!] = id2;
            return dict;
        }

        public static Dictionary<string, object?> ToDict(params (string, object?)[] items)
            => items.ToDictionary(t => t.Item1, t => t.Item2);

        public static IDataParameter[] SqlParams(params (string Name, object? Value)[] pairs)
            => [.. pairs.Select(p => CreateParameter(p.Name, p.Value))];

        public static Dictionary<string, object?> CovToParams(List<Cov> columns)
        {
            return columns.ToDictionary(c => c.Name, c => c.Value);
        }

        // ----------- WHERE 构建 -----------

        public static (string whereClause, IDataParameter[] parameters) SqlWhere(Dictionary<string, object?> keys, bool allowEmpty = true)
        {
            if (keys == null || keys.Count == 0)
            {
                if (!allowEmpty)
                    throw new InvalidOperationException("SQL WHERE 条件不能为空。可能导致误操作（如全表删除）");
                return ("", []);
            }

            var sb = new StringBuilder("WHERE ");
            var parameters = new List<IDataParameter>();
            int i = 0;

            foreach (var kvp in keys)
            {
                if (i++ > 0) sb.Append(" AND ");
                var paramName = $"@k{i}";
                sb.Append($"{Quote(kvp.Key)} = {paramName}");
                parameters.Add(CreateParameter(paramName, kvp.Value));
            }

            return (sb.ToString(), [.. parameters]);
        }

        protected static (string, IDataParameter[]) SqlWhere(object id, object? id2 = null, bool allowEmpty = false)
            => SqlWhere(ToDict(id, id2), allowEmpty);

        // ----------- SQL 执行 & 查询 -----------

        private static (IDbTransaction? trans, object?[] actualArgs, IDataParameter[]? parameters) ParseArgs(object?[] args)
        {
            IDbTransaction? trans = null;
            object?[] actualArgs = args;
            IDataParameter[]? parameters = null;

            int start = 0;
            if (args.Length > 0 && args[0] is IDbTransaction t)
            {
                trans = t;
                start = 1;
            }

            if (args.Length > start && args[^1] is IDataParameter[] p)
            {
                parameters = p;
                actualArgs = args[start..^1];
            }
            else
            {
                actualArgs = args[start..];
            }

            return (trans, actualArgs, parameters);
        }

        public static async Task<string> QueryResAsync(string sql, params object?[] args)
        {
            using var scope = new MetaData.ConnectionScope(_usePostgreSql);
            var (trans, actualArgs, explicitParams) = ParseArgs(args);
            var (resolvedSql, parameters) = ResolveSql(sql, actualArgs);
            if (explicitParams != null) parameters = [.. parameters, .. explicitParams];
            return await QueryResAsync(resolvedSql, "json", trans, parameters);
        }

        public static async Task<string> QueryResAsync(string sql, string format, params object?[] args)
        {
            using var scope = new MetaData.ConnectionScope(_usePostgreSql);
            var (trans, actualArgs, explicitParams) = ParseArgs(args);
            var (resolvedSql, parameters) = ResolveSql(sql, actualArgs);
            if (explicitParams != null) parameters = [.. parameters, .. explicitParams];
            return await QueryResAsync(resolvedSql, format, trans, parameters);
        }

        public static async Task<List<T>> QueryAsync<T>(string sql, params object?[] args)
        {
            using var scope = new MetaData.ConnectionScope(_usePostgreSql);
            var (trans, actualArgs, explicitParams) = ParseArgs(args);
            var (resolvedSql, parameters) = ResolveSql(sql, actualArgs);
            if (explicitParams != null) parameters = [.. parameters, .. explicitParams];
            return await QueryAsync<T>(resolvedSql, trans, parameters);
        }

        public static async Task<T?> QueryScalarAsync<T>(string sql, params object?[] args)
        {
            using var scope = new MetaData.ConnectionScope(_usePostgreSql);
            var (trans, actualArgs, explicitParams) = ParseArgs(args);
            var (resolvedSql, parameters) = ResolveSql(sql, actualArgs);
            if (explicitParams != null) parameters = [.. parameters, .. explicitParams];
            return await QueryScalarAsync<T>(resolvedSql, trans, parameters);
        }

        public static async Task<T?> QuerySingleAsync<T>(string sql, params object?[] args) where T : class, new()
        {
            using var scope = new MetaData.ConnectionScope(_usePostgreSql);
            var (trans, actualArgs, explicitParams) = ParseArgs(args);
            var (resolvedSql, parameters) = ResolveSql(sql, actualArgs);
            if (explicitParams != null) parameters = [.. parameters, .. explicitParams];
            return await QuerySingleAsync<T>(resolvedSql, trans, parameters);
        }

        public static async Task<List<T>> QueryListAsync<T>(string sql, params object?[] args) where T : new()
        {
            using var scope = new MetaData.ConnectionScope(_usePostgreSql);
            var (trans, actualArgs, explicitParams) = ParseArgs(args);
            var (resolvedSql, parameters) = ResolveSql(sql, actualArgs);
            if (explicitParams != null) parameters = [.. parameters, .. explicitParams];
            return await QueryListAsync<T>(resolvedSql, trans, parameters);
        }

        public static async Task<int> ExecAsync(string sql, params object?[] args)
        {
            using var scope = new MetaData.ConnectionScope(_usePostgreSql);
            var (trans, actualArgs, explicitParams) = ParseArgs(args);
            var (resolvedSql, parameters) = ResolveSql(sql, actualArgs);
            if (explicitParams != null) parameters = [.. parameters, .. explicitParams];
            return await ExecAsync(resolvedSql, trans, parameters);
        }

        public static async Task<T?> ExecScalarAsync<T>(string sql, params object?[] args)
        {
            using var scope = new MetaData.ConnectionScope(_usePostgreSql);
            var (trans, actualArgs, explicitParams) = ParseArgs(args);
            var (resolvedSql, parameters) = ResolveSql(sql, actualArgs);
            if (explicitParams != null) parameters = [.. parameters, .. explicitParams];
            return await ExecScalarAsync<T>(resolvedSql, trans, parameters);
        }

        // --- 明确的 Trans 接口 ---

        public static async Task<string> QueryResTransAsync(IDbTransaction trans, string sql, params object?[] args)
        {
            var (resolvedSql, parameters) = ResolveSql(sql, args);
            return await QueryResAsync(resolvedSql, "json", trans, parameters);
        }

        public static async Task<List<T>> QueryTransAsync<T>(IDbTransaction trans, string sql, params object?[] args)
        {
            var (resolvedSql, parameters) = ResolveSql(sql, args);
            return await QueryAsync<T>(resolvedSql, trans, parameters);
        }

        public static async Task<T?> QueryScalarTransAsync<T>(IDbTransaction trans, string sql, params object?[] args)
        {
            var (resolvedSql, parameters) = ResolveSql(sql, args);
            return await QueryScalarAsync<T>(resolvedSql, trans, parameters);
        }

        public static async Task<int> ExecTransAsync(IDbTransaction trans, string sql, params object?[] args)
        {
            var (resolvedSql, parameters) = ResolveSql(sql, args);
            return await ExecAsync(resolvedSql, trans, parameters);
        }

        public static string QueryRes(string sql, string format, IDbTransaction? trans = null)
        {
            return SQLConn.QueryRes(sql, format, "", trans);
        }

        public static async Task<string> QueryResAsync(string sql, string format, IDbTransaction? trans, IDataParameter[] parameters)
        {
            var (resolvedSql, _) = ResolveSql(sql);
            return await SQLConn.QueryResAsync(resolvedSql, format, "", trans, parameters);
        }

        public static async Task<string> QueryResAsync(string sql, string format, IDataParameter[] parameters)
            => await QueryResAsync(sql, format, null, parameters);

        public static async Task<string> QueryResAsync(string sql, string format, string countFormat, IDbTransaction? trans, IDataParameter[] parameters)
        {
            var (resolvedSql, _) = ResolveSql(sql);
            return await SQLConn.QueryResAsync(resolvedSql, format, countFormat, trans, parameters);
        }

        public static List<T> Query<T>(string sql, IDbTransaction? trans = null, IDataParameter[]? parameters = null)
        {
            return SQLConn.QueryList<T>(sql, trans, parameters ?? []) ?? [];
        }

        public static T? QuerySingle<T>(string sql, IDbTransaction? trans = null, IDataParameter[]? parameters = null) where T : class, new()
        {
            return Query<T>(sql, trans, parameters).FirstOrDefault();
        }

        public static T? QueryScalar<T>(string sql, IDbTransaction? trans = null, IDataParameter[]? parameters = null)
        {
            return SQLConn.QueryScalar<T>(sql, trans, parameters ?? []);
        }

        public static DataSet QueryDataset(string sql, IDataParameter[]? parameters = null)
        {
            return SQLConn.QueryDataset(sql, null, parameters ?? []);
        }

        public static DataSet QueryDataset(string sql, IDbTransaction? trans, IDataParameter[]? parameters = null)
        {
            return SQLConn.QueryDataset(sql, trans, parameters ?? []);
        }

        public static async Task<T?> QueryScalarAsync<T>(string sql, IDbTransaction? trans, IDataParameter[] parameters)
        {
            var (resolvedSql, _) = ResolveSql(sql);
            return await SQLConn.QueryScalarAsync<T>(resolvedSql, true, trans, parameters);
        }

        public static async Task<T?> QueryScalarAsync<T>(string sql, IDataParameter[] parameters)
            => await QueryScalarAsync<T>(sql, null, parameters);

        public static async Task<List<T>> QueryAsync<T>(string sql, IDbTransaction? trans, IDataParameter[] parameters)
        {
            var (resolvedSql, _) = ResolveSql(sql);
            return await SQLConn.QueryAsync<T>(resolvedSql, parameters, trans);
        }

        public static async Task<List<T>> QueryAsync<T>(string sql, IDataParameter[] parameters)
            => await QueryAsync<T>(sql, null, parameters);

        public static async Task<T?> QuerySingleAsync<T>(string sql, IDbTransaction? trans, IDataParameter[] parameters) where T : class, new()
        {
            var (resolvedSql, _) = ResolveSql(sql);
            return await SQLConn.QuerySingleAsync<T>(resolvedSql, trans, parameters);
        }

        public static async Task<List<T>> QueryListAsync<T>(string sql, IDbTransaction? trans, IDataParameter[] parameters) where T : new()
        {
            var (resolvedSql, _) = ResolveSql(sql);
            return await SQLConn.QueryListAsync<T>(resolvedSql, parameters, trans);
        }

        public static async Task<List<T>> QueryListAsync<T>(string sql, IDataParameter[] parameters) where T : new()
            => await QueryListAsync<T>(sql, null, parameters);

        public static async Task<DataSet> QueryDatasetAsync(string sql, IDataParameter[]? parameters = null)
        {
            return await QueryDatasetAsync(sql, null, parameters ?? []);
        }

        public static async Task<DataSet> QueryDatasetAsync(string sql, IDbTransaction? trans, IDataParameter[]? parameters = null)
        {
            var (resolvedSql, _) = ResolveSql(sql);
            return await SQLConn.QueryDatasetAsync(resolvedSql, trans, parameters ?? []);
        }

        public static async Task<int> ExecAsync(string sql, IDataParameter[] parameters)
        {
            return await ExecAsync(sql, null, parameters);
        }

        public static async Task<int> ExecAsync((string sql, IDataParameter[] parameters) sqlInfo, IDbTransaction? trans = null)
        {
            return await ExecAsync(sqlInfo.sql, trans, sqlInfo.parameters);
        }

        public static async Task<int> ExecAsync(string sql, IDbTransaction? trans, IDataParameter[] parameters)
        {
            var (resolvedSql, _) = ResolveSql(sql);
            // Logger.Debug($"[DB] Exec - SQL: {resolvedSql}");
            return await SQLConn.ExecAsync(resolvedSql, false, trans, parameters);
        }

        public static async Task<T?> ExecScalarAsync<T>(string sql, IDbTransaction? trans, IDataParameter[] parameters)
        {
            var (resolvedSql, _) = ResolveSql(sql);
            // Logger.Debug($"[DB] ExecScalar - SQL: {resolvedSql}");
            return await SQLConn.ExecScalarAsync<T>(resolvedSql, false, trans, parameters);
        }

        public static async Task<T?> ExecScalarAsync<T>(string sql, IDataParameter[] parameters)
            => await ExecScalarAsync<T>(sql, null, parameters);

        public static async Task<Dictionary<string, object>> ExecWithOutputAsync(string sql, IDataParameter[] parameters, string[] outputFields, IDbTransaction? trans = null)
        {
            var (resolvedSql, _) = ResolveSql(sql);
            return await SQLConn.ExecWithOutputAsync(resolvedSql, parameters, outputFields, trans);
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

        // ----------- Select & List 查询 -----------

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
                ? string.Join(", ", selectFields.Select(f => Quote(f)))
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

        public static Task<List<TDerived>> QueryListAsync(QueryOptions? options = null)
            => _instance.GetListAsync(options);

        public virtual async Task<List<TDerived>> GetListAsync(QueryOptions? options = null)
        {
            options ??= new QueryOptions { GetAll = true };

            string where = string.IsNullOrWhiteSpace(options.FilterSql) ? "" : $"WHERE {options.FilterSql}";
            string orderBy = string.IsNullOrWhiteSpace(options.OrderBy) ? $"ORDER BY {Quote(Key)} DESC" : $"ORDER BY {options.OrderBy}";

            string topClause = "";
            string pagingClause = "";

            if (options.GetAll == true)
            {
                // 全部获取，不分页不限制
            }
            else if (options.Top.HasValue)
            {
                topClause = SqlTop(options.Top.Value);
                pagingClause = SqlLimit(options.Top.Value);
            }
            else
            {
                // 分页
                int page = Math.Max(options.PageIndex ?? 1, 1);
                int size = Math.Max(options.PageSize ?? 20, 1);
                int offset = (page - 1) * size;

                if (IsPostgreSql)
                    pagingClause = $"LIMIT {size} OFFSET {offset}";
                else
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
            var sql = $"SELECT {string.Join(", ", fieldNames.Select(f => MetaData.SqlQuote(f)))} FROM {FullName} {where}";

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

        public static async Task<List<TDerived>> QueryWhere(string where, params object?[] args)
        {
            var (resolvedSql, parameters) = ResolveSql($"SELECT * FROM {FullName} {where.EnsureStartsWith("WHERE ")}", args);
            return await QueryListAsync<TDerived>(resolvedSql, null, parameters);
        }

        public static async Task<List<TDerived>> QueryWhere(string where, IDbTransaction? trans, IDataParameter[] parameters)
        {
            var sql = $"SELECT * FROM {FullName} {where.EnsureStartsWith("WHERE ")}";
            return await QueryListAsync<TDerived>(sql, trans, parameters);
        }

        public static async Task<List<TDerived>> QueryWhere(string where, IDataParameter[] parameters)
            => await QueryWhere(where, null, parameters);

        public static string QueryWhere(string select, string where, string orderBy, string format)
        {
            string sql = $"SELECT {select} FROM {FullName} WHERE {where} ORDER BY {orderBy}";
            return QueryRes(sql, format);
        }

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

        public static async Task<string> QueryWhereAsync(string select, string where, string orderBy, string format)
        {
            string sql = $"SELECT {select} FROM {FullName} WHERE {where} ORDER BY {orderBy}";
            return await QueryResAsync(sql, format);
        }

        public static async Task<string> QueryWhereAsync(string select, string where, string orderBy, string format, string totalFormat)
        {
            string sql = $"SELECT {select} FROM {FullName} WHERE {where} ORDER BY {orderBy}";
            string res = await QueryResAsync(sql, format);
            if (!string.IsNullOrEmpty(res) && !string.IsNullOrEmpty(totalFormat))
            {
                long count = await CountWhereAsync(where);
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

        public static async Task<string> QueryScalarAsync(string sql, IDbTransaction? trans = null, params IDataParameter[] parameters)
        {
            return await QueryScalarAsync<string>(sql, trans, parameters) ?? "";
        }

        public static List<T> QueryWhere<T>(string where, IDataParameter[]? parameters, IDbTransaction? trans, params string[] fieldNames) where T : new()
        {
            string fields = fieldNames == null || fieldNames.Length == 0 ? "*" : string.Join(", ", fieldNames.Select(f => MetaData.SqlQuote(f)));
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

        public static List<T> QueryWhere<T>(string where, params string[] fieldNames) where T : new()
        {
            return QueryWhere<T>(where, null, null, fieldNames);
        }

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

        // ----------- Get & Scalar 数据获取 -----------

        public static T Get<T>(string fieldName, object id, object? id2 = null)
        {
            if (typeof(T) == typeof(int)) return (T)(object)GetInt(fieldName, id, id2);
            if (typeof(T) == typeof(long)) return (T)(object)GetLong(fieldName, id, id2);
            if (typeof(T) == typeof(bool)) return (T)(object)GetBool(fieldName, id, id2);
            if (typeof(T) == typeof(string)) return (T)(object)GetValue(fieldName, id, id2);
            
            return GetFromDb<T>(fieldName, id, id2);
        }

        public static T GetFromDb<T>(string fieldName, object id, object? id2 = null, IDbTransaction? trans = null)
        {
            var (sql, parameters) = SqlGet(fieldName, id, id2);
            var res = QueryScalar<T>(sql, trans, parameters);
            return res ?? default!;
        }

        public static async Task<T> GetFromDbAsync<T>(string fieldName, object id, object? id2 = null, IDbTransaction? trans = null)
        {
            var (sql, parameters) = SqlGet(fieldName, id, id2);
            var res = await QueryScalarAsync<T>(sql, trans, parameters);
            return res ?? default!;
        }

        public static T GetDef<T>(string fieldName, object id, T def)
            => GetDef(fieldName, id, null, def);

        public static async Task<T> GetDefAsync<T>(string fieldName, object id, T def, IDbTransaction? trans = null)
            => await GetDefAsync(fieldName, id, null, def, trans);

        public static T GetDef<T>(string fieldName, object id, object? id2, T def)
        {
            var res = GetFromDb<T>(fieldName, id, id2);
            return res == null || res.Equals(default(T)) ? def : res;
        }

        public static async Task<T> GetDefAsync<T>(string fieldName, object id, object? id2, T def, IDbTransaction? trans = null)
        {
            var res = await GetFromDbAsync<T>(fieldName, id, id2, trans);
            return res == null || res.Equals(default(T)) ? def : res;
        }

        public static int GetInt(string fieldName, object id, object? id2 = null)
            => Get<int>(fieldName, id, id2, 0);

        public static async Task<int> GetIntAsync(string fieldName, object id, object? id2 = null, IDbTransaction? trans = null)
            => await GetFieldAsync<int>(fieldName, id, id2, 0, trans);

        public static long GetLong(string fieldName, object id, object? id2 = null)
            => Get<long>(fieldName, id, id2, 0L);

        public static async Task<long> GetLongAsync(string fieldName, object id, object? id2 = null, IDbTransaction? trans = null)
            => await GetFieldAsync<long>(fieldName, id, id2, 0L, trans);

        public static bool GetBool(string fieldName, object id, object? id2 = null)
        {
            return Get<bool>(fieldName, id, id2, false);
        }

        public static async Task<bool> GetBoolAsync(string fieldName, object id, object? id2 = null, IDbTransaction? trans = null)
        {
            return await GetFieldAsync<bool>(fieldName, id, id2, false, trans);
        }

        public static T Get<T>(string fieldName, object id, object? id2, T def, IDbTransaction? trans = null)
        {
            var res = GetFromDb<T>(fieldName, id, id2, trans);
            return res == null || res.Equals(default(T)) ? def : res;
        }

        public static async Task<T> GetAsync<T>(string fieldName, object id, object? id2, T def, IDbTransaction? trans = null)
        {
            var res = await GetFromDbAsync<T>(fieldName, id, id2, trans);
            return res == null || res.Equals(default(T)) ? def : res;
        }

        public static Dictionary<string, object?>? GetDict(object id, object? id2 = null, params string[] fieldNames)
        {
            var (where, parameters) = SqlWhere(id, id2);
            return GetDict(where, parameters, fieldNames);
        }

        public static async Task<Dictionary<string, object?>?> GetDictAsync(object id, object? id2 = null, params string[] fieldNames)
        {
            var (where, parameters) = SqlWhere(id, id2);
            return await GetDictAsync(where, parameters, fieldNames);
        }

        public static Dictionary<string, object?>? GetDict(object id, params string[] fieldNames)
        {
            return GetDict(id, null, fieldNames);
        }

        public static async Task<Dictionary<string, object?>?> GetDictAsync(object id, params string[] fieldNames)
        {
            return await GetDictAsync(id, null, fieldNames);
        }

        public static Dictionary<string, object?>? GetDict(string where, IDataParameter[]? parameters, params string[] fieldNames)
        {
            if (fieldNames == null || fieldNames.Length == 0)
                throw new ArgumentException("必须指定要查询的字段", nameof(fieldNames));

            var sql = $"SELECT {string.Join(", ", fieldNames.Select(f => MetaData.SqlQuote(f)))} FROM {FullName} {where}";
            var results = SQLConn.QueryList<dynamic>(sql, parameters ?? Array.Empty<IDataParameter>());
            var row = results?.FirstOrDefault();
            if (row == null) return null;

            var dict = new Dictionary<string, object?>(StringComparer.OrdinalIgnoreCase);
            var rowDict = (IDictionary<string, object>)row;
            foreach (var field in fieldNames)
            {
                dict[field] = rowDict.ContainsKey(field) ? rowDict[field] : null;
            }
            return dict;
        }

        public static async Task<Dictionary<string, object?>?> GetDictAsync(string where, IDataParameter[]? parameters, params string[] fieldNames)
        {
            if (fieldNames == null || fieldNames.Length == 0)
                throw new ArgumentException("必须指定要查询的字段", nameof(fieldNames));

            var sql = $"SELECT {string.Join(", ", fieldNames.Select(f => MetaData.SqlQuote(f)))} FROM {FullName} {where}";
            var results = await QueryAsync<dynamic>(sql, null, parameters ?? Array.Empty<IDataParameter>());
            var row = results.FirstOrDefault();
            if (row == null) return null;

            var dict = new Dictionary<string, object?>(StringComparer.OrdinalIgnoreCase);
            var rowDict = (IDictionary<string, object>)row;
            foreach (var field in fieldNames)
            {
                dict[field] = rowDict.ContainsKey(field) ? rowDict[field] : null;
            }
            return dict;
        }

        public static string GetValue(string fieldName, object id, object? id2 = null)
        {
            var (sql, parameters) = SqlGet(fieldName, id, id2);
            return QueryScalar<string>(sql, null, parameters) ?? "";
        }

        public static Guid GetGuid(string fieldName, object id, object? id2 = null)
            => Get<Guid>(fieldName, id, id2);

        public static async Task<byte[]?> GetBytes(string fieldName, object id, object? id2 = null, IDbTransaction? trans = null)
        {
            var (where, paras) = SqlWhere(id, id2);
            var result = await ExecScalarAsync<byte[]>($"SELECT {Quote(fieldName)} FROM {FullName} {where}", trans, paras);
            return result;
        }

        public static T GetValueAandB<T>(string select, string key, object value, string key2, object? value2, IDbTransaction? trans = null)
        {
            return QueryScalar<T>($"SELECT {Quote(select)} FROM {FullName} WHERE {Quote(key)} = @p1 AND {Quote(key2)} = @p2", trans, SqlParams(("@p1", value), ("@p2", value2))) ?? default!;
        }

        public static async Task<TDerived?> GetSingleAsync(string columns, object id, object? id2 = null, IDbTransaction? trans = null)
        {
            var (sql, parameters) = SqlSelect(columns, id, id2);
            return await QuerySingleAsync<TDerived>($"{sql}", trans, parameters);
        }

        public static async Task<TDerived?> GetSingleAsync(object id, object? id2 = null, IDbTransaction? trans = null)
        {
            return await GetSingleAsync("*", id, id2, trans);
        }

        public static (string, IDataParameter[]) SqlSelectForUpdate(string columns, object id, object? id2 = null)
        {
            var (where, parameters) = SqlWhere(id, id2);
            string sql = IsPostgreSql
                ? $"SELECT {columns} FROM {FullName} {where} {SqlForUpdate}"
                : $"SELECT {columns} FROM {FullName} {SqlLockHint} {where}";
            return (sql, parameters);
        }

        public static async Task<TDerived?> GetSingleForUpdateAsync(object id, object? id2 = null, IDbTransaction? trans = null)
        {
            var (sql, parameters) = SqlSelectForUpdate("*", id, id2);
            return await QuerySingleAsync<TDerived>(sql, trans, parameters);
        }

        public static async Task<TDerived> LoadForUpdateAsync(object id, object? id2 = null, IDbTransaction? trans = null)
            => await GetSingleForUpdateAsync(id, id2, trans) ?? throw new Exception($"{FullName} 主键 {id} {id2} 不存在");

        public static T? GetSingle<T>(object id, object? id2 = null, params string[] fieldNames) where T : new()
        {
            return GetList<T>(id, id2, fieldNames).FirstOrDefault();
        }

        public static async Task<IEnumerable<string>> GetRandomAsync(string[] fieldNames, int top = 1)
        {
            string fields = string.Join(", ", fieldNames.Select(fieldName => fieldName.IsNull() ? Quote(Key) : MetaData.SqlQuote(fieldName)));
            string sql = $"SELECT {SqlTop(top)}{fields} FROM {FullName} ORDER BY {SqlRandomOrder}{SqlLimit(top)}";
            DataSet ds = await QueryDatasetAsync(sql);
            if (ds == null || ds.Tables.Count == 0) return [];
            IEnumerable<string> res = ds.Tables[0].AsEnumerable().Select(dr => string.Join(", ", fieldNames.Select(fieldName => dr[fieldName].ToString())));
            return res;
        }

        public static async Task<IEnumerable<string>> GetRandomAsync(string fieldName, int top = 1)
        {
            return await GetRandomAsync([fieldName], top);
        }

        public static async Task<string> GetRandomAsync(string fieldName)
        {
            return (await GetRandomAsync(fieldName, 1)).FirstOrDefault().AsString();
        }

        public static IEnumerable<string> GetRandom(string[] fieldNames, int top = 1)
        {
            string fields = string.Join(", ", fieldNames.Select(fieldName => fieldName.IsNull() ? Quote(Key) : MetaData.SqlQuote(fieldName)));
            string sql = $"SELECT {SqlTop(top)}{fields} FROM {FullName} ORDER BY {SqlRandomOrder}{SqlLimit(top)}";
            DataSet ds = QueryDataset(sql);
            IEnumerable<string> res = ds.Tables[0].AsEnumerable().Select(dr => string.Join(", ", fieldNames.Select(fieldName => dr[fieldName].ToString())));
            return res;
        }

        public static IEnumerable<string> GetRandom(string fieldName, int top = 1)
        {
            return GetRandom([fieldName], top);
        }

        public static string GetRandom(string fieldName)
        {
            return GetRandom(fieldName, 1).FirstOrDefault().AsString();
        }

        private static List<Dictionary<string, object?>> GetDictsInternal(string whereClause, IDataParameter[] parameters, string[] fieldNames)
        {
            if (fieldNames == null || fieldNames.Length == 0)
                throw new ArgumentException("必须指定要查询的字段", nameof(fieldNames));

            var sql = $"SELECT {string.Join(", ", fieldNames.Select(f => MetaData.SqlQuote(f)))} FROM {FullName} {whereClause}";

            var result = new List<Dictionary<string, object?>>();
            var ds = QueryDataset(sql, null, parameters);

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

        public static (string, IDataParameter[]) SqlGetStr(string fieldName, object id, object? id2 = null)
        {
            var (where, parameters) = SqlWhere(id, id2);
            string quotedField = MetaData.SqlQuote(fieldName);
            string convertField = IsPostgreSql ? $"{quotedField}::text" : $"CONVERT(NVARCHAR(MAX), {quotedField})";
            string sql = $"SELECT {SqlIsNull(convertField, "''")} as res FROM {FullName} {where}";
            return (sql, parameters);
        }

        public static (string, IDataParameter[]) SqlGet(string fieldName, object id, object? id2 = null)
        {
            var (where, parameters) = SqlWhere(id, id2);
            string sql = $"SELECT {MetaData.SqlQuote(fieldName)} AS res FROM {FullName} {where}";
            return (sql, parameters);
        }

        public static object? GetObject<T>(string fieldName, object id, object? id2 = null)
        {
            var (sql, parameters) = SqlGet(fieldName, id, id2);
            var result = Query<T>(sql, null, parameters);
            return result.FirstOrDefault();
        }

        public static string SqlGetWhere(string fieldName, string where, string sOrderby = "")
        {
            return $"SELECT {SqlTop(1)}{Quote(fieldName)} FROM {FullName} {where.EnsureStartsWith("WHERE")} {sOrderby.EnsureStartsWith("ORDER BY")} {SqlLimit(1)}";
        }

        public static string GetWhere(string fieldName, string where, string sOrderby = "")
        {
            var sql = SqlGetWhere(fieldName, where, sOrderby);
            return Query<string>(sql).FirstOrDefault() ?? "";
        }

        public static async Task<string> GetWhereAsync(string fieldName, string where, string sOrderby = "", IDbTransaction? trans = null)
        {
            var sql = SqlGetWhere(fieldName, where, sOrderby);
            return (await QueryAsync<string>(sql, trans)).FirstOrDefault() ?? "";
        }

        public static async Task<T?> GetWhereAsync<T>(string fieldName, string where, string sOrderby = "", IDbTransaction? trans = null)
        {
            var res = await QueryScalarAsync<T>(SqlGetWhere(fieldName, where, sOrderby), trans);
            return res;
        }

        public static T? GetWhere<T>(string fieldName, string where, IDataParameter[] parameters)
        {
            var sql = $"SELECT {MetaData.SqlQuote(fieldName)} FROM {FullName} {where.EnsureStartsWith("WHERE ")}";
            return QueryScalar<T>(sql, null, parameters);
        }

        public static T? GetWhere<T>(string fieldName, string where, string sOrderby = "")
        {
            var res = QueryScalar<T>(SqlGetWhere(fieldName, where, sOrderby));
            return res ?? default;
        }

        public static string SqlNow()
        {
            return IsPostgreSql 
                ? $"SELECT to_char({SqlDateTime}, 'YYYY-MM-DD HH24:MI:SS')"
                : $"SELECT CONVERT(VARCHAR(19), {SqlDateTime}, 120)";
        }

        public static string GetDate(string format = "yyyy-MM-dd HH:mm:ss")
        {
            return (Query<string>(SqlNow()).FirstOrDefault() ?? "").AsDateTimeFormat(format);
        }

        public static DateTime GetDateTime()
        {
            return Convert.ToDateTime(GetDate());
        }

        public static string GetDateTime(string fieldName, object value, object? value2 = null, string format = "yyyy-MM-dd HH:mm:ss")
        {
            string cast = IsPostgreSql ? $"to_char({Quote(fieldName)}, 'YYYY-MM-DD HH24:MI:SS')" : $"CONVERT(VARCHAR(50), {Quote(fieldName)}, 120)";
            return GetValue(cast, value, value2).AsDateTimeFormat(format);
        }

        public static string GetNewId()
        {
            return Query<string>($"SELECT {SqlRandomId}").FirstOrDefault() ?? "";
        }

        public static string MaxId()
        {
            return Query<string>($"SELECT MAX({Quote(Key)}) FROM {FullName}").FirstOrDefault() ?? "";
        }

        // ----------- Count 统计 -----------

        public static (string, IDataParameter[]) SqlCount(string tableFullName, Dictionary<string, object?> keys)
        {
            var (where, parameters) = SqlWhere(keys, allowEmpty: true);
            return ($"SELECT COUNT(*) FROM {tableFullName} {where}", parameters);
        }

        public static (string, IDataParameter[]) SqlCount(string tableFullName, params (string, object?)[] keys)
            => SqlCount(tableFullName, ToDict(keys));

        public static long Count()
        {
            var (sql, parameters) = SqlCount(FullName);
            return QueryScalar<long>(sql, null, parameters);
        }

        public static async Task<long> CountAsync()
        {
            var (sql, parameters) = SqlCount(FullName);
            return (await QueryScalarAsync<long>(sql, null, parameters));
        }

        public static long CountWhere(string where, params IDataParameter[] parameters)
        {
            return QueryScalar<long>($"SELECT COUNT({Quote(Key)}) FROM {FullName} {where.EnsureStartsWith("WHERE")}", null, parameters).AsLong();
        }

        public static async Task<long> CountWhereAsync(string where, params object?[] args)
        {
            var (sql, parameters) = ResolveSql(where, args);
            return await CountWhereAsync(sql, parameters);
        }

        public static async Task<long> CountWhereAsync(string where, IDataParameter[] parameters)
        {
            return (await QueryScalarAsync<long>($"SELECT COUNT({Quote(Key)}) FROM {FullName} {where.EnsureStartsWith("WHERE")}", null, parameters)).AsLong();
        }

        public static long CountByKeyValue(string field, string key, object value)
        {
            return QueryScalar<long>($"SELECT COUNT({Quote(field)}) FROM {FullName} WHERE {Quote(key)} = @p1", null, SqlParams(("@p1", value))).AsLong();
        }

        public static async Task<long> CountByKeyValueAsync(string field, string key, object value)
        {
            return (await QueryScalarAsync<long>($"SELECT COUNT({Quote(field)}) FROM {FullName} WHERE {Quote(key)} = @p1", null, SqlParams(("@p1", value)))).AsLong();
        }

        public static long CountField(string fieldName, string keyField, object fieldValue)
        {
            return CountByKeyValue(fieldName, keyField, fieldValue);
        }

        public static async Task<long> CountFieldAsync(string fieldName, string keyField, object fieldValue)
        {
            return await CountByKeyValueAsync(fieldName, keyField, fieldValue);
        }

        public static long CountKey(object id)
        {
            return CountByKeyValue(Key, Key2, id);
        }

        public static async Task<long> CountKeyAsync(object id)
        {
            return await CountByKeyValueAsync(Key, Key2, id);
        }

        public static long CountKey2(object id)
        {
            return CountByKeyValue(Key2, Key, id);
        }

        public static async Task<long> CountKey2Async(object id)
        {
            return await CountByKeyValueAsync(Key2, Key, id);
        }

        // ----------- Exists 存在性检查 -----------

        public static (string, IDataParameter[]) SqlExists(object id, object? id2 = null)
        {
            var (where, parameters) = SqlWhere(id, id2);
            string sql = $"SELECT {SqlTop(1)} 1 FROM {FullName} {where} {SqlLimit(1)}";
            return (sql, parameters);
        }

        public static (string, IDataParameter[]) SqlExists(params (string, object?)[] keys)
            => SqlExists(ToDict(keys));

        public static (string, IDataParameter[]) SqlExists(string fieldName, object value, string fieldName2, object? value2)
        {
            var keys = new Dictionary<string, object?>
            {
                { fieldName, value ?? DBNull.Value },
                { fieldName2, value2 ?? DBNull.Value}
            };
            return SqlExists(FullName, keys);
        }

        public static (string, IDataParameter[]) SqlExists(string tableFullName, Dictionary<string, object?> keys)
        {
            var (where, parameters) = SqlWhere(keys, allowEmpty: false);
            string sql = $"SELECT {SqlTop(1)} 1 FROM {tableFullName} {where} {SqlLimit(1)}";
            return (sql, parameters);
        }

        public static bool Exists(object id, object? id2 = null)
        {
            var (sql, parameters) = SqlExists(id, id2);
            var result = QueryScalar<object>(sql, null, parameters);
            return result != null && result != DBNull.Value;
        }

        public static async Task<bool> ExistsAsync(object id, object? id2 = null, IDbTransaction? trans = null)
        {
            var (sql, parameters) = SqlExists(id, id2);
            var result = await QueryScalarAsync<object>(sql, trans, parameters);
            return result != null && result != DBNull.Value;
        }

        public static (string, IDataParameter[]) SqlExistsAandB(string fieldName, object value, string fieldName2, object? value2)
        {
            string sql = $"SELECT {SqlTop(1)} 1 FROM {FullName} WHERE {Quote(fieldName)} = @p1 AND {Quote(fieldName2)} = @p2 {SqlLimit(1)}";
            return (sql, SqlParams(("@p1", value), ("@p2", value2)));
        }

        public static bool ExistsAandB(string fieldName, object value, string fieldName2, object? value2)
        {
            var (sql, parameters) = SqlExistsAandB(fieldName, value, fieldName2, value2);
            var result = QueryScalar<object>(sql, null, parameters);
            return result != null && result != DBNull.Value;
        }

        public static async Task<bool> ExistsAandBAsync(string fieldName, object value, string fieldName2, object? value2, IDbTransaction? trans = null)
        {
            var (sql, parameters) = SqlExistsAandB(fieldName, value, fieldName2, value2);
            var result = await QueryScalarAsync<object>(sql, trans, parameters);
            return result != null && result != DBNull.Value;
        }

        public static bool ExistsField(string fieldName, object value)
        {
            string sql = $"SELECT {SqlTop(1)} 1 FROM {FullName} WHERE {Quote(fieldName)} = @p1 {SqlLimit(1)}";
            var result = QueryScalar<object>(sql, null, SqlParams(("@p1", value)));
            return result != null && result != DBNull.Value;
        }

        public static async Task<bool> ExistsFieldAsync(string fieldName, object value, IDbTransaction? trans = null)
        {
            string sql = $"SELECT {SqlTop(1)} 1 FROM {FullName} WHERE {Quote(fieldName)} = @p1 {SqlLimit(1)}";
            var result = await QueryScalarAsync<object>(sql, trans, SqlParams(("@p1", value)));
            return result != null && result != DBNull.Value;
        }

        public static bool ExistsWhere(string sWhere)
        {
            string sql = $"SELECT {SqlTop(1)} 1 FROM {FullName} {sWhere.EnsureStartsWith("WHERE")} {SqlLimit(1)}";
            var result = QueryScalar<object>(sql, null);
            return result != null && result != DBNull.Value;
        }

        public static async Task<bool> ExistsWhereAsync(string sWhere, IDbTransaction? trans = null)
        {
            string sql = $"SELECT {SqlTop(1)} 1 FROM {FullName} {sWhere.EnsureStartsWith("WHERE")} {SqlLimit(1)}";
            var result = await QueryScalarAsync<object>(sql, trans);
            return result != null && result != DBNull.Value;
        }

        // ----------- Insert 插入 -----------

        public virtual async Task<int> InsertAsync(IDbTransaction? trans = null)
        {
            var data = ToDictionary();
            var exclude = new List<string> { "Guid" };

            // 改进主键排除逻辑：如果主键是字符串或 Guid 且有值，则包含在 INSERT 中；如果是数字则排除（由数据库生成）
            if (data.TryGetValue(Key, out var idValue) || data.TryGetValue(Key.ToLower(), out idValue) || data.TryGetValue(Key.ToUpper(), out idValue))
            {
                bool shouldKeep = false;
                if (idValue is string s && !string.IsNullOrEmpty(s)) shouldKeep = true;
                else if (idValue is Guid g && g != Guid.Empty) shouldKeep = true;
                
                if (!shouldKeep)
                {
                    exclude.Add(Key);
                }
            }
            else
            {
                // 如果字典里没找到 Key，说明可能主键不是属性，或者命名不匹配，保守起见排除 Key
                exclude.Add(Key);
            }

            if (!string.IsNullOrEmpty(Key2))
            {
                if (data.TryGetValue(Key2, out var idValue2) || data.TryGetValue(Key2.ToLower(), out idValue2) || data.TryGetValue(Key2.ToUpper(), out idValue2))
                {
                    bool shouldKeep2 = false;
                    if (idValue2 is string s2 && !string.IsNullOrEmpty(s2)) shouldKeep2 = true;
                    else if (idValue2 is Guid g2 && g2 != Guid.Empty) shouldKeep2 = true;

                    if (!shouldKeep2)
                    {
                        exclude.Add(Key2);
                    }
                }
                else
                {
                    exclude.Add(Key2);
                }
            }

            var (sql, paras) = SqlInsertDict(data, [], exclude.ToArray());
            return await ExecAsync(sql, trans, paras);
        }

        public virtual async Task<T?> InsertAsync<T>(string field, IDbTransaction? trans = null) where T : struct
        {
            var data = ToDictionary();
            var exclude = new List<string> { "Guid" };

            if (data.TryGetValue(Key, out var idValue) || data.TryGetValue(Key.ToLower(), out idValue) || data.TryGetValue(Key.ToUpper(), out idValue))
            {
                bool shouldKeep = false;
                if (idValue is string s && !string.IsNullOrEmpty(s)) shouldKeep = true;
                else if (idValue is Guid g && g != Guid.Empty) shouldKeep = true;
                
                if (!shouldKeep)
                {
                    exclude.Add(Key);
                }
            }
            else
            {
                exclude.Add(Key);
            }

            var (sql, paras) = SqlInsertDict(data, [field], exclude.ToArray());
            var dict = await Database.SQLConn.ExecWithOutputAsync(sql, paras, [field], trans);
            return dict.TryGetValue(field, out var val) ? (T?)Convert.ChangeType(val, typeof(T)) : null;
        }

        public virtual async Task<(T1?, T2?)> InsertAsync<T1, T2>(string field1, string field2, IDbTransaction? trans = null) where T1 : struct where T2 : struct
        {
            var data = ToDictionary();
            var exclude = new List<string> { "Guid" };

            if (data.TryGetValue(Key, out var idValue) || data.TryGetValue(Key.ToLower(), out idValue) || data.TryGetValue(Key.ToUpper(), out idValue))
            {
                bool shouldKeep = false;
                if (idValue is string s && !string.IsNullOrEmpty(s)) shouldKeep = true;
                else if (idValue is Guid g && g != Guid.Empty) shouldKeep = true;
                
                if (!shouldKeep)
                {
                    exclude.Add(Key);
                }
            }
            else
            {
                exclude.Add(Key);
            }

            var (sql, paras) = SqlInsertDict(data, [field1, field2], exclude.ToArray());
            var dict = await Database.SQLConn.ExecWithOutputAsync(sql, paras, [field1, field2], trans);

            var val1 = dict.TryGetValue(field1, out var v1) ? (T1?)Convert.ChangeType(v1, typeof(T1)) : null;
            var val2 = dict.TryGetValue(field2, out var v2) ? (T2?)Convert.ChangeType(v2, typeof(T2)) : null;

            return (val1, val2);
        }

        public static async Task<int> InsertObjectAsync(object obj)
            => await InsertObjectAsync(obj, null);

        public static async Task<int> InsertObjectAsync(object obj, IDbTransaction? trans)
        {
            var (sql, paras) = SqlInsertDict(obj.ToFields());
            return await ExecScalarAsync<int>(sql, trans, paras);
        }

        public static async Task<Dictionary<string, object>?> InsertReturnFieldsAsync(object obj, params string[] outputFields)
            => await InsertReturnFieldsAsync(obj, null, outputFields);

        public static async Task<Dictionary<string, object>?> InsertReturnFieldsAsync(object obj, IDbTransaction? trans, params string[] outputFields)
        {
            var fields = obj.ToFields();
            var (sql, paras) = SqlInsertDict(fields, outputFields);
            var result = await ExecWithOutputAsync(sql, paras, outputFields, trans);
            if (result != null && result.Count > 0)
             {
                 var id = result.TryGetValue(Key, out var v1) ? v1 : (fields.TryGetValue(Key, out var v2) ? v2 : null);
                 var id2 = !string.IsNullOrEmpty(Key2) ? (result.TryGetValue(Key2, out var v3) ? v3 : (fields.TryGetValue(Key2, out var v4) ? v4 : null)) : null;
                 if (id != null)
                 {
                     await InvalidateAllCachesAsync(id, id2);
                     // Logger.Debug($"[Cache] InvalidateAll(InsertReturn) - Table: {FullName}, ID: {id}-{id2}");
                 }
             }
            return result;
        }

        public static int Insert(object columns, IDbTransaction? trans = null)
        {
            return Insert(MetaData.ToCovList(columns), trans);
        }

        public static async Task<int> InsertAsync(object columns, IDbTransaction? trans = null)
        {
            return await InsertAsync(MetaData.ToCovList(columns), trans);
        }

        public static int Insert(List<Cov> columns, IDbTransaction? trans = null)
        {
            var data = CovToParams(columns);
            var (sql, paras) = SqlInsertDict(data);
            return Exec(sql, trans, paras);
        }

        public static async Task<int> InsertAsync(List<Cov> columns, IDbTransaction? trans = null)
        {
            var data = CovToParams(columns);
            var exclude = new List<string> { "Guid" };

            // 改进主键排除逻辑：如果主键是字符串或 Guid 且有值，则包含在 INSERT 中；如果是数字则排除（由数据库生成）
            if (data.TryGetValue(Key, out var idValue) || data.TryGetValue(Key.ToLower(), out idValue) || data.TryGetValue(Key.ToUpper(), out idValue))
            {
                bool shouldKeep = false;
                if (idValue is string s && !string.IsNullOrEmpty(s)) shouldKeep = true;
                else if (idValue is Guid g && g != Guid.Empty) shouldKeep = true;
                else if (idValue is long l && l != 0) shouldKeep = true;
                else if (idValue is int i && i != 0) shouldKeep = true;
                
                if (!shouldKeep)
                {
                    exclude.Add(Key);
                }
            }
            else
            {
                // 如果是 Cov 列表，且没有包含 Key，说明可能希望数据库生成主键
                // 但如果是 Guid 或 String 类型的主键，通常需要显式提供
                // 这里我们默认如果不提供且数据库不是自增，可能会报错，但由 SqlInsertDict 处理
            }

            var (sql, paras) = SqlInsertDict(data, [], exclude.ToArray());
            var result = await ExecAsync(sql, trans, paras);
            if (result > 0)
             {
                 // 如果插入的数据包含主键，则尝试失效相关缓存（防止之前缓存了空结果）
                 var id = columns.FirstOrDefault(c => c.Name.Equals(Key, StringComparison.OrdinalIgnoreCase))?.Value;
                 var id2 = !string.IsNullOrEmpty(Key2) ? columns.FirstOrDefault(c => c.Name.Equals(Key2, StringComparison.OrdinalIgnoreCase))?.Value : null;
                 if (id != null)
                 {
                     await InvalidateAllCachesAsync(id, id2);
                     // Logger.Debug($"[Cache] InvalidateAll(Insert) - Table: {FullName}, ID: {id}-{id2}");
                 }
             }
            return result;
        }

        public static (string sql, IDataParameter[] paras) SqlInsert(object columns)
        {
            return SqlInsert(MetaData.ToCovList(columns));
        }

        public static (string sql, IDataParameter[] paras) SqlInsert(List<Cov> columns)
        {
            var data = CovToParams(columns);
            return SqlInsertDict(data);
        }

        public static (string sql, IDataParameter[] parameters) SqlInsertDict(
            Dictionary<string, object?> data,
            string[]? outputFields = null,
            string[]? excludeFields = null)
        {
            if (data == null || data.Count == 0)
                throw new ArgumentException("INSERT 操作必须包含字段和值");

            excludeFields ??= [];
            outputFields ??= [];

            var fields = new List<string>();
            var values = new List<string>();
            var parameters = new List<IDataParameter>();
            int i = 0;

            foreach (var kvp in data)
            {
                if (excludeFields.Contains(kvp.Key, StringComparer.OrdinalIgnoreCase))
                    continue;

                i++;
                string field = kvp.Key;
                object? value = kvp.Value;

                if (value is DateTime dt && dt == DateTime.MinValue)
                {
                    fields.Add(MetaData.SqlQuote(field));
                    values.Add(SqlDateTime);
                }
                else
                {
                    string paramName = $"@p{i}";
                    fields.Add(MetaData.SqlQuote(field));
                    values.Add(paramName);
                    parameters.Add(CreateParameter(paramName, value));
                }
            }

            string outputClause = "";
            string returningClause = "";

            if (outputFields.Length > 0)
            {
                if (IsPostgreSql)
                {
                    returningClause = $" RETURNING {string.Join(", ", outputFields.Select(f => MetaData.SqlQuote(f)))}";
                }
                else
                {
                    var insertedFields = string.Join(", ", outputFields.Select(f => $"INSERTED.{MetaData.SqlQuote(f)}"));
                    outputClause = $"OUTPUT {insertedFields} ";
                }
            }
           
            var sql = $"INSERT INTO {FullName} ({string.Join(", ", fields)}) {outputClause}VALUES ({string.Join(", ", values)}){returningClause}";
            return (sql, parameters.ToArray());
        }

        public static async Task<int> BatchInsertAsync(IEnumerable<object> items, IDbTransaction? trans = null)
        {
            if (items == null || !items.Any()) return 0;

            if (IsPostgreSql)
            {
                var firstItem = items.First();
                var props = firstItem is IDictionary<string, object?> dict 
                    ? dict.Keys.ToArray() 
                    : firstItem.GetType().GetProperties()
                        .Where(p => p.CanRead && !p.GetMethod!.IsStatic && p.Name != "Id")
                        .Select(p => p.Name).ToArray();

                var sqlBase = $"INSERT INTO {FullName} ({string.Join(", ", props.Select(p => MetaData.SqlQuote(p)))}) VALUES ";
                var rows = new List<string>();
                var parameters = new List<IDataParameter>();
                int paramIdx = 1;

                foreach (var item in items)
                {
                    var rowValues = new List<string>();
                    var itemDict = item is IDictionary<string, object?> d ? d : item.ToFields();
                    foreach (var prop in props)
                    {
                        var val = itemDict.TryGetValue(prop, out var v) ? v : null;
                        if (val is DateTime dt && dt == DateTime.MinValue)
                        {
                            rowValues.Add(MetaData.SqlDateTime);
                        }
                        else
                        {
                            var pName = $"@p{paramIdx++}";
                            rowValues.Add(pName);
                            parameters.Add(MetaData.CreateParameter(pName, val));
                        }
                    }
                    rows.Add($"({string.Join(", ", rowValues)})");
                }

                var sql = sqlBase + string.Join(", ", rows);
                return await ExecAsync(sql, trans, parameters.ToArray());
            }
            else
            {
                int count = 0;
                foreach (var item in items)
                {
                    count += await InsertAsync(MetaData.ToCovList(item), trans);
                }
                return count;
            }
        }

        public static (string sql, IDataParameter[] parameters) SqlSetValues(string sSet, object id, object? id2 = null)
        {
            var (where, parameters) = SqlWhere(id, id2);
            return ($"UPDATE {FullName} SET {sSet} {where}", parameters);
        }

        public static (string sql, IDataParameter[] parameters) SqlUpdateWhere(object columns, string where, params object?[] args)
        {
            var data = MetaData.ToCovList(columns);
            var setList = new List<string>();
            var parameters = new List<IDataParameter>();
            int i = 0;
            foreach (var col in data)
            {
                i++;
                string paramName = $"@u{i}";
                setList.Add($"{MetaData.SqlQuote(col.Name)} = {paramName}");
                parameters.Add(CreateParameter(paramName, col.Value));
            }
            var (sqlWhere, whereParams) = ResolveSql(where.EnsureStartsWith("WHERE"), args);
            parameters.AddRange(whereParams);
            return ($"UPDATE {FullName} SET {string.Join(", ", setList)} {sqlWhere}", parameters.ToArray());
        }

        public static int SetNow(string fieldName, object id, object? id2 = null, IDbTransaction? trans = null)
        {
            var (sqlWhere, parameters) = SqlWhere(id, id2);
            var sql = $"UPDATE {FullName} SET {MetaData.SqlQuote(fieldName)} = {SqlDateTime} {sqlWhere}";
            var result = Exec(sql, trans, parameters);
            if (result > 0)
            {
                if (!IsHighFrequency(fieldName))
                {
                    InvalidateCache(id, id2);
                }
                InvalidateFieldCache(fieldName, id, id2);
            }
            return result;
        }

        private static bool IsHighFrequency(string fieldName)
        {
            var prop = GetProperties().FirstOrDefault(p => string.Equals(p.Name, fieldName, StringComparison.OrdinalIgnoreCase));
            return prop?.GetCustomAttribute<HighFrequencyAttribute>() != null;
        }

        public static async Task<int> PlusAsync(string fieldName, object plusValue, object id, object? id2 = null, IDbTransaction? trans = null)
        {
            var (sql, paras) = SqlPlus(fieldName, plusValue, id, id2);
            var result = await ExecAsync(sql, trans, paras);
            if (result > 0)
            {
                if (!IsHighFrequency(fieldName))
                {
                    await InvalidateCacheAsync(id, id2);
                }
                await InvalidateFieldCacheAsync(fieldName, id, id2);
                // Logger.Debug($"[Cache] Invalidate(Plus) - Table: {FullName}, Field: {fieldName}, ID: {id}-{id2}");
            }
            return result;
        }

        public static int Plus(string fieldName, object plusValue, object id, object? id2 = null, IDbTransaction? trans = null)
        {
            var (sql, paras) = SqlPlus(fieldName, plusValue, id, id2);
            var result = Exec(sql, trans, paras);
            if (result > 0)
            {
                if (!IsHighFrequency(fieldName))
                {
                    InvalidateCache(id, id2);
                }
                InvalidateFieldCache(fieldName, id, id2);
            }
            return result;
        }

        public static int SetValue(string fieldName, object value, object id, object? id2 = null, IDbTransaction? trans = null)
        {
            var (sql, parameters) = SqlUpdate(fieldName, value, id, id2);
            var result = Exec(sql, trans, parameters);
            if (result > 0)
            {
                if (!IsHighFrequency(fieldName))
                {
                    InvalidateCache(id, id2);
                }
                InvalidateFieldCache(fieldName, id, id2);
            }
            return result;
        }

        public static async Task<int> SetNowAsync(string fieldName, object id, object? id2 = null, IDbTransaction? trans = null)
        {
            return await UpdateAsync($"{fieldName} = {SqlDateTime}", id, id2, trans);
        }

        public static async Task<int> SetValueAsync(string fieldName, object value, object id, object? id2 = null, IDbTransaction? trans = null)
        {
            var (sql, parameters) = SqlUpdate(fieldName, value, id, id2);
            var result = await ExecAsync(sql, trans, parameters);
            if (result > 0)
            {
                if (!IsHighFrequency(fieldName))
                {
                    await InvalidateCacheAsync(id, id2);
                }
                await InvalidateFieldCacheAsync(fieldName, id, id2);
                // Logger.Debug($"[Cache] Invalidate(SetValue) - Table: {FullName}, Field: {fieldName}, ID: {id}-{id2}");
            }
            return result;
        }

        // ----------- Update 更新 -----------

        public virtual int Insert(IDbTransaction? trans = null)
        {
            var data = ToDictionary();
            var exclude = new List<string> { "Guid" };

            if (data.TryGetValue(Key, out var idValue) || data.TryGetValue(Key.ToLower(), out idValue) || data.TryGetValue(Key.ToUpper(), out idValue))
            {
                bool shouldKeep = false;
                if (idValue is string s && !string.IsNullOrEmpty(s)) shouldKeep = true;
                else if (idValue is Guid g && g != Guid.Empty) shouldKeep = true;
                
                if (!shouldKeep)
                {
                    exclude.Add(Key);
                }
            }
            else
            {
                exclude.Add(Key);
            }

            var (sql, paras) = SqlInsertDict(data, exclude.ToArray());
            var result = Exec(sql, trans, paras);
            if (result > 0)
            {
                var id = data.TryGetValue(Key, out var v1) ? v1 : null;
                var id2 = !string.IsNullOrEmpty(Key2) && data.TryGetValue(Key2, out var v2) ? v2 : null;
                if (id != null)
                {
                    InvalidateAllCaches(id, id2);
                }
            }
            return result;
        }

        public virtual int Update(IDbTransaction? trans = null, params string[] excludeFields)
        {
            var data = ToDictionary();
            var exclude = new HashSet<string>(StringComparer.OrdinalIgnoreCase)
            {
                Key,
                "Id",
                "Guid",
            };
            if (!string.IsNullOrEmpty(Key2)) exclude.Add(Key2);
            foreach (var f in excludeFields) exclude.Add(f);

            var setData = data.Where(kvp => !exclude.Contains(kvp.Key)).ToDictionary(kvp => kvp.Key, kvp => kvp.Value);
            var whereData = data.Where(kvp => string.Equals(kvp.Key, Key, StringComparison.OrdinalIgnoreCase) || string.Equals(kvp.Key, Key2, StringComparison.OrdinalIgnoreCase)).ToDictionary(kvp => kvp.Key, kvp => kvp.Value);

            if (whereData.Count == 0) throw new InvalidOperationException("主键字段未赋值，无法更新");

            var (sql, paras) = SqlUpdate(setData, whereData);
            var result = Exec(sql, trans, paras);
            if (result > 0)
            {
                var id1 = whereData.Values.First();
                var id2 = whereData.Count > 1 ? whereData.Values.ElementAt(1) : null;

                bool hasNormalField = setData.Keys.Any(f => !IsHighFrequency(f));
                if (hasNormalField)
                {
                    InvalidateCache(id1!, id2);
                }

                foreach (var kv in setData)
                {
                    InvalidateFieldCache(kv.Key, id1!, id2);
                }
            }
            return result;
        }

        public virtual async Task<int> UpdateAsync(IDbTransaction? trans = null, params string[] excludeFields)
        {
            var data = ToDictionary();
            var exclude = new HashSet<string>(StringComparer.OrdinalIgnoreCase)
            {
                Key,
                "Id",
                "Guid",
            };
            if (!string.IsNullOrEmpty(Key2)) exclude.Add(Key2);
            foreach (var f in excludeFields) exclude.Add(f);

            var setData = data.Where(kvp => !exclude.Contains(kvp.Key)).ToDictionary(kvp => kvp.Key, kvp => kvp.Value);
            var whereData = data.Where(kvp => string.Equals(kvp.Key, Key, StringComparison.OrdinalIgnoreCase) || string.Equals(kvp.Key, Key2, StringComparison.OrdinalIgnoreCase)).ToDictionary(kvp => kvp.Key, kvp => kvp.Value);

            if (whereData.Count == 0) throw new InvalidOperationException("主键字段未赋值，无法更新");

            var (sql, paras) = SqlUpdate(setData, whereData);
            var result = await ExecAsync(sql, trans, paras);
            if (result > 0)
            {
                var id1 = whereData.Values.First();
                var id2 = whereData.Count > 1 ? whereData.Values.ElementAt(1) : null;

                // 检查 setData 中是否包含非高频字段。
                // 如果全是高频字段，则不失效行缓存。
                bool hasNormalField = setData.Keys.Any(f => !IsHighFrequency(f));
                if (hasNormalField)
                {
                    await InvalidateCacheAsync(id1!, id2);
                }

                foreach (var kv in setData)
                {
                    await InvalidateFieldCacheAsync(kv.Key, id1!, id2);
                }
                // Logger.Debug($"[Cache] Invalidate(UpdateInstance) - Table: {FullName}, ID: {id1}-{id2}, Fields: {string.Join(", ", setData.Keys)}");
            }
            return result;
        }

        public static (string sql, IDataParameter[] paras) SqlUpdate<T>(T entity, object id, object? id2 = null) where T : class
        {
            return SqlUpdate(entity, ToDict(id, id2));
        }

        public static (string sql, IDataParameter[] paras) SqlUpdate(string fieldName, object value, object id, object? id2 = null)
        {
            var setValues = new Dictionary<string, object?> { { fieldName, value } };
            return SqlUpdate(setValues, ToDict(id, id2));
        }

        public static (string sql, IDataParameter[] paras) SqlUpdate(Dictionary<string, object?> setValues, Dictionary<string, object?> whereKeys)
        {
            if (setValues.Count == 0) throw new ArgumentException("UPDATE 操作必须指定至少一个 SET 字段");

            var sb = new StringBuilder($"UPDATE {FullName} SET ");
            var parameters = new List<IDataParameter>();
            int i = 0;

            foreach (var kvp in setValues)
            {
                if (i++ > 0) sb.Append(", ");
                string field = kvp.Key;
                object? value = kvp.Value;

                if (value is DateTime dt && dt == DateTime.MinValue)
                    sb.Append($"{MetaData.SqlQuote(field)} = {SqlDateTime}");
                else
                {
                    string paramName = $"@u{i}";
                    sb.Append($"{MetaData.SqlQuote(field)} = {paramName}");
                    parameters.Add(CreateParameter(paramName, value));
                }
            }

            var (whereClause, whereParams) = SqlWhere(whereKeys, allowEmpty: false);
            parameters.AddRange(whereParams);

            return ($"{sb} {whereClause}", [.. parameters]);
        }

        public static (string sql, IDataParameter[] paras) SqlUpdate(List<Cov> updateColumns, object id, object? id2 = null)
        {
            return SqlUpdate(updateColumns, ToDict(id, id2));
        }

        public static int Update(object columns, object id, object? id2 = null, IDbTransaction? trans = null)
            => Update(MetaData.ToCovList(columns), id, id2, trans);

        public static async Task<int> UpdateAsync(object columns, object id, object? id2 = null, IDbTransaction? trans = null)
            => await UpdateAsync(MetaData.ToCovList(columns), id, id2, trans);

        public static int Update(List<Cov> columns, object id, object? id2 = null, IDbTransaction? trans = null)
        {
            var (sql, paras) = SqlUpdate(columns, ToDict(id, id2));
            var result = Exec(sql, trans, paras);
            if (result > 0)
            {
                InvalidateCache(id, id2);
                foreach (var col in columns)
                {
                    InvalidateFieldCache(col.Name, id, id2);
                }
            }
            return result;
        }

        public static async Task<int> UpdateAsync(List<Cov> columns, object id, object? id2 = null, IDbTransaction? trans = null)
        {
            var (sql, paras) = SqlUpdate(columns, ToDict(id, id2));
            var result = await ExecAsync(sql, trans, paras);
            if (result > 0)
            {
                await InvalidateCacheAsync(id, id2);
                foreach (var col in columns)
                {
                    await InvalidateFieldCacheAsync(col.Name, id, id2);
                }
                // Logger.Debug($"[Cache] Invalidate(UpdateStatic) - Table: {FullName}, ID: {id}-{id2}, Fields: {string.Join(", ", columns.Select(c => c.Name))}");
            }
            return result;
        }

        public static async Task<int> UpdateObjectAsync(object obj, object id, object? id2 = null, IDbTransaction? trans = null)
        {
            var fields = obj.ToFields();
            var (sql, paras) = SqlUpdate(fields, ToDict(id, id2));
            var result = await ExecAsync(sql, trans, paras);
            if (result > 0)
            {
                await InvalidateCacheAsync(id, id2);
                foreach (var kv in fields)
                {
                    await InvalidateFieldCacheAsync(kv.Key, id, id2);
                }
                // Logger.Debug($"[Cache] Invalidate(UpdateObject) - Table: {FullName}, ID: {id}-{id2}, Fields: {string.Join(", ", fields.Keys)}");
            }
            return result;
        }

        public static int Update(string sSet, object id, object? id2 = null, IDbTransaction? trans = null)
        {
            var (sqlWhere, parameters) = SqlWhere(id, id2);
            var sql = $"UPDATE {FullName} SET {sSet} {sqlWhere}";
            var result = Exec(sql, trans, parameters);
            if (result > 0) 
            {
                InvalidateAllCaches(id, id2);
            }
            return result;
        }

        public static async Task<int> UpdateAsync(string sSet, object id, object? id2 = null, IDbTransaction? trans = null)
        {
            var (sqlWhere, parameters) = SqlWhere(id, id2);
            var sql = $"UPDATE {FullName} SET {sSet} {sqlWhere}";
            var result = await ExecAsync(sql, trans, parameters);
            if (result > 0) 
            {
                await InvalidateAllCachesAsync(id, id2);
                // Logger.Debug($"[Cache] InvalidateAll(UpdateSet) - Table: {FullName}, ID: {id}-{id2}");
            }
            return result;
        }

        public static int UpdateWhere(string sSet, string sWhere, IDbTransaction? trans = null, params object?[] args)
        {
            var (sqlWhere, parameters) = ResolveSql(sWhere.EnsureStartsWith("WHERE"), args);
            var sql = $"UPDATE {FullName} SET {sSet} {sqlWhere}";
            var result = Exec(sql, trans, parameters);
            if (result > 0)
            {
                var idCandidates = ExtractIdsFromWhere(sWhere, args);
                foreach (var cand in idCandidates)
                {
                    InvalidateAllCaches(cand.id1, cand.id2);
                }
            }
            return result;
        }

        public static async Task<int> UpdateWhereAsync(string sSet, string sWhere, IDbTransaction? trans = null, params object?[] args)
        {
            var (sqlWhere, parameters) = ResolveSql(sWhere.EnsureStartsWith("WHERE"), args);
            var sql = $"UPDATE {FullName} SET {sSet} {sqlWhere}";
            var result = await ExecAsync(sql, trans, parameters);
            // Logger.Debug($"[DB] UpdateWhereAsync - Table: {FullName}, SQL: {sql}, Result: {result}");
            if (result > 0)
            {
                // 尝试解析 ID 并失效缓存
                var idCandidates = ExtractIdsFromWhere(sWhere, args);
                // Logger.Debug($"[Cache] Invalidate - Table: {FullName}, Extracted IDs: {string.Join(", ", idCandidates)}");
                foreach (var cand in idCandidates)
                {
                    await InvalidateAllCachesAsync(cand.id1, cand.id2);
                }
            }
            return result;
        }

        private static List<(object id1, object? id2)> ExtractIdsFromWhere(string sWhere, object?[] args)
        {
            var results = new List<(object id1, object? id2)>();
            if (args == null || args.Length == 0) return results;

            try
            {
                // 1. 尝试正则匹配主键 = {n}
                // 支持格式：Key = {0} AND Key2 = {1}
                var keyMatches = System.Text.RegularExpressions.Regex.Matches(sWhere, $@"{Key}\s*=\s*\{{(\d+)\}}", System.Text.RegularExpressions.RegexOptions.IgnoreCase);
                var key2Matches = string.IsNullOrEmpty(Key2) ? null : System.Text.RegularExpressions.Regex.Matches(sWhere, $@"{Key2}\s*=\s*\{{(\d+)\}}", System.Text.RegularExpressions.RegexOptions.IgnoreCase);

                foreach (System.Text.RegularExpressions.Match km in keyMatches)
                {
                    if (int.TryParse(km.Groups[1].Value, out int idx1) && idx1 < args.Length && args[idx1] != null)
                    {
                        object? id2 = null;
                        // 寻找同一个 WHERE 子句中是否有对应的 Key2
                        if (key2Matches != null)
                        {
                            foreach (System.Text.RegularExpressions.Match k2m in key2Matches)
                            {
                                if (int.TryParse(k2m.Groups[1].Value, out int idx2) && idx2 < args.Length)
                                {
                                    id2 = args[idx2];
                                    break; 
                                }
                            }
                        }
                        results.Add((args[idx1]!, id2));
                    }
                }

                // 2. 尝试匹配 IN 子句 (仅支持单主键)
                var inMatches = System.Text.RegularExpressions.Regex.Matches(sWhere, $@"{Key}\s+IN\s*\((.*?)\)", System.Text.RegularExpressions.RegexOptions.IgnoreCase);
                foreach (System.Text.RegularExpressions.Match im in inMatches)
                {
                    var content = im.Groups[1].Value;
                    var subMatches = System.Text.RegularExpressions.Regex.Matches(content, @"\{(\d+)\}");
                    foreach (System.Text.RegularExpressions.Match sm in subMatches)
                    {
                        if (int.TryParse(sm.Groups[1].Value, out int idx) && idx < args.Length && args[idx] != null)
                        {
                            results.Add((args[idx]!, null));
                        }
                    }
                }

                // 3. 兜底启发式逻辑：如果没匹配到但有参数，第一个参数通常是 ID
                if (results.Count == 0 && args.Length > 0 && args[0] != null)
                {
                    object? id2 = args.Length > 1 && !string.IsNullOrEmpty(Key2) ? args[1] : null;
                    results.Add((args[0]!, id2));
                }
            }
            catch { /* ignore */ }

            // 去重
            return results.GroupBy(x => new { x.id1, x.id2 })
                          .Select(g => (g.Key.id1, g.Key.id2))
                          .ToList();
        }

        public static async Task<int> UpdateWhereAsync(object columns, string sWhere, params object?[] args)
            => await UpdateWhereAsync(columns, sWhere, null, args);

        public static async Task<int> UpdateWhereAsync(object columns, string sWhere, IDbTransaction? trans, params object?[] args)
        {
            var data = MetaData.ToCovList(columns);
            var setList = new List<string>();
            var parameters = new List<IDataParameter>();
            int i = 0;

            foreach (var col in data)
            {
                i++;
                string paramName = $"@u{i}";
                setList.Add($"{MetaData.SqlQuote(col.Name)} = {paramName}");
                parameters.Add(CreateParameter(paramName, col.Value));
            }

            var (sqlWhere, whereParams) = ResolveSql(sWhere.EnsureStartsWith("WHERE"), args);
            parameters.AddRange(whereParams);

            var sql = $"UPDATE {FullName} SET {string.Join(", ", setList)} {sqlWhere}";
            var result = await ExecAsync(sql, trans, [.. parameters]);
            if (result > 0)
            {
                var idCandidates = ExtractIdsFromWhere(sWhere, args);
                foreach (var cand in idCandidates)
                {
                    await InvalidateCacheAsync(cand.id1, cand.id2);
                    foreach (var col in data)
                    {
                        await InvalidateFieldCacheAsync(col.Name, cand.id1, cand.id2);
                    }
                }
            }
            return result;
        }

        public static int UpdateWhere(object columns, string sWhere, IDbTransaction? trans = null, params object?[] args)
        {
            var data = MetaData.ToCovList(columns);
            var setList = new List<string>();
            var parameters = new List<IDataParameter>();
            int i = 0;

            foreach (var col in data)
            {
                i++;
                string paramName = $"@u{i}";
                setList.Add($"{MetaData.SqlQuote(col.Name)} = {paramName}");
                parameters.Add(CreateParameter(paramName, col.Value));
            }

            var (sqlWhere, whereParams) = ResolveSql(sWhere.EnsureStartsWith("WHERE"), args);
            parameters.AddRange(whereParams);

            var sql = $"UPDATE {FullName} SET {string.Join(", ", setList)} {sqlWhere}";
            var result = Exec(sql, trans, [.. parameters]);
            if (result > 0)
            {
                var idCandidates = ExtractIdsFromWhere(sWhere, args);
                foreach (var cand in idCandidates)
                {
                    InvalidateCache(cand.id1, cand.id2);
                    foreach (var col in data)
                    {
                        InvalidateFieldCache(col.Name, cand.id1, cand.id2);
                    }
                }
            }
            return result;
        }

        public static async Task<int> IncrementAsync(object increments, string sWhere, IDbTransaction? trans = null, params object?[] args)
        {
            var data = MetaData.ToCovList(increments);
            var setList = new List<string>();
            var parameters = new List<IDataParameter>();
            int i = 0;

            foreach (var col in data)
            {
                i++;
                string paramName = $"@i{i}";
                string field = MetaData.SqlQuote(col.Name);
                setList.Add($"{field} = {SqlIsNull(field, "0")} + {paramName}");
                parameters.Add(CreateParameter(paramName, col.Value));
            }

            var (sqlWhere, whereParams) = ResolveSql(sWhere.EnsureStartsWith("WHERE"), args);
            parameters.AddRange(whereParams);

            var sql = $"UPDATE {FullName} SET {string.Join(", ", setList)} {sqlWhere}";
            var result = await ExecAsync(sql, trans, [.. parameters]);
            
            // 如果 sWhere 是简单的 ID 匹配，我们需要失效缓存
            if (result > 0)
            {
                var idCandidates = ExtractIdsFromWhere(sWhere, args);
                foreach (var cand in idCandidates)
                {
                    await InvalidateCacheAsync(cand.id1, cand.id2);
                    foreach (var col in data)
                    {
                        await InvalidateFieldCacheAsync(col.Name, cand.id1, cand.id2);
                    }
                }
            }
            return result;
        }

        public static int Increment(object increments, string sWhere, IDbTransaction? trans = null, params object?[] args)
        {
            var data = MetaData.ToCovList(increments);
            var setList = new List<string>();
            var parameters = new List<IDataParameter>();
            int i = 0;

            foreach (var col in data)
            {
                i++;
                string paramName = $"@i{i}";
                string field = MetaData.SqlQuote(col.Name);
                setList.Add($"{field} = {SqlIsNull(field, "0")} + {paramName}");
                parameters.Add(CreateParameter(paramName, col.Value));
            }

            var (sqlWhere, whereParams) = ResolveSql(sWhere.EnsureStartsWith("WHERE"), args);
            parameters.AddRange(whereParams);

            var sql = $"UPDATE {FullName} SET {string.Join(", ", setList)} {sqlWhere}";
            var result = Exec(sql, trans, [.. parameters]);
            
            // 如果 sWhere 是简单的 ID 匹配，我们需要失效缓存
            if (result > 0)
            {
                var idCandidates = ExtractIdsFromWhere(sWhere, args);
                foreach (var cand in idCandidates)
                {
                    InvalidateCache(cand.id1, cand.id2);
                    foreach (var col in data)
                    {
                        InvalidateFieldCache(col.Name, cand.id1, cand.id2);
                    }
                }
            }
            return result;
        }

        public static (string sql, IDataParameter[] paras) SqlPlus(string fieldName, object plusValue, object id, object? id2 = null)
        {
            var whereDict = ToDict(id, id2);
            var (whereClause, whereParams) = SqlWhere(whereDict, allowEmpty: false);

            string field = MetaData.SqlQuote(fieldName);
            string sql = $"UPDATE {FullName} SET {field} = " + SqlIsNull(field, "0") + $" + @plusValue {whereClause}";

            var paramList = whereParams.ToList<IDataParameter>();
            paramList.Add(CreateParameter("@plusValue", plusValue ?? 0));

            return (sql, paramList.ToArray());
        }

        public static (string sql, IDataParameter[] paras) SqlUpdateOther(string fieldName, object fieldValue, object id, object? id2 = null)
        {
            var (where, paras) = SqlWhere(id, id2);
            return ($"UPDATE {FullName} SET {fieldName} = {fieldValue} {where}", paras);
        }

        // ----------- Delete 删除 -----------

        public virtual async Task<int> DeleteAsync(IDbTransaction? trans = null)
        {
            var where = new Dictionary<string, object?>();
            var type = GetType();

            var key1Value = (type.GetProperty(Key)?.GetValue(this)) ?? throw new InvalidOperationException($"主键字段 {Key} 未赋值");
            where[Key] = key1Value;

            if (!string.IsNullOrEmpty(Key2))
            {
                var key2Value = (type.GetProperty(Key2)?.GetValue(this)) ?? throw new InvalidOperationException($"主键字段 {Key2} 未赋值");
                where[Key2] = key2Value;
            }

            var (whereSql, parameters) = SqlWhere(where, allowEmpty: false);
            var sql = $"DELETE FROM {GetFullName()} {whereSql}";
            var result = await ExecAsync(sql, trans, parameters);
            if (result > 0)
            {
                var id1 = where[Key];
                var id2 = where.ContainsKey(Key2) ? where[Key2] : null;
                await InvalidateCacheAsync(id1!, id2);
            }
            return result;
        }

        public static (string, IDataParameter[]) SqlDelete(object id, object? id2 = null)
        {
            var dict = ToDict(id, id2);
            return SqlDelete(FullName, dict);
        }

        public static (string, IDataParameter[]) SqlDelete(string tableFullName, Dictionary<string, object?> keys)
        {
            var (where, parameters) = SqlWhere(keys, allowEmpty: false);
            return ($"DELETE FROM {tableFullName} {where}", parameters);
        }

        public static (string, IDataParameter[]) SqlDelete(string tableFullName, params (string, object?)[] keys)
            => SqlDelete(tableFullName, ToDict(keys));

        public static int Delete(object id, object? id2 = null, IDbTransaction? trans = null)
        {
            var (sql, paras) = SqlDelete(id, id2);
            var result = Exec(sql, trans, paras);
            if (result > 0) 
            {
                InvalidateAllCaches(id, id2);
            }
            return result;
        }

        public static async Task<int> DeleteAsync(object id, object? id2 = null, IDbTransaction? trans = null)
        {
            var (sql, paras) = SqlDelete(id, id2);
            var result = await ExecAsync(sql, trans, paras);
            if (result > 0) 
            {
                await InvalidateAllCachesAsync(id, id2);
                // Logger.Debug($"[Cache] InvalidateAll(Delete) - Table: {FullName}, ID: {id}-{id2}");
            }
            return result;
        }

        public static string SqlDeleteAll(object value)
            => $"DELETE FROM {FullName} WHERE {Quote(Key)} = {value.AsString().Quotes()}";

        public static int DeleteAll(object value, IDbTransaction? trans = null)
        {
            var result = Exec(SqlDeleteAll(value), trans);
            if (result > 0)
            {
                InvalidateAllCaches(value);
            }
            return result;
        }

        public static async Task<int> DeleteAllAsync(object value, IDbTransaction? trans = null)
        {
            var result = await ExecAsync(SqlDeleteAll(value), trans);
            if (result > 0)
            {
                await InvalidateAllCachesAsync(value);
                // Logger.Debug($"[Cache] InvalidateAll(DeleteAll) - Table: {FullName}, ID: {value}");
            }
            return result;
        }

        public static string SqlDeleteAll2(object value)
            => $"DELETE FROM {FullName} WHERE {Quote(Key2)} = {value.AsString().Quotes()}";

        public static int DeleteAll2(object value, IDbTransaction? trans = null)
        {
            var result = Exec(SqlDeleteAll2(value), trans);
            if (result > 0 && !string.IsNullOrEmpty(Key2))
            {
                // 注意：由于是按 Key2 删除，可能影响多个 Key1。
                // 如果没有 Key1 信息，我们无法精确失效 Key1 的缓存。
                // 但目前的 InvalidateAllCaches(id, id2) 主要是按 Key1 索引的。
                // 这是一个潜在的缓存不一致点。
            }
            return result;
        }

        public static async Task<int> DeleteAll2Async(object value, IDbTransaction? trans = null)
        {
            var result = await ExecAsync(SqlDeleteAll2(value), trans);
            if (result > 0 && !string.IsNullOrEmpty(Key2))
            {
                // 同上
            }
            return result;
        }

        public static async Task<int> DeleteWhereAsync(string sWhere, IDbTransaction? trans = null, params object?[] args)
        {
            var (sql, paras) = ResolveSql($"DELETE FROM {FullName} {sWhere.EnsureStartsWith("WHERE")}", args);
            var result = await ExecAsync(sql, trans, paras);
            if (result > 0)
            {
                var idCandidates = ExtractIdsFromWhere(sWhere, args);
                foreach (var cand in idCandidates)
                {
                    await InvalidateAllCachesAsync(cand.id1, cand.id2);
                }
            }
            return result;
        }

        public static int DeleteWhere(string sWhere, IDbTransaction? trans = null, params object?[] args)
        {
            var (sql, paras) = ResolveSql($"DELETE FROM {FullName} {sWhere.EnsureStartsWith("WHERE")}", args);
            var result = Exec(sql, trans, paras);
            if (result > 0)
            {
                var idCandidates = ExtractIdsFromWhere(sWhere, args);
                foreach (var cand in idCandidates)
                {
                    InvalidateAllCaches(cand.id1, cand.id2);
                }
            }
            return result;
        }

        public static async Task<int> DeleteByKeyValueAsync(string key, object value, IDbTransaction? trans = null)
        {
            var (where, paras) = SqlWhere(new Dictionary<string, object?> { { key, value } }, allowEmpty: false);
            var result = await ExecAsync($"DELETE FROM {FullName} {where}", trans, paras);
            if (result > 0 && (key.Equals(Key, StringComparison.OrdinalIgnoreCase) || key.Equals("id", StringComparison.OrdinalIgnoreCase)))
            {
                await InvalidateAllCachesAsync(value);
                // Logger.Debug($"[Cache] InvalidateAll(DeleteByKV) - Table: {FullName}, {key}: {value}");
            }
            return result;
        }

        public static int DeleteByKeyValue(string key, object value, IDbTransaction? trans = null)
        {
            var (where, paras) = SqlWhere(new Dictionary<string, object?> { { key, value } }, allowEmpty: false);
            var result = Exec($"DELETE FROM {FullName} {where}", trans, paras);
            if (result > 0 && (key.Equals(Key, StringComparison.OrdinalIgnoreCase) || key.Equals("id", StringComparison.OrdinalIgnoreCase)))
            {
                InvalidateAllCaches(value);
            }
            return result;
        }

        // ----------- Save & Upsert 保存 & 更新插入 -----------

        public virtual async Task<int> SaveAsync(IDbTransaction? trans = null)
        {
            var keyValues = GetKeyValues();
            bool isNew = true;
            bool hasValidKey = false;
            foreach (var kv in keyValues)
            {
                if (kv.Value != null && !kv.Value.Equals(0) && !kv.Value.Equals(Guid.Empty) && !kv.Value.Equals(string.Empty) && !kv.Value.Equals(DBNull.Value))
                {
                    hasValidKey = true;
                    break;
                }
            }

            if (hasValidKey)
            {
                var keys = new Dictionary<string, object?>();
                foreach (var kv in keyValues) keys.Add(kv.Name, kv.Value);
                
                var (sql, parameters) = SqlExists(FullName, keys);
                var result = await QueryScalarAsync<object>(sql, trans, parameters);
                if (result != null && result != DBNull.Value) isNew = false;
            }

            return isNew ? await InsertAsync(trans) : await UpdateAsync(trans);
        }

        public virtual int Save(IDbTransaction? trans = null)
        {
            var keyValues = GetKeyValues();
            bool isNew = true;
            bool hasValidKey = false;
            foreach (var kv in keyValues)
            {
                if (kv.Value != null && !kv.Value.Equals(0) && !kv.Value.Equals(Guid.Empty) && !kv.Value.Equals(string.Empty) && !kv.Value.Equals(DBNull.Value))
                {
                    hasValidKey = true;
                    break;
                }
            }

            if (hasValidKey)
            {
                var keys = new Dictionary<string, object?>();
                foreach (var kv in keyValues) keys.Add(kv.Name, kv.Value);
                
                var (sql, parameters) = SqlExists(FullName, keys);
                var result = QueryScalar<object>(sql, trans, parameters);
                if (result != null && result != DBNull.Value) isNew = false;
            }

            return isNew ? Insert(trans) : Update(trans);
        }

        public static async Task<int> UpsertAsync(object columns, IDbTransaction? trans = null)
        {
            var data = MetaData.ToCovList(columns).ToDictionary(c => c.Name, c => c.Value);
            var keys = Keys.Where(k => !string.IsNullOrEmpty(k)).ToArray();

            if (keys.Length == 0) throw new InvalidOperationException("Upsert 操作必须定义主键字段");
            foreach (var key in keys)
            {
                if (!data.ContainsKey(key)) throw new ArgumentException($"Upsert 操作的数据中必须包含主键字段 [{key}]");
            }

            if (IsPostgreSql)
            {
                var (insertSql, paras) = SqlInsertDict(data);
                var conflictFields = string.Join(", ", keys.Select(k => MetaData.SqlQuote(k)));
                var updateFields = data.Keys
                    .Where(k => !keys.Contains(k, StringComparer.OrdinalIgnoreCase) && k != "Id" && k != "Guid")
                    .Select(k => $"{MetaData.SqlQuote(k)} = EXCLUDED.{MetaData.SqlQuote(k)}");

                var upsertSql = $"{insertSql} ON CONFLICT ({conflictFields}) DO UPDATE SET {string.Join(", ", updateFields)}";
                var result = await ExecAsync(upsertSql, trans, paras);
                if (result > 0)
                {
                    var id1 = data[keys[0]];
                    var id2 = keys.Length > 1 ? data[keys[1]] : null;
                    await InvalidateCacheAsync(id1!, id2);
                    foreach (var kv in data)
                    {
                        await InvalidateFieldCacheAsync(kv.Key, id1!, id2);
                    }
                }
                return result;
            }
            else
            {
                var parameters = new List<IDataParameter>();
                var sourceSelects = new List<string>();
                int i = 0;

                foreach (var kv in data)
                {
                    i++;
                    string paramName = $"@p{i}";
                    parameters.Add(MetaData.CreateParameter(paramName, kv.Value));
                    sourceSelects.Add($"{paramName} AS {MetaData.SqlQuote(kv.Key)}");
                }

                var onClause = string.Join(" AND ", keys.Select(k => $"Target.{MetaData.SqlQuote(k)} = Source.{MetaData.SqlQuote(k)}"));
                var updateFields = data.Keys
                    .Where(k => !keys.Contains(k, StringComparer.OrdinalIgnoreCase) && k != "Id" && k != "Guid")
                    .Select(k => $"{MetaData.SqlQuote(k)} = Source.{MetaData.SqlQuote(k)}");

                var insertFields = data.Keys.Select(k => MetaData.SqlQuote(k));
                var insertValues = data.Keys.Select(k => $"Source.{MetaData.SqlQuote(k)}");

                var mergeSql = $@"
MERGE INTO {FullName} AS Target
USING (SELECT {string.Join(", ", sourceSelects)}) AS Source
ON {onClause}
WHEN MATCHED THEN
    UPDATE SET {string.Join(", ", updateFields)}
WHEN NOT MATCHED THEN
    INSERT ({string.Join(", ", insertFields)}) VALUES ({string.Join(", ", insertValues)});";

                var result = await ExecAsync(mergeSql, trans, [.. parameters]);
                if (result > 0)
                {
                    var id1 = data[keys[0]];
                    var id2 = keys.Length > 1 ? data[keys[1]] : null;
                    await InvalidateCacheAsync(id1!, id2);
                    foreach (var kv in data)
                    {
                        await InvalidateFieldCacheAsync(kv.Key, id1!, id2);
                    }
                }
                return result;
            }
        }

        public static int Upsert(object columns, IDbTransaction? trans = null)
        {
            var data = MetaData.ToCovList(columns).ToDictionary(c => c.Name, c => c.Value);
            var keys = Keys.Where(k => !string.IsNullOrEmpty(k)).ToArray();

            if (keys.Length == 0) throw new InvalidOperationException("Upsert 操作必须定义主键字段");
            foreach (var key in keys)
            {
                if (!data.ContainsKey(key)) throw new ArgumentException($"Upsert 操作的数据中必须包含主键字段 [{key}]");
            }

            if (IsPostgreSql)
            {
                var (insertSql, paras) = SqlInsertDict(data);
                var conflictFields = string.Join(", ", keys.Select(k => MetaData.SqlQuote(k)));
                var updateFields = data.Keys
                    .Where(k => !keys.Contains(k, StringComparer.OrdinalIgnoreCase) && k != "Id" && k != "Guid")
                    .Select(k => $"{MetaData.SqlQuote(k)} = EXCLUDED.{MetaData.SqlQuote(k)}");

                var upsertSql = $"{insertSql} ON CONFLICT ({conflictFields}) DO UPDATE SET {string.Join(", ", updateFields)}";
                var result = Exec(upsertSql, trans, paras);
                if (result > 0)
                {
                    var id1 = data[keys[0]];
                    var id2 = keys.Length > 1 ? data[keys[1]] : null;
                    InvalidateCache(id1!, id2);
                    foreach (var kv in data)
                    {
                        InvalidateFieldCache(kv.Key, id1!, id2);
                    }
                }
                return result;
            }
            else
            {
                var parameters = new List<IDataParameter>();
                var sourceSelects = new List<string>();
                int i = 0;

                foreach (var kv in data)
                {
                    i++;
                    string paramName = $"@p{i}";
                    parameters.Add(MetaData.CreateParameter(paramName, kv.Value));
                    sourceSelects.Add($"{paramName} AS {MetaData.SqlQuote(kv.Key)}");
                }

                var onClause = string.Join(" AND ", keys.Select(k => $"Target.{MetaData.SqlQuote(k)} = Source.{MetaData.SqlQuote(k)}"));
                var updateFields = data.Keys
                    .Where(k => !keys.Contains(k, StringComparer.OrdinalIgnoreCase) && k != "Id" && k != "Guid")
                    .Select(k => $"{MetaData.SqlQuote(k)} = Source.{MetaData.SqlQuote(k)}");

                var insertFields = data.Keys.Select(k => MetaData.SqlQuote(k));
                var insertValues = data.Keys.Select(k => $"Source.{MetaData.SqlQuote(k)}");

                var mergeSql = $@"
MERGE INTO {FullName} AS Target
USING (SELECT {string.Join(", ", sourceSelects)}) AS Source
ON {onClause}
WHEN MATCHED THEN
    UPDATE SET {string.Join(", ", updateFields)}
WHEN NOT MATCHED THEN
    INSERT ({string.Join(", ", insertFields)}) VALUES ({string.Join(", ", insertValues)});";

                var result = Exec(mergeSql, trans, [.. parameters]);
                if (result > 0)
                {
                    var id1 = data[keys[0]];
                    var id2 = keys.Length > 1 ? data[keys[1]] : null;
                    InvalidateCache(id1!, id2);
                    foreach (var kv in data)
                    {
                        InvalidateFieldCache(kv.Key, id1!, id2);
                    }
                }
                return result;
            }
        }

        // ----------- Cache 缓存管理 -----------

        protected static string GetCacheKey(params object[] keys)
            => $"MetaData:{FullName}:Id:{string.Join("_", keys)}";

        public static async Task<TDerived?> GetByKeyAsync(object key1, object? key2 = null, IDbTransaction? trans = null)
        {
            var cacheKey = key2 == null ? GetCacheKey(key1) : GetCacheKey(key1, key2);

            if (trans == null && MetaData.CacheService != null)
            {
                var cached = await MetaData.CacheService.GetAsync<TDerived>(cacheKey);
                if (cached != null) 
                {
                    // Logger.Debug($"[Cache] HIT - Table: {FullName}, Key: {cacheKey}");
                    return cached;
                }
                // Logger.Debug($"[Cache] MISS - Table: {FullName}, Key: {cacheKey}");
            }

            var (sql, parameters) = SqlSelect("*", key1, key2);
            // Logger.Debug($"[DB] Loading Row - Table: {FullName}, SQL: {sql}");
            var result = await QuerySingleAsync<TDerived>(sql, trans, parameters);

            if (trans == null && result != null && MetaData.CacheService != null)
            {
                // Logger.Debug($"[Cache] SET - Table: {FullName}, Key: {cacheKey}");
                await MetaData.CacheService.SetAsync(cacheKey, result, TimeSpan.FromMinutes(1));
            }

            return result;
        }

        public static TDerived? GetByKey(object key1, object? key2 = null, IDbTransaction? trans = null)
        {
            var cacheKey = key2 == null ? GetCacheKey(key1) : GetCacheKey(key1, key2);

            if (trans == null && MetaData.CacheService != null)
            {
                var cached = MetaData.CacheService.Get<TDerived>(cacheKey);
                if (cached != null) return cached;
            }

            var (sql, parameters) = SqlSelect("*", key1, key2);
            var result = QuerySingle<TDerived>(sql, trans, parameters);

            if (trans == null && result != null && MetaData.CacheService != null)
            {
                MetaData.CacheService.Set(cacheKey, result, TimeSpan.FromMinutes(1));
            }

            return result;
        }

        public static T GetCached<T>(string fieldName, object id, object? id2 = null, T? defaultValue = default, IDbTransaction? trans = null)
        {
            T LoadFromDb()
            {
                var (sql, paras) = SqlGet(fieldName, id, id2);
                // Logger.Debug($"[DB] Loading Field - Table: {FullName}, Field: {fieldName}, SQL: {sql}");
                return QueryScalar<T>(sql, trans, paras)!;               
            }

            if (trans != null || MetaData.CacheService == null || !MetaData.UseCache) return LoadFromDb();
            var cacheKey = id2 == null ? GetCacheKey(fieldName, id) : GetCacheKey(fieldName, id, id2);
            
            var result = MetaData.CacheService.GetOrAdd(cacheKey, LoadFromDb, TimeSpan.FromMinutes(1));
            // Logger.Debug($"[Cache] GetOrAdd - Table: {FullName}, Field: {fieldName}, Key: {cacheKey}");
            return result;
        }

        public static async Task<T> GetFieldAsync<T>(string fieldName, object id, object? id2 = null, T defaultValue = default!, IDbTransaction? trans = null) where T : struct
        {
            async Task<T> LoadFromDbAsync()
            {
                var (sql, paras) = SqlGet(fieldName, id, id2);
                var raw = await QueryScalarAsync<T>(sql, trans, paras);
                var val = SqlHelper.ConvertValue<T>(raw, defaultValue);
                // Logger.Debug($"[DB] LoadField - Table: {FullName}, Field: {fieldName}, ID: {id}-{id2}, Result: {val}");
                return val;
            }

            if (trans != null || MetaData.CacheService == null || !MetaData.UseCache || 
                fieldName == "IsGroup" || fieldName == "IsPrivate" || 
                fieldName == "IsPowerOn" || fieldName == "UseRight" || fieldName == "IsOpen") 
            {
                // Logger.Debug($"[Cache] Bypass - Table: {FullName}, Field: {fieldName}, Reason: {(trans != null ? "Transaction" : (MetaData.CacheService == null ? "NoCacheService" : (!MetaData.UseCache ? "Disabled" : "CriticalField")))} (at MetaData.cs:GetAsync, line 2232)");
                return await LoadFromDbAsync();
            }
            var cacheKey = id2 == null ? GetCacheKey(fieldName, id) : GetCacheKey(fieldName, id, id2);
            var result = await MetaData.CacheService.GetOrAddAsync(cacheKey, LoadFromDbAsync, TimeSpan.FromMinutes(1));
            // Logger.Debug($"[Cache] GetOrAddField - Table: {FullName}, Field: {fieldName}, Key: {cacheKey}, Value: {result}");
            return result;
        }

        public static (string, IDataParameter[]) SqlGetForUpdate(string fieldName, object id, object? id2 = null)
        {
            var (where, parameters) = SqlWhere(id, id2);
            string sql = IsPostgreSql
                ? $"SELECT {Quote(fieldName)} FROM {FullName} {where} {SqlForUpdate}"
                : $"SELECT {Quote(fieldName)} FROM {FullName} {SqlLockHint} {where}";
            return (sql, parameters);
        }

        public static async Task<T> GetForUpdateAsync<T>(string fieldName, object id, object? id2 = null, T defaultValue = default!, IDbTransaction? trans = null) where T : struct
        {
            var (sql, paras) = SqlGetForUpdate(fieldName, id, id2);
            // Logger.Debug($"[DB] GetForUpdate - Table: {FullName}, Field: {fieldName}, SQL: {sql}");
            var raw = await QueryScalarAsync<T>(sql, trans, paras);
            return SqlHelper.ConvertValue<T>(raw, defaultValue);
        }

        public static T GetForUpdate<T>(string fieldName, object id, object? id2 = null, T defaultValue = default!, IDbTransaction? trans = null) where T : struct
        {
            var (sql, paras) = SqlGetForUpdate(fieldName, id, id2);
            var raw = QueryScalar<T>(sql, trans, paras);
            return SqlHelper.ConvertValue<T>(raw, defaultValue);
        }

        public static async Task<string> GetValueAsync(string fieldName, object id, object? id2 = null, IDbTransaction? trans = null)
        {
            async Task<string> LoadFromDbAsync()
            {
                var (sql, parameters) = SqlGetStr(fieldName, id, id2);
                var res = await QueryScalarAsync<string>(sql, trans, parameters);
                var val = res ?? "";
                // Logger.Debug($"[DB] LoadValue - Table: {FullName}, Field: {fieldName}, ID: {id}-{id2}, Result: {val}");
                return val;
            }

            if (trans != null || MetaData.CacheService == null || !MetaData.UseCache) return await LoadFromDbAsync();
            var cacheKey = id2 == null ? GetCacheKey(fieldName, id) : GetCacheKey(fieldName, id, id2);
            var result = await MetaData.CacheService.GetOrAddAsync(cacheKey, LoadFromDbAsync, TimeSpan.FromMinutes(1));
            // Logger.Debug($"[Cache] GetOrAddValue - Table: {FullName}, Field: {fieldName}, Key: {cacheKey}, Value: {result}");
            return result;
        }

        public static async Task InvalidateCacheAsync(object key1, object? key2 = null)
        {
            if (MetaData.CacheService == null || !MetaData.UseCache) return;
            var cacheKey = key2 == null ? GetCacheKey(key1) : GetCacheKey(key1, key2);
            // Logger.Debug($"[Cache] InvalidateAsync - Table: {FullName}, Key: {cacheKey}");
            await MetaData.CacheService.RemoveAsync(cacheKey);
        }

        public static void InvalidateCache(object key1, object? key2 = null)
        {
            if (MetaData.CacheService == null || !MetaData.UseCache) return;
            var cacheKey = key2 == null ? GetCacheKey(key1) : GetCacheKey(key1, key2);
            // Logger.Debug($"[Cache] Invalidate - Table: {FullName}, Key: {cacheKey}");
            MetaData.CacheService.Remove(cacheKey);
        }

        /// <summary>
        /// 失效指定 ID 的行级缓存和所有列级缓存
        /// </summary>
        public static async Task InvalidateAllCachesAsync(object id, object? id2 = null)
        {
            if (MetaData.CacheService == null || !MetaData.UseCache) return;
            // Logger.Debug($"[Cache] InvalidateAllAsync - Table: {FullName}, ID: {id}-{id2}");
            await InvalidateCacheAsync(id, id2);
            var props = GetProperties();
            foreach (var prop in props)
            {
                await InvalidateFieldCacheAsync(prop.Name, id, id2);
            }
        }

        public static void InvalidateAllCaches(object id, object? id2 = null)
        {
            if (MetaData.CacheService == null || !MetaData.UseCache) return;
            // Logger.Debug($"[Cache] InvalidateAll - Table: {FullName}, ID: {id}-{id2}");
            InvalidateCache(id, id2);
            var props = GetProperties();
            foreach (var prop in props)
            {
                InvalidateFieldCache(prop.Name, id, id2);
            }
        }

        public static void InvalidateFieldCache(string fieldName, object key1, object? key2 = null)
        {
            if (MetaData.CacheService == null || !MetaData.UseCache) return;
            var cacheKey = key2 == null ? GetCacheKey(fieldName, key1) : GetCacheKey(fieldName, key1, key2);
            // Logger.Debug($"[Cache] InvalidateField - Table: {FullName}, Field: {fieldName}, Key: {cacheKey}");
            MetaData.CacheService.Remove(cacheKey);
        }

        public static async Task InvalidateFieldCacheAsync(string fieldName, object key1, object? key2 = null)
        {
            if (MetaData.CacheService == null || !MetaData.UseCache) return;
            var cacheKey = key2 == null ? GetCacheKey(fieldName, key1) : GetCacheKey(fieldName, key1, key2);
            // Logger.Debug($"[Cache] InvalidateFieldAsync - Table: {FullName}, Field: {fieldName}, Key: {cacheKey}");
            await MetaData.CacheService.RemoveAsync(cacheKey);
        }

        public async Task SaveAsync(ICacheService cacheService)
        {
            var keys = GetKeyValues().Select(k => k.Value).ToArray();
            var cacheKey = GetCacheKey(keys);
            await cacheService.SetAsync(cacheKey, this, TimeSpan.FromMinutes(1));
        }

        // ----------- 其他辅助方法 -----------

        public static async Task<MetaData.TransactionWrapper> BeginTransactionAsync(IDbTransaction? existingTrans = null, IsolationLevel level = IsolationLevel.ReadCommitted)
            => await MetaData.BeginTransactionAsync(existingTrans, level);
        public static MetaData.TransactionWrapper BeginTransaction() => MetaData.BeginTransaction();
        public static void CommitTransaction(IDbTransaction trans) => MetaData.CommitTransaction(trans);
        public static void RollbackTransaction(IDbTransaction trans) => MetaData.RollbackTransaction(trans);
        public static void ClearTransaction(IDbTransaction? trans = null) => MetaData.ClearTransaction(trans);

        public static async Task EnsureTableCreatedAsync()
        {
            if (_isTableChecked) return;
            try
            {
                var tableName = _instance.TableName;
                string sqlCheck = IsPostgreSql 
                    ? $"SELECT COUNT(*) FROM pg_tables WHERE schemaname = 'public' AND tablename = '{tableName.ToLower()}'"
                    : $"SELECT COUNT(*) FROM sys.tables WHERE name = '{tableName}'";
                
                var count = await QueryScalarAsync<int>(sqlCheck);
                if (count == 0)
                {
                    var sqlCreate = BotWorker.Infrastructure.Utils.Schema.SchemaSynchronizer.GenerateCreateTableSql<TDerived>();
                    var result = await SQLConn.ExecAsync(sqlCreate, isDebug: true);
                }
                _isTableChecked = true;
            }
            catch (Exception ex)
            {
                Console.WriteLine($"[ORM] Error ensuring table {FullName} exists: {ex.Message}");
            }
        }

        public List<(string Name, object Value)> GetKeyValues()
        {
            var list = new List<(string, object)>();
            var props = GetProperties();
            foreach (var key in Keys)
            {
                if (string.IsNullOrEmpty(key)) continue;
                var prop = props.FirstOrDefault(p => string.Equals(p.Name, key, StringComparison.OrdinalIgnoreCase)) 
                    ?? throw new Exception($"主键属性 {key} 不存在或被忽略");
                list.Add((key, prop.GetValue(this) ?? DBNull.Value));
            }
            return list;
        }

        protected virtual Dictionary<string, object> GetInsertFields()
        {
            return PropertyHelper.GetAll(GetType()).Where(p => p.IncludeInInsert).ToDictionary(p => p.ColumnName, p => p.GetValue(this) ?? DBNull.Value);
        }

        protected virtual Dictionary<string, object> GetUpdateFields()
        {
            return PropertyHelper.GetAll(GetType()).Where(p => p.IncludeInUpdate).ToDictionary(p => p.ColumnName, p => p.GetValue(this) ?? DBNull.Value);
        }

        public static string GetSqlValue(object value, string parameterName)
        {
            if (value is DateTime dateTimeValue && dateTimeValue == DateTime.MinValue) return $"{SqlIsNull(parameterName, SqlDateTime)}";
            else if (value is Guid guidValue && guidValue == Guid.Empty) return $"{SqlIsNull(parameterName, SqlRandomOrder)}";
            else return parameterName;
        }

        public static IDataParameter CreateParameter(string parameterName, object? value) => MetaData.CreateParameter(parameterName, value);

        public static string FormatValue(object value)
        {
            if (value is null) return "NULL";
            else if (value is string)
            {
                string str = EscapeSqlString(value.AsString());
                return IsPostgreSql ? $"'{str}'" : $"N'{str}'";
            }
            else if (value is DateTime dateTime)
            {
                if (dateTime == DateTime.MinValue) return SqlDateTime;
                else return $"'{dateTime:yyyy-MM-dd HH:mm:ss}'";
            }
            else return value.AsString();
        }

        private static string EscapeSqlString(string value) => value.Replace("'", "''");

        private static PropertyInfo[] GetProperties()
        {
            return _propertyCache.GetOrAdd(typeof(TDerived), type =>
            {
                return type.GetProperties()
                    .Where(prop =>
                    {
                        if (prop.GetIndexParameters().Length > 0) return false;
                        if (prop.GetCustomAttribute<DbIgnoreAttribute>() != null) return false;
                        if (!prop.CanWrite) return false;
                        if (!prop.CanRead || !prop.GetMethod!.IsPublic || !prop.SetMethod!.IsPublic) return false;
                        if (prop.GetMethod!.IsStatic) return false;
                        if (prop.Name.StartsWith("_") || prop.Name.StartsWith("$")) return false;
                        return true;
                    }).ToArray();
            });
        }

        public Dictionary<string, object?> ToDictionary()
        {
            var dict = new Dictionary<string, object?>();
            var props = GetProperties();
            foreach (var prop in props) dict[prop.Name] = prop.GetValue(this);
            return dict;
        }

        public static async Task<TDerived> LoadAsync(object id, object? id2 = null)
            => await GetSingleAsync(id, id2) ?? throw new Exception($"主键属性 {id} {id2}不存在");

        public static void SyncCacheField(long qq, long groupId, string field, object value)
        {
            if (MetaData.CacheService == null || !MetaData.UseCache) return;
            var id2 = string.IsNullOrEmpty(Key2) ? null : (groupId == 0 ? null : (object)groupId);
            var cacheKey = id2 == null ? GetCacheKey(field, qq) : GetCacheKey(field, qq, id2);
            MetaData.CacheService.Set(cacheKey, value, TimeSpan.FromMinutes(1));
            // 同时失效行级缓存，因为行数据已经变了
            InvalidateCache(qq, id2);
            // Logger.Debug($"[CacheSync] {FullName}: {qq}-{groupId} {field} = {value}");
        }

        public static async Task SyncCacheFieldAsync(long qq, long groupId, string field, object value)
        {
            if (MetaData.CacheService == null || !MetaData.UseCache) return;
            var id2 = string.IsNullOrEmpty(Key2) ? null : (groupId == 0 ? null : (object)groupId);
            var cacheKey = id2 == null ? GetCacheKey(field, qq) : GetCacheKey(field, qq, id2);
            await MetaData.CacheService.SetAsync(cacheKey, value, TimeSpan.FromMinutes(1));
            // 同时失效行级缓存，因为行数据已经变了
            await InvalidateCacheAsync(qq, id2);
            // Logger.Debug($"[CacheSyncAsync] {FullName}: {qq}-{groupId} {field} = {value}");
        }

        public static void SyncCacheField(long qq, string field, object value)
            => SyncCacheField(qq, 0, field, value);

        public static async Task SyncCacheFieldAsync(long qq, string field, object value)
            => await SyncCacheFieldAsync(qq, 0, field, value);

        public static async Task<TDerived?> LoadAsync(Guid guid)
            => await GetSingleAsync(guid);
    }

    public abstract class MetaDataGuid<TDerived> : MetaData<TDerived> where TDerived : MetaDataGuid<TDerived>, new()
    {
        public long Id { get; set; }
        public Guid Guid { get; set; } = default;

        protected virtual string GuidFieldName => "Guid";
        protected virtual string IdFieldName => "Id";

        public static string GuidField { get; private set; }
        public static string IdField { get; private set; }

        static MetaDataGuid()
        {
            var instance = new TDerived();            
            GuidField = instance.GuidFieldName;
            IdField = instance.IdFieldName;
        }

        public static new async Task<TDerived?> LoadAsync(Guid guid)
        {
            if (Key.Equals(GuidField, StringComparison.OrdinalIgnoreCase))
                return await GetSingleAsync(guid);
            else
                return await GetSingleAsync(GetId(guid));
        }

        public static async Task<TDerived?> LoadAsync(long Id)
        {
            if (Key.Equals(IdField, StringComparison.OrdinalIgnoreCase))            
                return await GetSingleAsync(Id);            
            else
                return await GetSingleAsync(GetGuid(Id));
        }

        public static long GetId(Guid guid)
        {  
            if (Key.Equals(GuidField, StringComparison.OrdinalIgnoreCase))            
                return GetLong(IdField, guid);            
            else            
                return GetWhere<long>(IdField, $"{Quote(GuidField)} = @guid", SqlParams(("@guid", guid)));            
        }

        public static long GetId(string guid) => GetId(Guid.Parse(guid));

        public static async Task<TDerived?> GetByGuidAsync(string guid)
        {
            return (await QueryWhere($"{Quote(GuidField)} = @guid", SqlParams(("@guid", guid)))).FirstOrDefault();
        }

        public static async Task<TDerived?> GetByGuidAsync(Guid guid)
        {
            return (await QueryWhere($"{Quote(GuidField)} = @guid", SqlParams(("@guid", guid)))).FirstOrDefault();
        }

        public static Guid GetGuid(long id)
        {
            if (Key.Equals(IdField, StringComparison.OrdinalIgnoreCase))            
                return Get<Guid>(GuidField, id);            
            else            
                return GetWhere<Guid>(GuidField, $"{Quote(IdField)} = @id", SqlParams(("@id", id)));            
        }

        public static async Task<Guid> GetGuidAsync(long id, IDbTransaction? trans = null)
        {
            if (Key.Equals(IdField, StringComparison.OrdinalIgnoreCase))
                return await GetAsync<Guid>(GuidField, id, null, default, trans);
            else
                return await GetWhereAsync<Guid>(GuidField, $"{Quote(IdField)} = @id", "", trans);
        }

        public static Dictionary<string, object?>? GetDict(Guid guid, params string[] fields)
            => GetDicts($"{Quote(GuidField)} = @guid", SqlParams(("@guid", guid)) , fields).FirstOrDefault();

        public static Dictionary<string, object?>? GetDict(long id, params string[] fields)
            => GetDicts($"{Quote(IdField)} = @id", SqlParams(("@id", id)), fields).FirstOrDefault();
    }
}
