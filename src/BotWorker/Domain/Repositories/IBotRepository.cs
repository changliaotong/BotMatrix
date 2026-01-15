using System.Collections.Generic;
using System.Threading.Tasks;
using BotWorker.Domain.Entities;

namespace BotWorker.Domain.Repositories
{
    public interface IBotRepository : IBaseRepository<BotInfo>
    {
        Task<bool> GetIsCreditAsync(long botUin);
        Task<long> GetRobotAdminAsync(long botUin);
        Task<string> GetBotGuidAsync(long botUin);
        Task<bool> IsRobotAsync(long qq);
    }
}
