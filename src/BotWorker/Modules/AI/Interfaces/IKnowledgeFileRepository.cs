using System.Collections.Generic;
using System.Threading.Tasks;
using BotWorker.Modules.AI.Models;

namespace BotWorker.Modules.AI.Interfaces
{
    public interface IKnowledgeFileRepository
    {
        Task<IEnumerable<KnowledgeFile>> GetFilesByGroupAsync(long groupId);
        Task<long> AddAsync(KnowledgeFile file);
        Task<bool> UpdateAsync(KnowledgeFile file);
        Task<bool> MarkFileEmbeddedAsync(long fileId);
        Task<bool> MarkEmbeddingFailedAsync(long fileId, string error);
        Task<IEnumerable<KnowledgeFile>> GetPendingEmbeddingFilesAsync(long groupId);
        Task<bool> SetEnabledAsync(long fileId, bool enabled);
        Task<KnowledgeFile?> GetByIdAsync(long fileId);
        Task<bool> DeleteAsync(long fileId);
    }
}
