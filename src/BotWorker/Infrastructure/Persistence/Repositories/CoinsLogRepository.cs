using System.Threading.Tasks;
using BotWorker.Infrastructure.Persistence.Repositories;
using BotWorker.Domain.Repositories;
using BotWorker.Domain.Entities;
using Dapper;

namespace BotWorker.Infrastructure.Persistence.Repositories
{
    public class CoinsLogRepository : BaseRepository<CoinsLog>, ICoinsLogRepository
    {
        public CoinsLogRepository(string? connectionString = null) 
            : base("coins", connectionString ?? GlobalConfig.LogConnection)
        {
        }

        public async Task<int> AddLogAsync(long botUin, long groupId, string groupName, long qq, string name, int coinsType, long coinsAdd, long coinsValue, string coinsInfo, IDbTransaction? trans = null)
        {
            var log = new CoinsLog
            {
                BotUin = botUin,
                GroupId = groupId,
                GroupName = groupName,
                UserId = qq,
                UserName = name,
                CoinsType = coinsType,
                CoinsAdd = coinsAdd,
                CoinsValue = coinsValue + coinsAdd,
                CoinsInfo = coinsInfo,
                InsertDate = System.DateTime.Now
            };

            return (int)await InsertAsync(log, trans);
        }
    }
}
