using System.Collections.Generic;
using System.Threading.Tasks;
using BotWorker.Modules.AI.Models;

namespace BotWorker.Modules.AI.Interfaces
{
    public interface ILLMCallLogRepository : IRepository<LLMCallLog, long>
    {
        Task<IEnumerable<LLMCallLog>> GetByAgentIdAsync(long agentId);
        Task<IEnumerable<LLMCallLog>> GetByModelIdAsync(long modelId);
    }
}
