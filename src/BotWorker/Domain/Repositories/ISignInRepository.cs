using System.Threading.Tasks;
using BotWorker.Domain.Entities;

namespace BotWorker.Domain.Repositories
{
    public interface ISignInRepository : IBaseRepository<GroupSignIn>
    {
        Task<int> AddSignInAsync(long botUin, long groupId, long qq, string info, System.Data.IDbTransaction? trans = null);
        Task<long> GetTodaySignCountAsync(long groupId);
        Task<long> GetYesterdaySignCountAsync(long groupId);
        Task<long> GetUserMonthSignCountAsync(long groupId, long qq);
    }
}
