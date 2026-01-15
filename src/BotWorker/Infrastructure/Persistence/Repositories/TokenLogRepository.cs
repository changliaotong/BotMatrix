using System.Threading.Tasks;
using BotWorker.Domain.Entities;
using BotWorker.Domain.Repositories;
using Dapper;

namespace BotWorker.Infrastructure.Persistence.Repositories
{
    public class TokenLogRepository : BaseRepository<TokensLog>, ITokensLogRepository
    {
        public TokenLogRepository(string? connectionString = null) 
            : base("token_log", connectionString ?? GlobalConfig.LogConnection)
        {
        }

        public async Task<int> AddLogAsync(long botUin, long groupId, string groupName, long userId, string userName, long tokensAdd, long tokensValue, string tokensInfo, IDbTransaction? trans = null)
        {
            var log = new TokensLog
            {
                BotUin = botUin,
                GroupId = groupId,
                GroupName = groupName,
                UserId = userId,
                UserName = userName,
                TokensAdd = tokensAdd,
                TokensValue = tokensValue + tokensAdd,
                TokensInfo = tokensInfo,
                InsertDate = System.DateTime.Now
            };

            return (int)await InsertAsync(log, trans);
        }

        public async Task<long> GetDayTokensGroupAsync(long groupId, long userId)
        {
            string sql = $@"
                SELECT COALESCE(SUM(tokens_add), 0) 
                FROM {_tableName} 
                WHERE group_id = @groupId AND user_id = @userId 
                AND insert_date::date = CURRENT_DATE AND tokens_add < 0";
            
            using var conn = CreateConnection();
            return await conn.ExecuteScalarAsync<long>(sql, new { groupId, userId });
        }

        public async Task<long> GetDayTokensAsync(long userId)
        {
            string sql = $@"
                SELECT COALESCE(SUM(tokens_add), 0) 
                FROM {_tableName} 
                WHERE user_id = @userId 
                AND insert_date::date = CURRENT_DATE AND tokens_add < 0";
            
            using var conn = CreateConnection();
            return await conn.ExecuteScalarAsync<long>(sql, new { userId });
        }
    }
}
