using System.Threading.Tasks;
using BotWorker.Domain.Entities;
using BotWorker.Domain.Models.BotMessages;

namespace BotWorker.Domain.Repositories
{
    public interface IBotLogRepository : IBaseRepository<BotLog>
    {
        Task<int> LogAsync(string info, string memo, BotMessage bm);
    }
}
