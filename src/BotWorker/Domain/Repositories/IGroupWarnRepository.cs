using System.Threading.Tasks;
using BotWorker.Domain.Entities;

namespace BotWorker.Domain.Repositories
{
    public interface IGroupWarnRepository : IBaseRepository<GroupWarn>
    {
        Task<long> CountByGroupAndUserAsync(long groupId, long userId);
        Task<int> DeleteByGroupAndUserAsync(long groupId, long userId);
    }
}
