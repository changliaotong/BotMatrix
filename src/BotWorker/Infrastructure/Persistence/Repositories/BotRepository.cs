using System.Collections.Generic;
using System.Threading.Tasks;
using BotWorker.Domain.Entities;
using BotWorker.Domain.Repositories;
using Dapper;

namespace BotWorker.Infrastructure.Persistence.Repositories
{
    public class BotRepository : BaseRepository<BotInfo>, IBotRepository
    {
        protected override string KeyField => "bot_uin";

        public BotRepository(string? connectionString = null)
            : base("bot_info", connectionString ?? GlobalConfig.BaseInfoConnection)
        {
        }

        public async Task<bool> GetIsCreditAsync(long botUin)
        {
            return await GetValueAsync<bool>("is_credit", botUin);
        }

        public async Task<long> GetRobotAdminAsync(long botUin)
        {
            return await GetValueAsync<long>("admin_id", botUin);
        }

        public async Task<string> GetBotGuidAsync(long botUin)
        {
            return await GetValueAsync<string?>("bot_guid", botUin) ?? string.Empty;
        }

        public async Task<bool> IsRobotAsync(long qq)
        {
            const string sql = "SELECT COUNT(1) FROM bot_info WHERE valid NOT IN (0, 4, 5) AND bot_uin = @qq";
            using var conn = CreateConnection();
            return await conn.ExecuteScalarAsync<int>(sql, new { qq }) > 0;
        }
    }
}
