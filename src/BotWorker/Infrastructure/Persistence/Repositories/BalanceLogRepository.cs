using System.Data;
using System.Threading.Tasks;
using BotWorker.Domain.Entities;
using BotWorker.Domain.Repositories;
using Dapper;

namespace BotWorker.Infrastructure.Persistence.Repositories
{
    public class BalanceLogRepository : BaseRepository<BalanceLog>, IBalanceLogRepository
    {
        public BalanceLogRepository(string? connectionString = null) 
            : base("Balance", connectionString ?? GlobalConfig.BaseInfoConnection)
        {
        }

        public async Task<int> AddLogAsync(long botUin, long groupId, string groupName, long userId, string userName, decimal balanceAdd, decimal balanceValue, string balanceInfo, IDbTransaction? trans = null)
        {
            var log = new BalanceLog
            {
                BotUin = botUin,
                GroupId = groupId,
                GroupName = groupName,
                UserId = userId,
                UserName = userName,
                BalanceAdd = balanceAdd,
                BalanceValue = balanceValue,
                BalanceInfo = balanceInfo,
                InsertDate = System.DateTime.Now
            };

            return (int)await InsertAsync(log, trans);
        }
    }
}
