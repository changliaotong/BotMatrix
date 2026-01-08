using Newtonsoft.Json;
using System.Text;
using Microsoft.Data.SqlClient;
using sz84.common;
using BotWorker.Infrastructure.Extensions;

namespace BotWorker.Infrastructure.Persistence.Database
{
    //远程执行sql
    public static class Remote
    {
        //ExecTrans
        public static async Task<int> ExecTransAsync(List<(string, SqlParameter[])> sqls, bool isDebug = true)
        {
            return await ExecTransAsyncAPI(sqls, isDebug);
        }

        //Exec
        public static async Task<int> ExecAsync(string sql, bool isDebug = true, params SqlParameter[] parameters)
        {
            return await ExecTransAsync([(sql, parameters)], isDebug);
        }

        public static async Task<int> ExecTransAsyncAPI(List<(string, SqlParameter[])> sqls, bool isDebug = true)
        {
            var request = new ExecTransRequest
            {
                Sqls = sqls.Select(tuple => (tuple.Item1, tuple.Item2.ToDTOs())).ToList(),
                IsDebug = isDebug
            };

            var json = await ExecTransAsyncAPI(request);
            return json?.AsObject<int>() ?? -1;
        }

        public static List<SqlParameterDTO> ToDTOs(this SqlParameter[] sqlParameters)
        {
            return sqlParameters?.Select(p => new SqlParameterDTO
            {
                Name = p.ParameterName,
                Value = p.Value,
                DbType = p.SqlDbType.ToString()
            }).ToList() ?? [];
        }

        public static async Task<List<T>?> QueryAsync<T>(string sql, bool isDebug = true, params SqlParameter[] parameters)
        {
            return await QueryAsyncAPI<T>(sql, isDebug, parameters);
        }

        public static async Task<List<T>> QueryAsyncAPI<T>(string sql, bool isDebug, params SqlParameter[] parameters)
        {
            var request = new SqlRequest
            {
                Sql = sql,
                Parameters = parameters.ToDTOs(),
                IsDebug = isDebug
            };

            var json = await QueryAsyncAPI(request);
            return json?.AsObject<List<T>>() ?? [];
        }

        public static async Task<string> QueryAsyncAPI(SqlRequest request)
        {
            var url = $"{Common.url}/api/sql/query";
            return await GetWebResponse(url, request);
        }

        public static async Task<string> ExecTransAsyncAPI(ExecTransRequest request)
        {
            var url = $"{Common.url}/api/sql/exec";
            return await GetWebResponse(url, request);
        }

        public static async Task<string> GetWebResponse(string url, object request)
        {
            try
            {
                var json = JsonConvert.SerializeObject(request);
                var response = await PostDataAsync(url, json);

                if (response.IsSuccessStatusCode)
                {
                    var result = await response.Content.ReadAsStringAsync();
                    return result;
                }
                else
                {
                    SQLConn.DbDebug("Error: " + response.StatusCode, "SqlService:GetWebAPI");
                    return string.Empty;
                }
            }
            catch (Exception ex)
            {
                SQLConn.DbDebug("Exception: " + ex.Message, "SqlService:GetWebAPI");
                return string.Empty;
            }
        }

        public static async Task<HttpResponseMessage> PostDataAsync<T>(string url, T data)
        {
            HttpClient client = new();
            client.DefaultRequestHeaders.Add("X-Api-Key", Common.apiKey);
            var json = JsonConvert.SerializeObject(data);
            var content = new StringContent(json, Encoding.UTF8, "application/json");
            var response = await client.PostAsync(url, content);
            response.EnsureSuccessStatusCode(); // 检查响应是否成功
            return response;
        }
    }

    public class SqlRequest
    {
        public string Sql { get; set; } = string.Empty;

        public List<SqlParameterDTO> Parameters { get; set; } = [];

        public bool IsDebug { get; set; } = true;
    }

    public class ExecTransRequest
    {
        public List<(string, List<SqlParameterDTO>)>? Sqls { get; set; }

        public bool IsDebug { get; set; } = true;
    }

    //public class SqlParameterDTO
    //{
    //    public string Name { get; set; } = string.Empty;   // SQL 参数名
    //    public object? Value { get; set; }  // 参数值
    //    public string DbType { get; set; } = string.Empty; // 数据类型 (如: Int, String)

    //    public static List<SqlParameterDTO> ToDTOs(List<SqlParameter>? sqlParameters)
    //    {
    //        var parameterDTOs = new List<SqlParameterDTO>();

    //        if (sqlParameters != null)
    //        {
    //            foreach (var param in sqlParameters)
    //            {
    //                var paramDto = new SqlParameterDTO
    //                {
    //                    Name = param.ParameterName,        // 获取参数名
    //                    Value = param.Value,               // 获取参数值
    //                    DbType = param.SqlDbType.ToString() // 将 SqlDbType 转换为字符串
    //                };

    //                parameterDTOs.Add(paramDto);
    //            }
    //        }

    //        return parameterDTOs;
    //    }

    //}

}

