using System.Data;
using System.Linq;
using System.Runtime.CompilerServices;
using System.Text.Json;
using BotWorker.Common;
using Newtonsoft.Json;
using System.Data.Common;
using Microsoft.Data.SqlClient;
using Npgsql;

namespace BotWorker.Infrastructure.Persistence.Database
{
    public static partial class SQLConn
    {        
        public static string ConnString => GlobalConfig.ConnString;

        private static void LogSql(string sql, IDataParameter[]? parameters)
        {
            if (parameters == null || parameters.Length == 0)
            {
                Console.WriteLine($"[DB SQL] {sql}");
                return;
            }

            var paramStrings = parameters.Select(p => $"{p.ParameterName}={p.Value ?? "NULL"}");
            Console.WriteLine($"[DB SQL] {sql} | {string.Join(", ", paramStrings)}");
        }
        
        private static DbDataAdapter CreateDataAdapter(IDbCommand command)
        {
            return DbProviderFactory.CreateDataAdapter(command);
        }

        private static IDataParameter CreateParameter(string name, object value)
        {
            return DbProviderFactory.CreateParameter(name, value);
        }

        public static object? ExecScalar(string sql, IDataParameter[]? parameters = null)
        {
            using var conn = DbProviderFactory.CreateConnection();
            using var cmd = conn.CreateCommand();
            cmd.CommandText = sql;
            cmd.CommandTimeout = 60;
            if (parameters != null)
            {
                foreach (var p in parameters) cmd.Parameters.Add(p);
            }

            try
            {
                LogSql(sql, parameters);
                Console.WriteLine($"[DB INFO] Opening connection to: {GlobalConfig.DbType} (ConnString length: {ConnString?.Length ?? 0})");
                conn.Open();
                Console.WriteLine("[DB INFO] Connection opened successfully.");
                return cmd.ExecuteScalar();
            }
            catch (Exception ex)
            {
                Console.WriteLine($"[DB ERROR] Connection failed: {ex.Message}");
                DbDebug(ex.Message, "ExecScalar");
                return null;
            }        
        }

        public static T? ExecScalar<T>(string sql, IDataParameter[]? parameters = null)
        {
            var result = ExecScalar(sql, parameters);
            if (result == null || result == DBNull.Value)
                return default;

            result = SqlHelper.ConvertValue<T>(result, default!);

            return result != null ? (T)Convert.ChangeType(result, typeof(T)) : default!;
        }

        public static T? QueryScalar<T>(string sql, IDbTransaction? trans = null, params IDataParameter[] parameters)
        {
            return QueryScalarInternal<T>(sql, trans, parameters);
        }

        private static T? QueryScalarInternal<T>(string sql, IDbTransaction? trans, IDataParameter[] parameters,
            [CallerMemberName] string callerName = "",
            [CallerFilePath] string callerFile = "",
            [CallerLineNumber] int callerLine = 0)
        {
            trans ??= ORM.MetaData.CurrentTransaction.Value;
            IDbConnection? conn = trans?.Connection;

            if (trans != null && conn == null) trans = null;

            bool isNewConn = false;
            if (conn == null)
            {
                conn = DbProviderFactory.CreateConnection();
                conn.Open();
                isNewConn = true;
            }

            try
            {
                LogSql(sql, parameters);
                using var cmd = conn.CreateCommand();
                cmd.CommandText = sql;
                cmd.Transaction = trans;
                cmd.CommandTimeout = 60;
                if (parameters != null && parameters.Length > 0)
                {
                    foreach (var p in parameters) cmd.Parameters.Add(p);
                }

                object? result = cmd.ExecuteScalar();

                if (result == null || result == DBNull.Value)
                    return default;

                result = SqlHelper.ConvertValue<T>(result, default!);

                return result != null ? (T)Convert.ChangeType(result, typeof(T)) : default!;
            }
            catch (Exception ex)
            {
                var debugInfo = $"[QueryScalar Exception]\n" +
                                $"- Caller: {callerName}\n" +
                                $"- File: {Path.GetFileName(callerFile)}\n" +
                                $"- Line: {callerLine}\n" +
                                $"- SQL: {sql}\n" +
                                $"- Message: {ex.Message}\n" +
                                $"- StackTrace: {ex.StackTrace}";

                DbDebug(debugInfo, "QueryScalar");
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

        public static List<T>? QueryList<T>(string sql, params IDataParameter[] parameters)
        {
            return QueryList<T>(sql, true, parameters);
        }

        public static List<T>? QueryList<T>(string sql, bool isDebug = true, params IDataParameter[] parameters)
        {
            var json = QueryAsJson(sql, isDebug, parameters);
            return JsonConvert.DeserializeObject<List<T>>(json);
        }

        public static string QueryAsJson(string sql, bool isDebug, params IDataParameter[] parameters)
        {
            // 如果没有传递参数，则使用一个空数组
            parameters ??= [];

            DataSet ds = QueryDataset(sql, parameters);
            var results = new List<object>();

            if (ds.Tables.Count > 0)
            {
                if (ds.Tables[0].Columns.Count == 1)  // 只有一列时，简化为值列表
                {
                    foreach (DataRow dr in ds.Tables[0].Rows)
                    {
                        results.Add(dr[0]);  // 直接获取唯一列的值
                    }
                }
                else
                {
                    foreach (DataRow dr in ds.Tables[0].Rows)
                    {
                        var row = new Dictionary<string, object>();

                        // 遍历 DataRow 的每一列，创建键值对
                        foreach (DataColumn column in ds.Tables[0].Columns)
                        {
                            row[column.ColumnName] = dr[column];
                        }
                        results.Add(row);  // 添加字典
                    }
                }
            }
            return JsonConvert.SerializeObject(results);
        }

        public static IDataParameter[] CreateParameters(params (string Name, object Value)[] args)
        {
            return args.Select(a => DbProviderFactory.CreateParameter(a.Name, a.Value)).ToArray();
        }

        public static IDataAdapter QueryDataAdapter(string query, params IDataParameter[] parameters)
        { 
            using var conn = DbProviderFactory.CreateConnection();
            using var cmd = conn.CreateCommand();
            cmd.CommandText = query;
            foreach (var p in parameters) cmd.Parameters.Add(p);
            
            var adapter = CreateDataAdapter(cmd);
            DataSet dataSet = new();
            try
            {
                adapter.Fill(dataSet);
            }
            catch (Exception ex)
            {
                DbDebug(ex.Message, "SQLConn.QueryDataAdapter");
                throw;
            }
            return adapter;
        }

        public static DataSet QueryPage(string tableName, string keyField, string orderBy, int page, int pageSize, string sWhere, string sSelect = "*")
        {
            if (GlobalConfig.DbType == DatabaseType.SqlServer)
            {
                var parameters = new IDataParameter[]
                {
                    CreateParameter("TableName", tableName),
                    CreateParameter("KeyField", keyField),
                    CreateParameter("OrderField", orderBy),
                    CreateParameter("PageIndex", page),
                    CreateParameter("PageSize", pageSize),
                    CreateParameter("Where", sWhere),
                    CreateParameter("SelectField", sSelect)
                };
                return QueryDataset("sp_GetPage", null, parameters);
            }
            else
            {
                int offset = (page - 1) * pageSize;
                string whereClause = string.IsNullOrWhiteSpace(sWhere) ? "" : $" WHERE {sWhere}";
                string sql = $"SELECT {sSelect} FROM {tableName}{whereClause} ORDER BY {orderBy} LIMIT {pageSize} OFFSET {offset}";
                return QueryDataset(sql);
            }
        }

        public static async Task<byte[]?> ExecuteScalarAsync(string sql, params IDataParameter[] parameters)
        {
            using var conn = DbProviderFactory.CreateConnection();
            try
            {
                LogSql(sql, parameters);
                Console.WriteLine($"[DB INFO] Opening async connection (ExecuteScalarAsync) to: {GlobalConfig.DbType}");
                if (conn is DbConnection dbConn) await dbConn.OpenAsync(); else conn.Open();
                Console.WriteLine("[DB INFO] Async connection opened successfully.");
                using var command = conn.CreateCommand();
                command.CommandText = sql;
                var processedParameters = ProcessParameters(parameters);
                if (processedParameters != null)
                {
                    foreach (var p in processedParameters) command.Parameters.Add(p);
                }

                var result = await (command as DbCommand)?.ExecuteScalarAsync()!;
                return result as byte[];
            }
            catch (Exception ex)
            {
                DbDebug($"{ex.Message}\n{sql}", "ExecuteScalarAsync");
                return null;
            }
        }

        private static IDataParameter[] ProcessParameters(IDataParameter[] parameters)
        {
            var processedParameters = new List<IDataParameter>();

            foreach (var param in parameters)
            {
                if (param.Value is JsonElement jsonElement)
                {
                    // 仅处理符合特定格式的 JsonElement
                    IDataParameter processedParam = ProcessJsonElement(param);
                    processedParameters.Add(processedParam);
                }
                else
                {
                    processedParameters.Add(param);
                }
            }

            return [.. processedParameters];
        }

        private static IDataParameter ProcessJsonElement(IDataParameter param)
        {
            if (param.Value is JsonElement jsonElement)
            {
                if (jsonElement.ValueKind == JsonValueKind.Number)
                {
                    // 将 JsonElement 中的数字转换为 int 类型
                    string numericValue = jsonElement.ToString();
                    if (int.TryParse(numericValue, out int intValue))
                    {
                        param.Value = intValue;
                    }
                }
                // 如果不是符合条件的 JsonElement，则保持原样
            }
            return param;
        }

        public static DataSet QueryDataset(string sql, IDbTransaction? trans = null, params IDataParameter[] parameters)
        {
            trans ??= ORM.MetaData.CurrentTransaction.Value;
            IDbConnection? conn = trans?.Connection;

            if (trans != null && conn == null) trans = null;

            bool isNewConn = false;
            if (conn == null)
            {
                conn = DbProviderFactory.CreateConnection();
                Console.WriteLine($"[DB INFO] Opening connection (QueryDataset) to: {GlobalConfig.DbType}");
                conn.Open();
                Console.WriteLine("[DB INFO] Connection opened successfully.");
                isNewConn = true;
            }

            try
            {
                LogSql(sql, parameters);
                using var command = conn.CreateCommand();
                command.CommandText = sql;
                command.Transaction = trans;
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
                DbDebug($"{ex.Message}\nSQL: {sql}", "QueryDataset");
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

        public static DataSet QueryDataset(string sql, params IDataParameter[] parameters)
        {
            return QueryDataset(sql, null, parameters);
        }

        public static DataTable ExecuteQuery(string sql, IDbTransaction? trans = null, params IDataParameter[] parameters)
        {
            trans ??= ORM.MetaData.CurrentTransaction.Value;
            IDbConnection? conn = trans?.Connection;
            bool isNewConn = false;
            if (conn == null)
            {
                conn = DbProviderFactory.CreateConnection();
                conn.Open();
                isNewConn = true;
            }

            try
            {
                LogSql(sql, parameters);
                using var command = conn.CreateCommand();
                command.CommandText = sql;
                command.Transaction = trans;
                command.CommandTimeout = 60;
                var processedParameters = ProcessParameters(parameters);
                if (processedParameters != null)
                {
                    foreach (var p in processedParameters) command.Parameters.Add(p);
                }
                var adapter = CreateDataAdapter(command);
                DataTable dataTable = new();
                adapter.Fill(dataTable);
                return dataTable;
            }
            catch (Exception ex)
            {
                DbDebug($"Error:{ex.Message}\nSQL: {sql}", "ExecuteQuery");
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

        public static DataTable ExecuteQuery(string sql, params IDataParameter[] parameters)
        {
            return ExecuteQuery(sql, null, parameters);
        }

        // 取得服务器当前时间
        public static DateTime GetDate()
        {
            return Convert.ToDateTime(GetTimeStamp());
        }

        // 取得服务器当前时间
        public static string GetTimeStamp()
        {
            string sql = GlobalConfig.DbType == DatabaseType.SqlServer ? "select getdate()" : "select to_char(now(), 'YYYY-MM-DD HH24:MI:SS')";
            return Query(sql);
        }

        public static (DataSet? ds, int i) QueryOrExec(string sql, params IDataParameter[] parameters)
        {
            return QueryOrExec(sql, true, parameters);
        }

        public static (DataSet? ds, int i) QueryOrExec(string sql, bool isDebug, params IDataParameter[] parameters)
        {
            if (sql.TrimStart().StartsWith("SELECT", StringComparison.OrdinalIgnoreCase))
            {
                using DataSet? ds = QueryDataset(sql, parameters);
                if (ds != null && ds.Tables.Count > 0)
                    return (ds, ds.Tables[0].Rows.Count);
                else
                    return (null, 0);
            }
            else
            {
                return (null, Exec(sql, isDebug, parameters));
            }
        }

        public static string Query(string sql, params IDataParameter[] parameters)
        {
            return QueryScalar<string>(sql, null, parameters) ?? "";
        }

        public static T Query<T>(string sql)
        {
            return Query<T>(sql, true, null, []);
        }

        public static T Query<T>(string sql, params IDataParameter[] parameters)
        {
            return Query<T>(sql, true, null, parameters);
        }

        public static T Query<T>(string sql, bool isDebug = true, IDbTransaction? trans = null, params IDataParameter[] parameters)
        {
            trans ??= ORM.MetaData.CurrentTransaction.Value;
            try
            {
                using DataSet ds = QueryDataset(sql, trans, parameters);
                if (ds != null && ds.Tables.Count > 0 && ds.Tables[0].Rows.Count > 0)
                {
                    var val = ds.Tables[0].Rows[0][0];
                    if (val == DBNull.Value || val == null)
                        return default!;

                    if (val is T tVal)
                        return tVal;

                    // 尝试字符串转换，支持常见类型转换
                    return val.ToString()!.As<T>();
                }
                return default!;
            }
            catch (Exception ex)
            {
                if (isDebug)
                    DbDebug($"{ex.Message}\n{sql}");
                else
                    Logger.Error($"{ex.Message}\n{sql}");
                return default!;
            }
        }

        public static int Exec(string sql)
        {
            return Exec(sql, Array.Empty<IDataParameter>());
        }

        public static int Exec(params (string sql, IDataParameter[] parameters)[] sqlAndParameters)
        {
            return ExecTrans(sqlAndParameters);
        }

        public static int Exec(string sql, params IDataParameter[] parameters)
        {
            return Exec(sql, true, null, parameters);
        }

        public static int Exec(string sql, bool isDebug, params IDataParameter[] parameters)
        {
            return Exec(sql, isDebug, null, parameters);
        }

        // 执行sql命令
        public static int Exec(string sql, bool isDebug = true, IDbTransaction? trans = null, params IDataParameter[] parameters)
        {
            trans ??= ORM.MetaData.CurrentTransaction.Value;
            IDbConnection? conn = trans?.Connection;
            bool isNewConn = false;

            if (conn == null)
            {
                conn = DbProviderFactory.CreateConnection();
                conn.Open();
                isNewConn = true;
            }

            try
            {
                LogSql(sql, parameters);
                using var command = conn.CreateCommand();
                command.CommandText = sql;
                command.Transaction = trans;
                command.CommandTimeout = 60;
                if (parameters != null)
                {
                    var processedParameters = ProcessParameters(parameters);
                    foreach (var p in processedParameters) command.Parameters.Add(p);
                }
                return command.ExecuteNonQuery();
            }
            catch (Exception ex)
            {
                if (isDebug)
                    DbDebug($"{ex.Message}\nSQL: {sql}", "Exec");
                else
                    Logger.Error($"{ex.Message}\nSQL: {sql}", ex);
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

        public static object? ExecuteScalar(string query, params IDataParameter[] parameters)
        {
            LogSql(query, parameters);
            using var conn = DbProviderFactory.CreateConnection();
            conn.Open();
            using var command = conn.CreateCommand();
            command.CommandText = query;
            command.CommandTimeout = 60;
            if (parameters != null)
            {
                foreach (var p in parameters) command.Parameters.Add(p);
            }

            return command.ExecuteScalar();
        }

        public static IDataReader QueryDataReader(string sql, params IDataParameter[] parameters)
        {
            LogSql(sql, parameters);
            var conn = DbProviderFactory.CreateConnection();
            try
            {
                conn.Open();

                using var command = conn.CreateCommand();
                command.CommandText = sql;
                command.CommandTimeout = 60;
                if (parameters != null)
                {
                    foreach (var p in parameters) command.Parameters.Add(p);
                }

                return command.ExecuteReader(CommandBehavior.CloseConnection);
            }
            catch (Exception ex)
            {                
                DbDebug($"{ex.Message}\nSQL: {sql}", "QueryDataReader");
                conn.Dispose();
                throw;
            }
        }

        public static void BulkInsert(string tableName, DataTable dataTable)
        {
            if (GlobalConfig.DbType == DatabaseType.SqlServer)
            {
                using var bulkCopy = new SqlBulkCopy(ConnString);
                bulkCopy.DestinationTableName = tableName;
                try
                {
                    bulkCopy.WriteToServer(dataTable);
                }
                catch (Exception ex)
                {
                    DbDebug($"BulkInsert Error: {ex.Message}", "BulkInsert");
                    throw;
                }
            }
            else if (GlobalConfig.DbType == DatabaseType.PostgreSql)
            {
                try
                {
                    using var conn = new NpgsqlConnection(ConnString);
                    conn.Open();
                    
                    var columns = string.Join(", ", dataTable.Columns.Cast<DataColumn>().Select(c => $"\"{c.ColumnName}\""));
                    using var writer = conn.BeginBinaryImport($"COPY \"{tableName}\" ({columns}) FROM STDIN (FORMAT BINARY)");

                    foreach (DataRow row in dataTable.Rows)
                    {
                        writer.StartRow();
                        foreach (DataColumn col in dataTable.Columns)
                        {
                            writer.Write(row[col]);
                        }
                    }
                    writer.Complete();
                }
                catch (Exception ex)
                {
                    DbDebug($"BulkInsert Error: {ex.Message}", "BulkInsert");
                    throw;
                }
            }
        }

        public static long GetAutoId(string tableName)
        {
            string sql = GlobalConfig.DbType == DatabaseType.SqlServer ? "SELECT SCOPE_IDENTITY()" : "SELECT lastval()"; 
            return QueryScalar<long>(sql);
        }

        public static string QueryRes(string sql, string format = "{0}", string countFormat = "", IDbTransaction? trans = null)
        {
            DataSet ds = QueryDataset(sql, trans);
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

        // 调试信息 红字显示 并写入数据库
        public static int DbDebug(object bugInfo, string? bugGroup = null)
        {
            var msg = bugInfo.AsString();
            Debug(msg);
            Console.WriteLine($"[DB DEBUG][{bugGroup}] {msg}");
            return 1;
        }

        private static void Debug(string message)
        {
             Logger.Debug(message);
        }
    }
}
