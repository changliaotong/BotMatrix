using Microsoft.Data.SqlClient;
using Newtonsoft.Json;

namespace BotWorker.Infrastructure.Persistence.Database
{
    public static partial class SQLConn
    {


        /// <summary>
        /// 执行插入SQL，返回long和Guid两个主键
        /// </summary>
        public static async Task<(long Id, Guid Guid)> ExecuteInsertReturnKeysAsync(string sql, SqlParameter[] parameters)
        {
            await using var conn = new SqlConnection(ConnString);
            await conn.OpenAsync();
            await using var cmd = new SqlCommand(sql, conn);
            cmd.Parameters.AddRange([.. parameters]);

            await using var reader = await cmd.ExecuteReaderAsync();
            if (await reader.ReadAsync())
            {
                long id = reader.GetInt64(0);
                Guid guid = reader.GetGuid(1);
                return (id, guid);
            }

            throw new Exception("插入操作未返回主键和 GUID");
        }
    }
}

