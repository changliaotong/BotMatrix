using System.Data;
using System.Data.Common;
using System.Text.RegularExpressions;
using BotWorker.Infrastructure.Extensions;

namespace BotWorker.Infrastructure.Persistence
{
    public static class SqlHelper
    {
        public static ICacheService? CacheService { get; set; }
        public static bool UseCache { get; set; } = false;
        public static bool IsPostgreSql => true;

        public static string SqlLimit(long n) => $" LIMIT {n}";

        public static string SqlTop(long n) => IsPostgreSql ? string.Empty : $" TOP {n} ";

        public static string SqlIsNull(string field, string def) => $"COALESCE({field}, {def})";

        public static string SqlRandomOrder => "RANDOM()";

        public static string SqlRandomId => "gen_random_uuid()";

        public static string SqlDateTime => "CURRENT_TIMESTAMP";

        public static string SqlLen(string field) => $"LENGTH({field})";

        public static string SqlForUpdate => " FOR UPDATE";

        public static string SqlDateAdd(string unit, object number, string start)
        {
            return unit.ToUpper() switch
            {
                "YEAR" or "YY" or "YYYY" => $"({start} + interval '{number} year')",
                "MONTH" or "MM" or "M" => $"({start} + interval '{number} month')",
                "DAY" or "DD" or "D" => $"({start} + interval '{number} day')",
                "HOUR" or "HH" => $"({start} + interval '{number} hour')",
                "MINUTE" or "MI" or "N" => $"({start} + interval '{number} minute')",
                "SECOND" or "SS" or "S" => $"({start} + interval '{number} second')",
                _ => $"({start} + interval '{number} {unit}')"
            };
        }

        public static string SqlDateDiff(string unit, string start, string end)
        {
            return unit.ToUpper() switch
            {
                "YEAR" or "YY" or "YYYY" => $"EXTRACT(YEAR FROM AGE({end}, {start}))",
                "MONTH" or "MM" or "M" => $"(EXTRACT(YEAR FROM AGE({end}, {start})) * 12 + EXTRACT(MONTH FROM AGE({end}, {start})))",
                "DAY" or "DD" or "D" => $"EXTRACT(DAY FROM ({end} - {start}))",
                "HOUR" or "HH" => $"(EXTRACT(EPOCH FROM ({end} - {start})) / 3600)",
                "MINUTE" or "MI" or "N" => $"(EXTRACT(EPOCH FROM ({end} - {start})) / 60)",
                "SECOND" or "SS" or "S" => $"EXTRACT(EPOCH FROM ({end} - {start}))",
                _ => $"EXTRACT(EPOCH FROM ({end} - {start}))"
            };
        }

        public static string SqlQuote(string identifier)
        {
            if (string.IsNullOrWhiteSpace(identifier)) return identifier;
            if (identifier.Contains('.') || identifier.Contains('"') || identifier.Contains('[') || identifier.Contains(' '))
                return identifier;

            return $"\"{identifier.ToLower()}\"";
        }

        public static object? ConvertValue(object? value, Type targetType)
        {
            if (value == null || value == DBNull.Value)
            {
                if (targetType.IsValueType && Nullable.GetUnderlyingType(targetType) == null)
                    return Activator.CreateInstance(targetType);
                return null;
            }

            try
            {
                if (targetType == typeof(Guid))
                {
                    if (value is Guid g) return g;
                    if (value is byte[] bytes && bytes.Length == 16) return new Guid(bytes);
                    if (Guid.TryParse(value.ToString(), out var guid)) return guid;
                    return Guid.Empty;
                }

                if (targetType.IsEnum)
                {
                    return Enum.Parse(targetType, value.ToString()!, ignoreCase: true);
                }

                if (targetType == typeof(bool))
                {
                    if (value is int i) return i != 0;
                    if (value is long l) return l != 0;
                    if (value is string s) return s == "1" || s.Equals("true", StringComparison.OrdinalIgnoreCase);
                    if (value is bool b) return b;
                }

                if (targetType == typeof(int) && value is long l2) return (int)l2;
                if (targetType == typeof(long) && value is int i2) return (long)i2;

                return Convert.ChangeType(value, targetType);
            }
            catch
            {
                if (targetType.IsValueType && Nullable.GetUnderlyingType(targetType) == null)
                    return Activator.CreateInstance(targetType);
                return null;
            }
        }

        public static T ConvertValue<T>(object? value, T? defaultValue = default)
        {
            if (value == null || value == DBNull.Value)
                return defaultValue!;

            var targetType = typeof(T);

            try
            {
                // Guid 特殊处理
                if (targetType == typeof(Guid))
                {
                    if (value is Guid g)
                        return (T)(object)g;

                    if (value is byte[] bytes && bytes.Length == 16)
                        return (T)(object)new Guid(bytes);

                    if (Guid.TryParse(value.ToString(), out var guid))
                        return (T)(object)guid;

                    return defaultValue!;
                }

                // Enum 特殊处理
                if (targetType.IsEnum)
                {
                    return (T)Enum.Parse(targetType, value.ToString()!, ignoreCase: true);
                }

                // bool 特殊处理
                if (targetType == typeof(bool))
                {
                    if (value is int i)
                        return (T)(object)(i != 0);
                    if (value is long l)
                        return (T)(object)(l != 0);
                    if (value is byte b1)
                        return (T)(object)(b1 != 0);
                    if (value is short s1)
                        return (T)(object)(s1 != 0);
                    if (value is decimal d)
                        return (T)(object)(d != 0);
                    if (value is string s)
                        return (T)(object)(s == "1" || s.Equals("true", StringComparison.OrdinalIgnoreCase));
                    if (value is bool b)
                        return (T)(object)b;
                }

                // 其它类型正常转换
                return (T)Convert.ChangeType(value, targetType);
            }
            catch
            {
                return defaultValue!;
            }
        }

        public static string GetColumnName(System.Reflection.PropertyInfo prop)
        {
            // 先尝试获取自定义的 ColumnAttribute 特性
            var attr = System.Reflection.CustomAttributeExtensions.GetCustomAttribute<ColumnAttribute>(prop);
            if (attr != null && !string.IsNullOrEmpty(attr.Name))
            {
                return attr.Name;
            }
            // 没有特性或没指定名称就用属性名
            return prop.Name;
        }

        public static object? ConvertFromDbValue(object? dbValue, System.Reflection.PropertyInfo prop)
        {
            var attr = System.Reflection.CustomAttributeExtensions.GetCustomAttribute<ColumnAttribute>(prop);
            if (attr?.ConverterType != null)
            {
                if (Activator.CreateInstance((Type)attr.ConverterType) is IValueConverter converter)
                {
                    return converter.ConvertFromProvider(dbValue);
                }
                else
                {
                    throw new InvalidOperationException($"ConverterType {attr.ConverterType} must implement IValueConverter");
                }
            }

            return ConvertValue(dbValue, prop.PropertyType);
        }

        public static (string Sql, IDataParameter[] Parameters) ResolveSql(string sql, params object?[] args)
        {
            return ResolveSql(sql, null, args);
        }

        public static (string Sql, IDataParameter[] Parameters) ResolveSql(string sql, IEnumerable<string>? columns, params object?[] args)
        {
            if (string.IsNullOrWhiteSpace(sql)) return (sql, []);

            if (columns != null)
            {
                foreach (var col in columns)
                {
                    if (string.IsNullOrEmpty(col)) continue;
                    var pattern = @"(?<![\[""])\b" + Regex.Escape(col) + @"\b(?![\]""])";
                    sql = Regex.Replace(sql, pattern, "[" + col + "]");
                }
            }

            if (IsPostgreSql)
            {
                sql = Regex.Replace(sql, @"\[([^\]]+)\]", m => "\"" + m.Groups[1].Value.ToLower() + "\"");
            }

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
                        parameters.Add(SQLConn.CreateParameter(paramName, args[i]));
                    }
                }
            }

            return (sql, parameters.ToArray());
        }

        public static IDbTransaction? Unwrap(IDbTransaction? trans)
        {
            if (trans is TransactionWrapper wrapper) return wrapper.Transaction;
            return trans;
        }

        public static TransactionWrapper BeginTransaction(IDbTransaction? existingTrans = null, IsolationLevel level = IsolationLevel.ReadCommitted)
        {
            if (existingTrans != null)
            {
                var trans = existingTrans;
                if (trans is TransactionWrapper wrapper)
                    trans = wrapper.Transaction;
                return new TransactionWrapper(trans, false);
            }
            var conn = SQLConn.GetConnection();
            if (conn.State != ConnectionState.Open) conn.Open();
            return new TransactionWrapper(conn.BeginTransaction(level), true);
        }

        public static async Task<TransactionWrapper> BeginTransactionAsync(IDbTransaction? existingTrans = null, IsolationLevel level = IsolationLevel.ReadCommitted)
        {
            if (existingTrans != null)
            {
                var trans = existingTrans;
                if (trans is TransactionWrapper wrapper)
                    trans = wrapper.Transaction;
                return new TransactionWrapper(trans, false);
            }
            var conn = SQLConn.GetConnection();
            if (conn.State != ConnectionState.Open) await ((DbConnection)conn).OpenAsync();
            return new TransactionWrapper(await ((DbConnection)conn).BeginTransactionAsync(level), true);
        }

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
                        try { _inner.Rollback(); } catch { }
                    }

                    var conn = _inner.Connection;
                    _inner.Dispose();
                    if (_ownConnection && conn != null)
                    {
                        conn.Close();
                        conn.Dispose();
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
                        try
                        {
                            if (_inner is DbTransaction dbTrans) await dbTrans.RollbackAsync();
                            else _inner.Rollback();
                        }
                        catch { }
                    }

                    var conn = _inner.Connection;
                    if (_inner is DbTransaction dbTrans2) await dbTrans2.DisposeAsync();
                    else _inner.Dispose();

                    if (_ownConnection && conn != null)
                    {
                        if (conn is DbConnection dbConn)
                        {
                            await dbConn.CloseAsync();
                            await dbConn.DisposeAsync();
                        }
                        else
                        {
                            conn.Close();
                            conn.Dispose();
                        }
                    }
                    _disposed = true;
                }
            }
        }
    }
}
