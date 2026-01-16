using System.Threading.Tasks;
using BotWorker.Common;
using BotWorker.Domain.Repositories;
using Dapper;

namespace BotWorker.Infrastructure.Persistence.Repositories
{
    public class IDCRepository : BaseRepository<object>, IIDCRepository
    {
        public IDCRepository() : base("IDC", GlobalConfig.BaseInfoConnection)
        {
        }

        public async Task<string?> GetAreaNameAsync(string areaCode)
        {
            string sql = $"SELECT dq FROM {_tableName} WHERE bm = @areaCode";
            using var conn = CreateConnection();
            return await conn.QueryFirstOrDefaultAsync<string>(sql, new { areaCode });
        }

        public async Task<string?> GetRandomBmAsync(string? dq = null)
        {
            string sql = $"SELECT bm FROM {_tableName}";
            if (!string.IsNullOrEmpty(dq))
            {
                sql += " WHERE dq LIKE @dq";
            }
            sql += " ORDER BY RANDOM() LIMIT 1";

            using var conn = CreateConnection();
            return await conn.QueryFirstOrDefaultAsync<string>(sql, new { dq = $"%{dq}%" });
        }
    }
}
