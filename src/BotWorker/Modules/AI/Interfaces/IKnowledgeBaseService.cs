using System.Collections.Generic;
using System.Threading.Tasks;

namespace BotWorker.Modules.AI.Interfaces
{
    public interface IKnowledgeBaseService
    {
        Task<string> BuildPrompt(string query, long groupId);
        Task<List<BotWorker.Modules.AI.Plugins.KnowledgeBaseService.KnowledgeResult>?> GetKnowledgesAsync(long groupId, string question);
    }
}
