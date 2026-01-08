using Newtonsoft.Json;
using System.Text;
using Microsoft.Data.SqlClient;
using BotWorker.Infrastructure.Extensions;

namespace BotWorker.Infrastructure.Persistence.Database
{
    //远程执行sql
    public static class Remote
    {    


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

