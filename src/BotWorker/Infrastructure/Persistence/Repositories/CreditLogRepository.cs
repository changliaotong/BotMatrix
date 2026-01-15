using System.Threading.Tasks;
using BotWorker.Domain.Entities;
using BotWorker.Domain.Repositories;
using Dapper;

namespace BotWorker.Infrastructure.Persistence.Repositories
{
    public class CreditLogRepository : BaseRepository<CreditLog>, ICreditLogRepository
    {
        public CreditLogRepository(string? connectionString = null) 
            : base("credit_log", connectionString ?? GlobalConfig.LogConnection)
        {
        }

        public async Task<int> AddLogAsync(long botUin, long groupId, string groupName, long userId, string userName, long creditAdd, long creditValue, string creditInfo, IDbTransaction? trans = null)
        {
            var log = new CreditLog
            {
                BotUin = botUin,
                GroupId = groupId,
                GroupName = groupName,
                UserId = userId,
                UserName = userName,
                CreditAdd = creditAdd,
                CreditValue = creditValue + creditAdd,
                CreditInfo = creditInfo,
                CreatedAt = System.DateTime.Now
            };

            return (int)await InsertAsync(log, trans);
        }

        public async Task<int> CreditCountAsync(long userId, string creditInfo, int second = 60)
        {
            string sql = $@"
                SELECT count(id) 
                FROM {_tableName} 
                WHERE user_id = @userId 
                AND credit_info LIKE @creditInfo 
                AND ABS(EXTRACT(EPOCH FROM (CURRENT_TIMESTAMP - created_at))) <= @second";
            
            using var conn = CreateConnection();
            return await conn.ExecuteScalarAsync<int>(sql, new { 
                userId, 
                creditInfo = $"%{creditInfo}%", 
                second 
            });
        }
    }
}
