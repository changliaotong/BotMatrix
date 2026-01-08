using System.Data;
using System.Runtime.CompilerServices;
using System.Text.Json;
using BotWorker.Common;
using Microsoft.Data.SqlClient;
using Newtonsoft.Json;

namespace BotWorker.Infrastructure.Persistence.Database
{
    public static partial class SQLConn
    {        
        public static string ConnString => GlobalConfig.ConnString;
        public static readonly SqlConnection conn = new(ConnString);

        public static object? ExecScalar(string sql, SqlParameter[]? parameters = null)
        {
            using var conn = new SqlConnection(ConnString);
            using var cmd = new SqlCommand(sql, conn);
            if (parameters != null)
                cmd.Parameters.AddRange(parameters);

            try
            {
                conn.Open();
                return cmd.ExecuteScalar();
            }
            catch (Exception ex)
            {
                DbDebug(ex.Message, "ExecScalar");
                return null;
            }        
        }

        public static T? ExecScalar<T>(string sql, SqlParameter[]? parameters = null)
        {
            var result = ExecScalar(sql, parameters);
            if (result == null || result == DBNull.Value)
                return default;

            result = SqlHelper.ConvertValue<T>(result, default!);

            return result != null ? (T)Convert.ChangeType(result, typeof(T)) : default!;
        }

        public static T? QueryScalar<T>(string sql, params SqlParameter[] parameters)
        {
            return QueryScalarInternal<T>(sql, parameters);
        }

        private static T? QueryScalarInternal<T>(string sql, SqlParameter[] parameters,
            [CallerMemberName] string callerName = "",
            [CallerFilePath] string callerFile = "",
            [CallerLineNumber] int callerLine = 0)
        {
            using var conn = new SqlConnection(ConnString);
            using var cmd = new SqlCommand(sql, conn);
            if (parameters != null && parameters.Length > 0)
                cmd.Parameters.AddRange(parameters);
            try
            {
                conn.Open();

                //ShowMessage(sql);

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
        }

        public static List<T>? QueryList<T>(string sql, params SqlParameter[] parameters)
        {
            return QueryList<T>(sql, true, parameters);
        }

        public static List<T>? QueryList<T>(string sql, bool isDebug = true, params SqlParameter[] parameters)
        {
            var json = QueryAsJson(sql, isDebug, parameters);
            return JsonConvert.DeserializeObject<List<T>>(json);
        }

        public static string QueryAsJson(string sql, bool isDebug, params SqlParameter[] parameters)
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

        public static SqlDataAdapter QueryDataAdapter(string query, params SqlParameter[] parameters)
        { 
            SqlDataAdapter dataAdapter = new(query, GetConn());
            DataSet dataSet = new();
            try
            {
                dataAdapter.Fill(dataSet);
            }
            catch (Exception ex)
            {
                DbDebug(ex.Message, "SQLConn.QueryDataAdapter");
                throw;
            }
            return dataAdapter;
        }

        public static DataSet QueryPage(string tableName, string keyField, string orderBy, int page, int pageSize, string sWhere, string sSelect = "*")
        {
            page = page <= 0 ? 1 : page;
            sWhere = sWhere.Trim().StartsWith("and", StringComparison.OrdinalIgnoreCase) ? sWhere : sWhere.IsNull() ? "" : $" and {sWhere}";
            orderBy = orderBy.IsNull() ? keyField : orderBy;
            string sql = $"exec sz84_robot..sp_GetPage {tableName.Quotes()}, {keyField.Quotes()}, {orderBy.Quotes()},  {page} , {pageSize} ,{sWhere.Quotes()}, {sSelect.Quotes()}";
            return QueryDataset(sql);
        }

        public static async Task<byte[]?> ExecuteScalarAsync(string sql, params SqlParameter[] parameters)
        {
            using SqlConnection conn = new(GetConn());
            try
            {
                conn.Open();
                SqlCommand command = new(sql, conn);
                parameters = ProcessParameters(parameters);
                if (parameters != null)
                {
                    command.Parameters.AddRange(parameters);
                }

                var result = await command.ExecuteScalarAsync();
                return result as byte[];
            }
            catch (Exception ex)
            {
                DbDebug($"{ex.Message}\n{sql}", "ExecuteScalarAsync");
                return null;
            }
        }

        private static SqlParameter[] ProcessParameters(SqlParameter[] parameters)
        {
            var processedParameters = new List<SqlParameter>();

            foreach (var param in parameters)
            {
                if (param.Value is JsonElement jsonElement)
                {
                    // 仅处理符合特定格式的 JsonElement
                    SqlParameter processedParam = ProcessJsonElement(param);
                    processedParameters.Add(processedParam);
                }
                else
                {
                    processedParameters.Add(param);
                }
            }

            return [.. processedParameters];
        }

        private static SqlParameter ProcessJsonElement(SqlParameter param)
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

        public static DataSet QueryDataset(string sql, params SqlParameter[] parameters)
        {
            using SqlConnection conn = new(GetConn());
            try
            {
                conn.Open();

                using SqlCommand command = new(sql, conn);
                if (parameters != null)
                {
                    parameters = ProcessParameters(parameters);
                    command.Parameters.AddRange(parameters);
                }

                using SqlDataAdapter adapter = new(command);
                DataSet ds = new();
                adapter.Fill(ds);
                return ds;
            }
            catch (Exception ex)
            {                
                DbDebug($"{ex.Message}\nSQL: {sql}", "QueryDataset");
                throw;
            }
        }

        public static IEnumerable<SqlDataReader> QueryReader(string query, params SqlParameter[] parameters)
        {
            using SqlConnection conn = new(GetConn());
            conn.Open();
            using SqlCommand command = new(query, conn);
            if (parameters != null) command.Parameters.AddRange(parameters);
            yield return command.ExecuteReader();
        }

        public static DataTable ExecuteQuery(string sql, params SqlParameter[] parameters)
        {
            using SqlConnection conn = new(GetConn());
            try
            {
                conn.Open();
                using SqlCommand command = new(sql, conn);
                parameters = ProcessParameters(parameters);
                if (parameters != null)
                {
                    command.Parameters.AddRange(parameters);
                }
                using SqlDataAdapter adapter = new(command);
                DataTable dataTable = new();
                adapter.Fill(dataTable);
                return dataTable;
            }
            catch (Exception ex)
            {
                DbDebug($"Error:{ex.Message}", "ExecuteQuery");
                throw;
            }

        }

        // 取得服务器当前时间
        public static DateTime GetDate()
        {
            return Convert.ToDateTime(GetTimeStamp());
        }

        // 取得服务器当前时间
        public static string GetTimeStamp()
        {
            return Query("select convert(varchar(19), GETDATE(), 120)");
        }

        public static (DataSet? ds, int i) QueryOrExec(string sql, params SqlParameter[] parameters)
        {
            return QueryOrExec(sql, true, parameters);
        }

        public static (DataSet? ds, int i) QueryOrExec(string sql, bool isDebug, params SqlParameter[] parameters)
        {
            if (sql.StartsWith("SELECT", StringComparison.OrdinalIgnoreCase))
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

        public static string Query(string sql, params SqlParameter[] parameters)
        {
            return QueryScalar<string>(sql, parameters) ?? "";
        }

        public static T Query<T>(string sql)
        {
            return Query<T>(sql, true, []);
        }

        public static T Query<T>(string sql, params SqlParameter[] parameters)
        {
            return Query<T>(sql, true, parameters);
        }

        public static T Query<T>(string sql, bool isDebug = true, params SqlParameter[] parameters)
        {
            try
            {
                using DataSet ds = QueryDataset(sql, parameters);
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
                    Debug($"{ex.Message}\n{sql}");
                return default!;
            }
        }

        public static int Exec(string sql)
        {
            return Exec(sql, []);
        }

        public static int Exec(params (string sql, SqlParameter[] parameters)[] sqlAndParameters)
        {
            return ExecTrans(sqlAndParameters);
        }

        public static int Exec(string sql, params SqlParameter[] parameters)
        {
            return Exec(sql, true, parameters);
        }

        // 执行sql命令
        public static int Exec(string sql, bool isDebug, params SqlParameter[] parameters)
        {
            using SqlConnection conn = new(GetConn());
            SqlCommand command = new(sql, conn);

            try
            {
                conn.Open();
                SqlParameter[] processedParameters = ProcessParameters(parameters);
                if (parameters != null)
                {
                    command.Parameters.AddRange(parameters);
                }

                int rowsAffected = command.ExecuteNonQuery();
                return rowsAffected;
            }
            catch (Exception ex)
            {
                if (isDebug)
                    DbDebug($"{ex.Message}\n{sql}");
                else 
                    Debug($"{ex.Message}\n{sql}");
                return -1;
            }
        }

        public static object ExecuteScalar(string query, params SqlParameter[] parameters)
        {
            using SqlConnection conn = new(GetConn());
            conn.Open();
            using SqlCommand command = new(query, conn);
            if (parameters != null)
            {
                command.Parameters.AddRange(parameters);
            }

            return command.ExecuteScalar();
        }

        public static void ExecSP(string spName, params SqlParameter[] parameters)
        {
            using SqlConnection conn = new(GetConn());
            conn.Open();
            using SqlCommand command = new(spName, conn);
            command.CommandType = CommandType.StoredProcedure;
            if (parameters != null)
            {
                command.Parameters.AddRange(parameters);
            }

            command.ExecuteNonQuery();
        }

        public static SqlDataReader QueryDataReaderPage(string sql, SqlParameter[] parameters, int page, int size)
        {
            int offset = (page - 1) * size;
            sql = sql + " OFFSET " + offset + " ROWS FETCH NEXT " + size + " ROWS ONLY";

            return QueryDataReader(sql, parameters);
        }

        public static SqlDataReader QueryDataReader(string sql, params SqlParameter[] parameters)
        {
            using SqlConnection conn = new(GetConn());
            try
            {
                conn.Open();

                using SqlCommand command = new(sql, conn);
                if (parameters != null)
                {
                    command.Parameters.AddRange(parameters);
                }

                return command.ExecuteReader();
            }
            catch (SqlException sqlEx)
            { 
                DbDebug($"{sqlEx.Message}\nSQL: {sql}", "QueryDataReader - SQL Exception");
                throw;
            }
            catch (Exception ex)
            {                
                DbDebug($"{ex.Message}\nSQL: {sql}", "QueryDataReader - General Exception");
                throw;
            }
        }

        public static DataSet QueryDatasetPage(string sql, int page, int size)
        {
            int offset = (page - 1) * size;
            sql += " OFFSET " + offset + " ROWS FETCH NEXT " + size + " ROWS ONLY";
            return QueryDataset(sql);
        }

        public static void BulkInsert(string tableName, DataTable dataTable)
        {
            try
            {
                using SqlConnection connection = new(GetConn());
                connection.Open();

                using SqlBulkCopy bulkCopy = new(connection)
                {
                    DestinationTableName = tableName,
                    BatchSize = 1000,       // 可根据需要调整
                    BulkCopyTimeout = 60    // 秒，默认30秒，有时大数据需要调大
                };

                // 如果 DataTable 列名和数据库表列名不完全对应，这里可以添加 ColumnMappings
                // 例如:
                // bulkCopy.ColumnMappings.Add("SourceColumnName", "DestinationColumnName");

                bulkCopy.WriteToServer(dataTable);
            }
            catch (Exception ex)
            {
                DbDebug($"BulkInsert Error: {ex.Message}", "BulkInsert");
                throw;
            }
        }

        public static void BulkUpdate(string tableName, List<DataRow> rowsToUpdate)
        {
            // 打开数据库连接
            OpenConnection();

            // 使用 SqlBulkCopy 执行批量更新
            using (SqlBulkCopy bulkCopy = new(conn))
            {
                bulkCopy.DestinationTableName = tableName;
                bulkCopy.ColumnMappings.Clear();
                foreach (DataColumn column in rowsToUpdate[0].Table.Columns)
                {
                    bulkCopy.ColumnMappings.Add(column.ColumnName, column.ColumnName);
                }

                // 创建临时表，用于存储更新的数据
                DataTable tempTable = rowsToUpdate[0].Table.Clone();
                foreach (DataRow row in rowsToUpdate)
                {
                    tempTable.ImportRow(row);
                }

                // 将临时表写入数据库
                bulkCopy.WriteToServer(tempTable);
            }

            // 关闭数据库连接
            CloseConnection();
        }

        public static long GetAutoId(string tableName)
        {
            return Query($"SELECT IDENT_CURRENT({tableName.Quotes()})").AsLong();
        }

        public static string QueryRes(string sql, string format = "{0}", string countFormat = "")
        {
            using SqlConnection conn = new(GetConn());
            try
            {
                conn.Open();
                using SqlDataReader reader = new SqlCommand(sql, conn).ExecuteReader();
                var rows = new List<string>();
                int k = 1;
                while (reader.Read())
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
                DbDebug($"Ex.Message:{ex.Message}\nSQL:{sql}", "QueryRes");
                return "";
            }
        }


        // 调试信息 红字显示 并写入数据库
        public static int DbDebug(object bugInfo, string? bugGroup = null)
        {
            Debug(bugInfo.AsString());
            return Bug.Insert(bugInfo, bugGroup);
        }

    }
}

