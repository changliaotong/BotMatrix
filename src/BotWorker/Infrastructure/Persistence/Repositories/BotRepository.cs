using System;
using System.Collections.Generic;
using System.Threading.Tasks;
using BotWorker.Domain.Entities;
using BotWorker.Domain.Repositories;
using Dapper;

namespace BotWorker.Infrastructure.Persistence.Repositories
{
    public class BotRepository : BaseRepository<BotInfo>, IBotRepository
    {
        public BotRepository(string? connectionString = null)
            : base("bot_info", connectionString ?? GlobalConfig.BaseInfoConnection)
        {
        }

        public async Task<BotInfo?> GetByBotUinAsync(long botUin)
        {
            string sql = $"SELECT * FROM {_tableName} WHERE bot_uin = @botUin";
            using var conn = CreateConnection();
            return await conn.QueryFirstOrDefaultAsync<BotInfo>(sql, new { botUin });
        }

        public async Task<bool> IsRobotAsync(long userId)
        {
            string sql = $"SELECT COUNT(1) FROM {_tableName} WHERE bot_uin = @userId";
            using var conn = CreateConnection();
            return await conn.ExecuteScalarAsync<int>(sql, new { userId }) > 0;
        }
    }
}
