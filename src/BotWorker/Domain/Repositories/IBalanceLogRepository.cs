using System.Threading.Tasks;
using System.Data;
using BotWorker.Domain.Entities;

namespace BotWorker.Domain.Repositories
{
    public interface IBalanceLogRepository : IBaseRepository<BalanceLog>
    {
        Task<int> AddLogAsync(long botUin, long groupId, string groupName, long userId, string userName, decimal balanceAdd, decimal balanceValue, string balanceInfo, IDbTransaction? trans = null);
    }
}
