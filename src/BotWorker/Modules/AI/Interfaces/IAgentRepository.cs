using System;
using System.Collections.Generic;
using System.Threading.Tasks;
using BotWorker.Modules.AI.Models;

namespace BotWorker.Modules.AI.Interfaces
{
    public interface IAgentRepository : IRepository<Agent, long>
    {
        Task<Agent?> GetByGuidAsync(Guid guid);
        Task<Agent?> GetByNameAsync(string name);
        Task<IEnumerable<Agent>> GetPublicAgentsAsync();
        Task<bool> IncrementUsedTimesAsync(long id, int increment = 1);
        Task<bool> IncrementSubscriptionCountAsync(long id, int increment = 1);
        Task<bool> ExistsByNameAndUserAsync(string name, long userId);
    }
}
