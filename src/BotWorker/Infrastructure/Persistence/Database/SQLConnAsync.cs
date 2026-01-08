using System.Data;
using System.Data.Common;
using System.Diagnostics;
using Newtonsoft.Json;

namespace BotWorker.Infrastructure.Persistence.Database
{
    public static partial class SQLConn
    {
        public static async Task<Dictionary<string, object>> ExecWithOutputAsync(string sql, IDataParameter[] parameters, string[] outputFields, IDbTransaction? trans = null)
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
                using var cmd = conn.CreateCommand();
                cmd.CommandText = sql;
                cmd.Transaction = trans;
                
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
                cmd.Transaction = trans;
                
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

        public static async Task<List<T>> QueryAsync<T>(string sql, IDataParameter[]? parameters = null)
        {
            using var conn = DbProviderFactory.CreateConnection();
            if (conn is DbConnection dbConn) await dbConn.OpenAsync(); else conn.Open();

            using var cmd = conn.CreateCommand();
            cmd.CommandText = sql;
            var processedParameters = ProcessParameters(parameters ?? Array.Empty<IDataParameter>());
            if (processedParameters != null)
            {
                foreach (var p in processedParameters) cmd.Parameters.Add(p);
            }

            using var reader = await (cmd as DbCommand)?.ExecuteReaderAsync()!;
            var results = new List<T>();

            var type = typeof(T);
            var isSimple = IsSimpleType(type);
            var isDynamic = type == typeof(object) || type.Name == "Object";

            while (await reader.ReadAsync())
            {
                if (isSimple)
                {
                    results.Add((T)reader[0]);
                }
                else if (isDynamic)
                {
                    var expando = new System.Dynamic.ExpandoObject();
                    var dict = (IDictionary<string, object?>)expando;
                    for (int i = 0; i < reader.FieldCount; i++)
                    {
                        dict[reader.GetName(i)] = reader.IsDBNull(i) ? null : reader.GetValue(i);
                    }
                    results.Add((T)(object)expando);
                }
                else
                {
                    // Fallback to basic object mapping if T is a class
                    var item = Activator.CreateInstance<T>();
                    var props = type.GetProperties();
                    foreach (var prop in props)
                    {
                        if (HasColumn(reader, prop.Name) && !reader.IsDBNull(reader.GetOrdinal(prop.Name)))
                        {
                            prop.SetValue(item, reader[prop.Name]);
                        }
                    }
                    results.Add(item);
                }
            }

            return results;
        }

        public static async Task<List<T>> QueryListAsync<T>(string sql, IDataParameter[]? parameters = null, IDbTransaction? trans = null) where T : new()
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
                using var cmd = conn.CreateCommand();
                cmd.CommandText = sql;
                cmd.Transaction = trans;
                var processedParameters = ProcessParameters(parameters ?? Array.Empty<IDataParameter>());
                if (processedParameters != null)
                {
                    foreach (var p in processedParameters) cmd.Parameters.Add(p);
                }

                using var reader = await (cmd as DbCommand)?.ExecuteReaderAsync()!;
                var result = new List<T>();
                var properties = typeof(T).GetProperties();

                while (await reader.ReadAsync())
                {
                    var item = new T();
                    foreach (var prop in properties)
                    {
                        if (HasColumn(reader, prop.Name) && !reader.IsDBNull(reader.GetOrdinal(prop.Name)))
                        {
                            prop.SetValue(item, reader[prop.Name]);
                        }
                    }
                    result.Add(item);
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

        public static async Task<T?> QuerySingleAsync<T>(string sql, params IDataParameter[] parameters) where T : class, new()
        {
            using var conn = DbProviderFactory.CreateConnection();
            if (conn is DbConnection dbConn) await dbConn.OpenAsync(); else conn.Open();

            using var cmd = conn.CreateCommand();
            cmd.CommandText = sql;
            var processedParameters = ProcessParameters(parameters);
            if (processedParameters != null)
            {
                foreach (var p in processedParameters) cmd.Parameters.Add(p);
            }

            using var reader = await (cmd as DbCommand)?.ExecuteReaderAsync()!;
            if (!await reader.ReadAsync()) return null;

            var entity = new T();
            var props = typeof(T).GetProperties();

            foreach (var prop in props)
            {
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

            return entity;
        }

        public static async Task<T> ExecScalarAsync<T>(string sql, params IDataParameter[] parameters) where T : struct
        {
            return await ExecScalarAsync<T>(sql, true, null, parameters);
        }

        public static async Task<T> ExecScalarAsync<T>(string sql, bool isDebug = false, IDbTransaction? trans = null, params IDataParameter[] parameters) where T : struct
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
                using var cmd = conn.CreateCommand();
                cmd.CommandText = sql;
                cmd.Transaction = trans;
                var processedParameters = ProcessParameters(parameters);
                if (processedParameters != null)
                {
                    foreach (var p in processedParameters) cmd.Parameters.Add(p);
                }

                var result = await (cmd as DbCommand)?.ExecuteScalarAsync()!;

                if (result == null || result == DBNull.Value)
                    return default!;

                result = SqlHelper.ConvertValue<T>(result, default!);

                return result != null ? (T)Convert.ChangeType(result, typeof(T)) : default!;
            }
            catch (Exception ex)
            {
                if (isDebug)
                    DbDebug($"{ex.Message}\n{sql}", "ExecScalarAsync");
                else
                    Debug($"{ex.Message}\n{sql}");
                return default!;
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
            var processedParameters = ProcessParameters(parameters);
            if (processedParameters != null)
            {
                foreach (var p in processedParameters) cmd.Parameters.Add(p);
            }

            if (conn is DbConnection dbConn) await dbConn.OpenAsync(); else conn.Open();
            // 注意这里不使用 using，调用方负责关闭 reader 和连接
            return await (cmd as DbCommand)?.ExecuteReaderAsync(System.Data.CommandBehavior.CloseConnection)!;
        }      

        public static async Task<string> QueryAsJsonAsync(string sql, params IDataParameter[]? parameters)
        {
            parameters ??= [];

            var results = new List<object>();

            using var conn = DbProviderFactory.CreateConnection();
            if (conn is DbConnection dbConn) await dbConn.OpenAsync(); else conn.Open();

            using var cmd = conn.CreateCommand();
            cmd.CommandText = sql;
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

        public static async Task<string> QueryResAsync(string sql, string format = "{0}", string countFormat = "")
        {
            using var conn = DbProviderFactory.CreateConnection();
            try
            {
                if (conn is DbConnection dbConn) await dbConn.OpenAsync(); else conn.Open();
                using var cmd = conn.CreateCommand();
                cmd.CommandText = sql;
                using var reader = await (cmd as DbCommand)?.ExecuteReaderAsync()!;
                var rows = new List<string>();
                int k = 1;
                while (await reader.ReadAsync())
                {
                    object[] rowValues = new object[reader.FieldCount];
                    for (int j = 0; j < reader.FieldCount; j++)
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
            catch (Exception ex)
            {
                DbDebug($"Ex.Message:{ex.Message}\nSQL:{sql}", "QueryResAsync");
                return "";
            }
        }

        public static async Task<int> ExecAsync(string sql, params IDataParameter[] parameters)
        {
             return await ExecAsync(sql, true, null, parameters);
        }

        public static async Task<int> ExecAsync(string sql, bool isDebug = false, IDbTransaction? trans = null, params IDataParameter[] parameters)
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
                using var cmd = conn.CreateCommand();
                cmd.CommandText = sql;
                cmd.Transaction = trans;
                var processedParameters = ProcessParameters(parameters);
                if (processedParameters != null)
                {
                    foreach (var p in processedParameters) cmd.Parameters.Add(p);
                }

                return await (cmd as DbCommand)?.ExecuteNonQueryAsync()!;
            }
            catch (Exception ex)
            {
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
                using var command = conn.CreateCommand();
                command.CommandText = sql;
                command.Transaction = trans;
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

