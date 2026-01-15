using System.Collections.Generic;
using System.Threading.Tasks;
using BotWorker.Modules.AI.Models;

namespace BotWorker.Modules.AI.Interfaces
{
    public interface IAgentTagRepository
    {
        Task<long> CreateTagAsync(AgentTag tag);
        Task<bool> UpdateTagAsync(AgentTag tag);
        Task<bool> DeleteTagAsync(long tagId);
        Task<AgentTag?> GetTagByIdAsync(long tagId);
        Task<AgentTag?> GetTagByNameAsync(string name);
        Task<IEnumerable<AgentTag>> GetAllTagsAsync();
        
        Task<bool> AddTagToAgentAsync(long agentId, long tagId);
        Task<bool> RemoveTagFromAgentAsync(long agentId, long tagId);
        Task<IEnumerable<AgentTag>> GetTagsByAgentIdAsync(long agentId);
    }
}
