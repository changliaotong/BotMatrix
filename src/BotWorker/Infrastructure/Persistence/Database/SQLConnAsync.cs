using System.Data;
using System.Data.Common;
using Newtonsoft.Json;
using BotWorker.Infrastructure.Persistence;

namespace BotWorker.Infrastructure.Persistence.Database
{
    public static partial class SQLConn
    {
        public static async Task<Dictionary<string, object>> ExecWithOutputAsync(string sql, IDataParameter[] parameters, string[] outputFields, IDbTransaction? trans = null)
        {
            IDbConnection? conn = trans?.Connection;
            
            if (trans != null && conn == null) trans = null; // 事务已失效

            bool isNewConn = false;

            if (conn == null)
            {
                conn = DbProviderFactory.CreateConnection();
                // Console.WriteLine($"[DB INFO] Opening async connection to: {GlobalConfig.DbType}");
                if (conn is DbConnection dbConn) await dbConn.OpenAsync(); else conn.Open();
                // Console.WriteLine("[DB INFO] Async connection opened successfully.");
                isNewConn = true;
            }

            try
            {
                LogSql(sql, parameters);
                using var cmd = conn.CreateCommand();
                cmd.CommandText = sql;
                cmd.Transaction = Unwrap(trans);
                cmd.CommandTimeout = 60; // 增加超时时间到60秒
                
                var processedParameters = ProcessParameters(parameters);
                if (processedParameters != null)
                {
                    foreach (var p in processedParameters) cmd.Parameters.Add(p);
                }

                using var reader = await (cmd as DbCommand)?.ExecuteReaderAsync()!;

                var result = new Dictionary<string, object>(StringComparer.OrdinalIgnoreCase);

                if (await reader.ReadAsync())
                {
                    foreach (var field in outputFields)
                    {
                        var value = reader[field];
                        result[field] = value == DBNull.Value ? null! : value;
                    }
                }

                return result;
            }
            finally
            {
                if (isNewConn && conn != null)
                {
                    conn.Dispose();
                }
            }
        }

        private static readonly HashSet<Type> SimpleTypes = new()
        {
            typeof(string),
            typeof(decimal),
            typeof(DateTime),
            typeof(Guid),
            typeof(TimeSpan),
            typeof(DateOnly),
            typeof(TimeOnly)
        };

        private static bool IsSimpleType(Type type)
        {
            return type.IsPrimitive || SimpleTypes.Contains(type) || type.IsEnum;
        }

        public static async Task<T?> QueryScalarAsync<T>(string sql, params IDataParameter[] parameters)
        {
            return await QueryScalarAsync<T>(sql, true, null, parameters);
        }

        public static async Task<T?> QueryScalarAsync<T>(string sql, bool isDebug = true, IDbTransaction? trans = null, params IDataParameter[] parameters)
        {
            IDbConnection? conn = trans?.Connection;

            if (trans != null && conn == null) trans = null;

            bool isNewConn = false;

            if (conn == null)
            {
                conn = DbProviderFactory.CreateConnection();
                if (conn is DbConnection dbConn) await dbConn.OpenAsync(); else conn.Open();
                isNewConn = true;
            }

            try
            {
                LogSql(sql, parameters);
                using var cmd = conn.CreateCommand();
                cmd.CommandText = sql;
                cmd.Transaction = Unwrap(trans);
                cmd.CommandTimeout = 60;
                
                var processedParameters = ProcessParameters(parameters);
                if (processedParameters != null)
                {
                    foreach (var p in processedParameters) cmd.Parameters.Add(p);
                }

                object? result = await (cmd as DbCommand)?.ExecuteScalarAsync()!;

                if (result == null || result == DBNull.Value)
                    return default;

                result = SqlHelper.ConvertValue<T>(result, default!);

                return result != null ? (T)Convert.ChangeType(result, typeof(T)) : default!;
            }
            catch (Exception ex)
            {
                var errorMsg = $"[DB ERROR] Async operation failed: {ex.Message}\nSQL: {sql}";
                Console.WriteLine(errorMsg);
                System.Diagnostics.Debug.WriteLine(errorMsg);
                if (isDebug)
                    DbDebug($"{ex.Message}\n{sql}", "QueryScalarAsync");
                return default;
            }
            finally
            {
                if (isNewConn && conn != null)
                {
                    conn.Dispose();
                }
            }
        }

        public static async Task<List<T>> QueryListAsync<T>(string sql, params IDataParameter[] parameters) where T : new()
        {
            return await QueryListAsync<T>(sql, parameters, null);
        }

        public static async Task<List<T>> QueryAsync<T>(string sql, IDataParameter[]? parameters = null, IDbTransaction? trans = null)
        {
            IDbConnection? conn = trans?.Connection;

            if (trans != null && conn == null) trans = null;

            bool isNewConn = false;

            if (conn == null)
            {
                conn = DbProviderFactory.CreateConnection();
                if (conn is DbConnection dbConn) await dbConn.OpenAsync(); else conn.Open();
                isNewConn = true;
            }

            try
            {
                LogSql(sql, parameters);
                using var cmd = conn.CreateCommand();
                cmd.CommandText = sql;
                cmd.Transaction = Unwrap(trans);
                cmd.CommandTimeout = 60;
                var processedParameters = ProcessParameters(parameters ?? Array.Empty<IDataParameter>());
                if (processedParameters != null)
                {
                    foreach (var p in processedParameters) cmd.Parameters.Add(p);
                }

                var results = new List<T>();
                using (var reader = await (cmd as DbCommand)?.ExecuteReaderAsync()!)
                {
                    var type = typeof(T);
                    var properties = type.GetProperties();

                    while (await reader.ReadAsync())
                    {
                        if (type.IsPrimitive || type == typeof(string) || type == typeof(decimal) || type == typeof(DateTime) || type == typeof(Guid))
                        {
                            results.Add((T)SqlHelper.ConvertValue(reader[0], type)!);
                        }
                        else if (type == typeof(object) || type.Name.Contains("AnonymousType"))
                        {
                            // 处理匿名类型或 object
                            var dict = new System.Dynamic.ExpandoObject() as IDictionary<string, object?>;
                            for (int i = 0; i < reader.FieldCount; i++)
                            {
                                dict[reader.GetName(i)] = reader.IsDBNull(i) ? null : reader.GetValue(i);
                            }
                            results.Add((T)(object)dict);
                        }
                        else
                        {
                            var item = Activator.CreateInstance<T>();
                            foreach (var prop in properties)
                            {
                                var colName = SqlHelper.GetColumnName(prop);
                                if (HasColumn(reader, colName))
                                {
                                    var dbValue = reader[colName];
                                    prop.SetValue(item, SqlHelper.ConvertFromDbValue(dbValue, prop));
                                }
                            }
                            results.Add(item);
                        }
                    }
                }

                return results;
            }
            catch (Exception ex)
            {
                Console.WriteLine($"[DB ERROR] Async operation failed: {ex.Message}");
                DbDebug($"{ex.Message}\nSQL: {sql}", "QueryAsync");
                return new List<T>();
            }
            finally
            {
                if (isNewConn && conn != null)
                {
                    conn.Dispose();
                }
            }
        }

        public static async Task<List<T>> QueryListAsync<T>(string sql, IDataParameter[]? parameters = null, IDbTransaction? trans = null) where T : new()
        {
            IDbConnection? conn = trans?.Connection;

            if (trans != null && conn == null) trans = null;

            bool isNewConn = false;

            if (conn == null)
            {
                conn = DbProviderFactory.CreateConnection();
                if (conn is DbConnection dbConn) await dbConn.OpenAsync(); else conn.Open();
                isNewConn = true;
            }

            try
            {
                LogSql(sql, parameters);
                using var cmd = conn.CreateCommand();
                cmd.CommandText = sql;
                cmd.Transaction = Unwrap(trans);
                cmd.CommandTimeout = 60;
                var processedParameters = ProcessParameters(parameters ?? Array.Empty<IDataParameter>());
                if (processedParameters != null)
                {
                    foreach (var p in processedParameters) cmd.Parameters.Add(p);
                }

                var result = new List<T>();
                var properties = typeof(T).GetProperties();

                using (var reader = await (cmd as DbCommand)?.ExecuteReaderAsync()!)
                {
                    while (await reader.ReadAsync())
                    {
                        var item = new T();
                        foreach (var prop in properties)
                        {
                            if (!prop.CanWrite) continue;
                            if (HasColumn(reader, prop.Name) && !reader.IsDBNull(reader.GetOrdinal(prop.Name)))
                            {
                                prop.SetValue(item, SqlHelper.ConvertValue(reader[prop.Name], prop.PropertyType));
                            }
                        }
                        result.Add(item);
                    }
                }

                return result;
            }
            catch (Exception ex)
            {
                Console.WriteLine($"[DB ERROR] Async operation failed: {ex.Message}");
                DbDebug($"{ex.Message}\nSQL: {sql}", "QueryListAsync");
                return new List<T>();
            }
            finally
            {
                if (isNewConn && conn != null)
                {
                    conn.Dispose();
                }
            }
        }

        public static async Task<T?> QuerySingleAsync<T>(string sql, IDbTransaction? trans = null, params IDataParameter[] parameters) where T : class, new()
        {
            IDbConnection? conn = trans?.Connection;

            if (trans != null && conn == null) trans = null;

            bool isNewConn = false;

            if (conn == null)
            {
                conn = DbProviderFactory.CreateConnection();
                if (conn is DbConnection dbConn) await dbConn.OpenAsync(); else conn.Open();
                isNewConn = true;
            }

            try
            {
                LogSql(sql, parameters);
                using var cmd = conn.CreateCommand();
                cmd.CommandText = sql;
                cmd.Transaction = Unwrap(trans);
                cmd.CommandTimeout = 60;
                var processedParameters = ProcessParameters(parameters);
                if (processedParameters != null)
                {
                    foreach (var p in processedParameters) cmd.Parameters.Add(p);
                }

                T? entity = null;
                using (var reader = await (cmd as DbCommand)?.ExecuteReaderAsync()!)
                {
                    if (await reader.ReadAsync())
                    {
                        entity = new T();
                        var props = typeof(T).GetProperties();

                        foreach (var prop in props)
                        {
                            if (!prop.CanWrite) continue;
                            var colName = SqlHelper.GetColumnName(prop);

                            int ordinal;
                            try
                            {
                                ordinal = reader.GetOrdinal(colName);
                            }
                            catch (IndexOutOfRangeException)
                            {
                                continue;
                            }

                            var dbValue = reader.IsDBNull(ordinal) ? DBNull.Value : reader.GetValue(ordinal);
                            var value = SqlHelper.ConvertFromDbValue(dbValue, prop);
                            prop.SetValue(entity, value);
                        }
                    }
                }

                return entity;
            }
            catch (Exception ex)
            {
                Console.WriteLine($"[DB ERROR] Async operation failed: {ex.Message}");
                DbDebug($"{ex.Message}\nSQL: {sql}", "QuerySingleAsync");
                return null;
            }
            finally
            {
                if (isNewConn && conn != null)
                {
                    conn.Dispose();
                }
            }
        }

        public static async Task<T> ExecScalarAsync<T>(string sql, params IDataParameter[] parameters) where T : struct
        {
            return await ExecScalarAsync<T>(sql, true, null, parameters);
        }

        public static async Task<T?> ExecScalarAsync<T>(string sql, bool isDebug = false, IDbTransaction? trans = null, params IDataParameter[] parameters)
        {
            IDbConnection? conn = trans?.Connection;

            if (trans != null && conn == null) trans = null;

            bool isNewConn = false;

            if (conn == null)
            {
                conn = DbProviderFactory.CreateConnection();
                if (conn is DbConnection dbConn) await dbConn.OpenAsync(); else conn.Open();
                isNewConn = true;
            }

            try
            {
                using var cmd = conn.CreateCommand();
                cmd.CommandText = sql;
                cmd.Transaction = Unwrap(trans);
                cmd.CommandTimeout = 60;
                var processedParameters = ProcessParameters(parameters);
                if (processedParameters != null)
                {
                    foreach (var p in processedParameters) cmd.Parameters.Add(p);
                }

                var result = await (cmd as DbCommand)?.ExecuteScalarAsync()!;

                if (result == null || result == DBNull.Value)
                    return default;

                result = SqlHelper.ConvertValue<T>(result, default!);

                return result != null ? (T)Convert.ChangeType(result, typeof(T)) : default;
            }
            catch (Exception ex)
            {
                if (isDebug)
                    DbDebug($"{ex.Message}\n{sql}", "ExecScalarAsync");
                else
                    Debug($"{ex.Message}\n{sql}");
                return default;
            }
            finally
            {
                if (isNewConn && conn != null)
                {
                    conn.Dispose();
                }
            }
        }

        // 传入SQL和参数，返回执行后的 IDataReader
        public static async Task<IDataReader> ExecuteReaderAsync(string sql, params IDataParameter[] parameters)
        {
            var conn = DbProviderFactory.CreateConnection();
            var cmd = conn.CreateCommand();
            cmd.CommandText = sql;
            cmd.CommandTimeout = 60;
            var processedParameters = ProcessParameters(parameters);
            if (processedParameters != null)
            {
                foreach (var p in processedParameters) cmd.Parameters.Add(p);
            }

            if (conn is DbConnection dbConn) await dbConn.OpenAsync(); else conn.Open();
            // 注意这里不使用 using，调用方负责关闭 reader 和连接
            return await (cmd as DbCommand)?.ExecuteReaderAsync(System.Data.CommandBehavior.CloseConnection)!;
        }      

        public static async Task<string> QueryAsJsonAsync(string sql, IDataParameter[]? parameters = null, IDbTransaction? trans = null)
        {
            parameters ??= [];
            var results = new List<object>();

            IDbConnection? conn = trans?.Connection;

            if (trans != null && conn == null) trans = null;

            bool isNewConn = false;

            if (conn == null)
            {
                conn = DbProviderFactory.CreateConnection();
                if (conn is DbConnection dbConn) await dbConn.OpenAsync(); else conn.Open();
                isNewConn = true;
            }

            try
            {
                using var cmd = conn.CreateCommand();
                cmd.CommandText = sql;
                cmd.Transaction = Unwrap(trans);
                cmd.CommandTimeout = 60;
                var processedParameters = ProcessParameters(parameters);
                if (processedParameters != null)
                {
                    foreach (var p in processedParameters) cmd.Parameters.Add(p);
                }

                using var reader = await (cmd as DbCommand)?.ExecuteReaderAsync()!;

                if (reader.HasRows)
                {
                    var fieldCount = reader.FieldCount;

                    if (fieldCount == 1)
                    {
                        // 只有一列时，直接添加值
                        while (await reader.ReadAsync())
                        {
                            results.Add(reader[0]);
                        }
                    }
                    else
                    {
                        // 多列：添加字典
                        while (await reader.ReadAsync())
                        {
                            var row = new Dictionary<string, object?>();
                            for (int i = 0; i < fieldCount; i++)
                            {
                                var value = reader.GetValue(i);
                                row[reader.GetName(i).ToLower()] = value == DBNull.Value ? null : value;
                            }
                            results.Add(row);
                        }
                    }
                }

                return JsonConvert.SerializeObject(results);
            }
            catch (Exception ex)
            {
                DbDebug($"{ex.Message}\nSQL: {sql}", "QueryAsJsonAsync");
                return "[]";
            }
            finally
            {
                if (isNewConn && conn != null)
                {
                    conn.Dispose();
                }
            }
        }

        public static async Task<string> QueryResAsync(string sql, string format = "{0}", string countFormat = "", IDbTransaction? trans = null, params IDataParameter[] parameters)
        {
            DataSet ds = await QueryDatasetAsync(sql, trans, parameters);
            if (ds == null || ds.Tables.Count == 0)
                return "";

            var rows = new List<string>();
            int k = 1;
            foreach (DataRow reader in ds.Tables[0].Rows)
            {
                object[] rowValues = new object[ds.Tables[0].Columns.Count];
                for (int j = 0; j < ds.Tables[0].Columns.Count; j++)
                {
                    rowValues[j] = reader[j] is DBNull ? "" : reader[j];
                }

                string formattedRow = string.Format(format.Replace("{i}", k.ToString()), rowValues);
                rows.Add(formattedRow);
                k++;
            }

            countFormat = countFormat.Replace("{c}", (k - 1).ToString());
            return string.Join("", rows) + countFormat;
        }

        public static async Task<int> ExecAsync(string sql, params IDataParameter[] parameters)
        {
             return await ExecAsync(sql, true, null, parameters);
        }

        public static async Task<int> ExecAsync(string sql, bool isDebug = false, IDbTransaction? trans = null, params IDataParameter[] parameters)
        {
            IDbConnection? conn = trans?.Connection;

            if (trans != null && conn == null) trans = null;

            bool isNewConn = false;

            if (conn == null)
            {
                conn = DbProviderFactory.CreateConnection();
                if (conn is DbConnection dbConn) await dbConn.OpenAsync(); else conn.Open();
                isNewConn = true;
            }

            try
            {
                LogSql(sql, parameters);
                using var cmd = conn.CreateCommand();
                cmd.CommandText = sql;
                cmd.Transaction = SqlHelper.Unwrap(trans);
                cmd.CommandTimeout = 60;
                var processedParameters = ProcessParameters(parameters);
                if (processedParameters != null)
                {
                    foreach (var p in processedParameters) cmd.Parameters.Add(p);
                }

                int result = await (cmd as DbCommand)?.ExecuteNonQueryAsync()!;
                return result;
            }
            catch (Exception ex)
            {
                Console.WriteLine($"[DB ERROR] ExecAsync failed: {ex.Message}\nSQL: {sql}");
                if (isDebug)
                    DbDebug($"{ex.Message}\n{sql}", "ExecAsync");
                else
                    Logger.Error($"{ex.Message}\n{sql}", ex);
                return -1;
            }
            finally
            {
                if (isNewConn && conn != null)
                {
                    conn.Dispose();
                }
            }
        }

        public static async Task<DataSet> QueryDatasetAsync(string sql, IDbTransaction? trans = null, params IDataParameter[] parameters)
        {
            IDbConnection? conn = trans?.Connection;
            bool isNewConn = false;
            if (conn == null)
            {
                conn = DbProviderFactory.CreateConnection();
                if (conn is DbConnection dbConn) await dbConn.OpenAsync(); else conn.Open();
                isNewConn = true;
            }

            try
            {
                LogSql(sql, parameters);
                using var command = conn.CreateCommand();
                command.CommandText = sql;
                command.Transaction = Unwrap(trans);
                command.CommandTimeout = 60;
                var processedParameters = ProcessParameters(parameters);
                if (processedParameters != null)
                {
                    foreach (var p in processedParameters) command.Parameters.Add(p);
                }
                var adapter = CreateDataAdapter(command);
                DataSet dataSet = new();
                adapter.Fill(dataSet);

                return dataSet;
            }
            catch (Exception ex)
            {
                DbDebug($"{ex.Message}\nSQL: {sql}", "QueryDatasetAsync");
                throw;
            }
            finally
            {
                if (isNewConn && conn != null)
                {
                    conn.Dispose();
                }
            }
        }

        public static async Task<DataSet> QueryDatasetAsync(string sql, params IDataParameter[] parameters)
        {
            return await QueryDatasetAsync(sql, null, parameters);
        }

        /// <summary>
        /// 判断DataReader是否包含指定列名
        /// </summary>
        public static bool HasColumn(IDataReader reader, string columnName)
        {
            for (int i = 0; i < reader.FieldCount; i++)
            {
                if (string.Equals(reader.GetName(i), columnName, StringComparison.OrdinalIgnoreCase))
                    return true;
            }
            return false;
        }
    }
}

