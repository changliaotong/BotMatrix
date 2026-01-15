using System.Threading.Tasks;
using BotWorker.Domain.Entities;

namespace BotWorker.Domain.Repositories
{
    public interface ICoinsLogRepository : IBaseRepository<CoinsLog>
    {
        Task<int> AddLogAsync(long botUin, long groupId, string groupName, long qq, string name, int coinsType, long coinsAdd, long coinsValue, string coinsInfo, IDbTransaction? trans = null);
    }
}
