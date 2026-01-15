using System.Threading.Tasks;
using BotWorker.Domain.Entities;

namespace BotWorker.Domain.Repositories
{
    public interface IBotHintsRepository : IBaseRepository<BotHints>
    {
        Task<string> GetHintAsync(string cmd);
    }
}
