using System.Diagnostics;
using Microsoft.Data.SqlClient;
using Newtonsoft.Json;

namespace BotWorker.Infrastructure.Persistence.Database
{
    public static partial class SQLConn
    {
        public static async Task<Dictionary<string, object>> ExecWithOutputAsync(string sql, SqlParameter[] parameters, string[] outputFields, SqlTransaction? trans = null)
        {
            SqlConnection? conn = trans?.Connection;
            bool isNewConn = false;

            if (conn == null)
            {
                conn = new SqlConnection(ConnString);
                await conn.OpenAsync();
                isNewConn = true;
            }

            try
            {
                using var cmd = new SqlCommand(sql, conn, trans);
                cmd.Parameters.AddRange(parameters);

                using var reader = await cmd.ExecuteReaderAsync();

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
                    await conn.DisposeAsync();
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

        public static async Task<List<T>> QueryListAsync<T>(string sql, SqlParameter[]? parameters = null, SqlTransaction? trans = null) where T : new()
        {
            SqlConnection? conn = trans?.Connection;
            bool isNewConn = false;

            if (conn == null)
            {
                conn = new SqlConnection(ConnString);
                await conn.OpenAsync();
                isNewConn = true;
            }

            try
            {
                using var cmd = new SqlCommand(sql, conn, trans);
                if (parameters != null)
                {
                    cmd.Parameters.AddRange(parameters);
                }

                using var reader = await cmd.ExecuteReaderAsync();
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
                    await conn.DisposeAsync();
                }
            }
        }

        public static async Task<T?> QuerySingleAsync<T>(string sql, params  SqlParameter[] parameters) where T : class, new()
        {
            using var conn = new SqlConnection(ConnString);
            using var cmd = new SqlCommand(sql, conn);
            cmd.Parameters.AddRange(parameters);
            await conn.OpenAsync();

            using var reader = await cmd.ExecuteReaderAsync();
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

        public static async Task<T> ExecScalarAsync<T>(string sql, params SqlParameter[] parameters) where T : struct
        {
            return await ExecScalarAsync<T>(sql, true, null, parameters);
        }

        public static async Task<T> ExecScalarAsync<T>(string sql, bool isDebug = false, SqlTransaction? trans = null, params SqlParameter[] parameters) where T : struct
        {
            SqlConnection? conn = trans?.Connection;
            bool isNewConn = false;

            if (conn == null)
            {
                conn = new SqlConnection(ConnString);
                await conn.OpenAsync();
                isNewConn = true;
            }

            try
            {
                using var cmd = new SqlCommand(sql, conn, trans);
                if (parameters != null && parameters.Length > 0)
                    cmd.Parameters.AddRange(parameters);

                var result = await cmd.ExecuteScalarAsync();

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
                    await conn.DisposeAsync();
                }
            }
        }

        // 传入SQL和参数，返回执行后的 SqlDataReader
        public static async Task<SqlDataReader> ExecuteReaderAsync(string sql, params SqlParameter[] parameters)
        {
            var conn = new SqlConnection(ConnString);
            var cmd = new SqlCommand(sql, conn);
            cmd.Parameters.AddRange([.. parameters]);

            await conn.OpenAsync();
            // 注意这里不使用 using，调用方负责关闭 reader 和连接
            return await cmd.ExecuteReaderAsync(System.Data.CommandBehavior.CloseConnection);
        }      

        public static async Task<string> QueryAsJsonAsync(string sql, params SqlParameter[]? parameters)
        {
            parameters ??= [];

            var results = new List<object>();

            using var conn = new SqlConnection(ConnString);
            await conn.OpenAsync();

            using var cmd = new SqlCommand(sql, conn);
            if (parameters.Length > 0)
            {
                cmd.Parameters.AddRange(parameters);
            }

            using var reader = await cmd.ExecuteReaderAsync();

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

        public static async Task<int> ExecAsync(string sql, params SqlParameter[] parameters)
        {
             return await ExecAsync(sql, true, null, parameters);
        }

        public static async Task<int> ExecAsync(string sql, bool isDebug = false, SqlTransaction? trans = null, params SqlParameter[] parameters)
        {
            SqlConnection? conn = trans?.Connection;
            bool isNewConn = false;

            if (conn == null)
            {
                conn = new SqlConnection(ConnString);
                await conn.OpenAsync();
                isNewConn = true;
            }

            try
            {
                using var cmd = new SqlCommand(sql, conn, trans);
                if (parameters != null && parameters.Length > 0)
                    cmd.Parameters.AddRange(parameters);

                return await cmd.ExecuteNonQueryAsync();
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
                    await conn.DisposeAsync();
                }
            }
        }

        /// <summary>
        /// 判断DataReader是否包含指定列名
        /// </summary>
        public static bool HasColumn(SqlDataReader reader, string columnName)
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

