using System.Collections.Generic;
using System.Threading.Tasks;
using BotWorker.Modules.AI.Models;

namespace BotWorker.Modules.AI.Interfaces
{
    public interface IAgentSubscriptionRepository
    {
        Task<bool> IsSubscribedAsync(long userId, long agentId);
        Task<int> SubscribeAsync(long userId, long agentId, bool isSub = true);
        Task<IEnumerable<long>> GetSubscribedAgentIdsAsync(long userId);
    }
}
