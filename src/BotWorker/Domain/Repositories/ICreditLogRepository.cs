using System.Threading.Tasks;
using BotWorker.Domain.Entities;

namespace BotWorker.Domain.Repositories
{
    public interface ICreditLogRepository : IBaseRepository<CreditLog>
    {
        Task<int> AddLogAsync(long botUin, long groupId, string groupName, long qq, string name, long creditAdd, long creditValue, string creditInfo);
        Task<int> CreditCountAsync(long userId, string creditInfo, int second = 60);
    }
}
