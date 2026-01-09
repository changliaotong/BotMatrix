using System.Collections.Generic;
using System.Threading.Tasks;

namespace BotWorker.Agents.Interfaces
{
    public interface IKnowledgeBaseService
    {
        Task<string> BuildPrompt(string query, long groupId);
        Task<List<BotWorker.Agents.Plugins.KnowledgeBaseService.KnowledgeResult>?> GetKnowledgesAsync(long groupId, string question);
    }
}
