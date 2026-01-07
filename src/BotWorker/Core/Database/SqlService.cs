using Newtonsoft.Json;
using System.Text;
using Microsoft.Data.SqlClient;
using BotWorker.Common;
using BotWorker.Core.Database;

namespace BotWorker.Services
{
    public static class SqlService
    {
        public static async Task<int> ExecTransAsyncAPI(bool isDebug = true, params (string, SqlParameter[])[] sqls)
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
            var apiUrl = $"{url}/api/sql/query";
            return await GetWebResponse(apiUrl, request);
        }

        public static async Task<string> ExecTransAsyncAPI(ExecTransRequest request)
        {
            var apiUrl = $"{url}/api/sql/exec";
            return await GetWebResponse(apiUrl, request);
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
            client.DefaultRequestHeaders.Add("X-Api-Key", apiKey);
            var json = JsonConvert.SerializeObject(data);
            var content = new StringContent(json, Encoding.UTF8, "application/json");
            var response = await client.PostAsync(url, content);
            response.EnsureSuccessStatusCode(); // 检查响应是否成功
            return response;
        }
    }
}
