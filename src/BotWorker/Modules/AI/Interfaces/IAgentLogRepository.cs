using System.Collections.Generic;
using System.Threading.Tasks;
using BotWorker.Modules.AI.Models;

namespace BotWorker.Modules.AI.Interfaces
{
    public interface IAgentLogRepository
    {
        Task<long> AddAsync(AgentLog log);
        Task<IEnumerable<AgentLog>> GetByUserIdAsync(long userId, int limit = 20);
        Task<IEnumerable<AgentLog>> GetByAgentIdAsync(long agentId, int limit = 20);
    }
}
