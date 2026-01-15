using System.Threading.Tasks;
using System.Data;
using BotWorker.Domain.Entities;

namespace BotWorker.Domain.Repositories
{
    public interface IIncomeRepository : IBaseRepository<Income>
    {
        Task<long> AddAsync(Income income, IDbTransaction? trans = null);
        Task<float> GetTotalAsync(long userId);
        Task<float> GetTotalLastYearAsync(long userId);
        Task<bool> IsVipOnceAsync(long groupId);
        Task<int> GetClientLevelAsync(long userId);
        Task<string> GetLevelListAsync(long groupId);
        Task<string> GetLeverOrderAsync(long groupId, long userId);
        Task<string> GetStatAsync(string range);
    }
}
