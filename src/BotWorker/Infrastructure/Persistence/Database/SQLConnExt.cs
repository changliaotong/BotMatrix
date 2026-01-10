using System.Data;
using System.Data.Common;
using Newtonsoft.Json;

namespace BotWorker.Infrastructure.Persistence.Database
{
    public static partial class SQLConn
    {
        /// <summary>
        /// 执行插入SQL，返回long和Guid两个主键
        /// </summary>
        public static async Task<(long Id, Guid Guid)> ExecuteInsertReturnKeysAsync(string sql, IDataParameter[] parameters)
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
            if (await reader.ReadAsync())
            {
                long id = Convert.ToInt64(reader.GetValue(0));
                Guid guid = reader.GetGuid(1);
                return (id, guid);
            }

            throw new Exception("插入操作未返回主键和 GUID");
        }
    }
}

