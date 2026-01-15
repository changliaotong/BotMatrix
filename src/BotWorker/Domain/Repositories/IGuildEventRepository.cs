using System.Threading.Tasks;
using BotWorker.Domain.Entities;

namespace BotWorker.Domain.Repositories
{
    public interface IGuildEventRepository : IBaseRepository<GuildEvent>
    {
        Task<int> AddAsync(GuildEvent guildEvent);
    }
}
