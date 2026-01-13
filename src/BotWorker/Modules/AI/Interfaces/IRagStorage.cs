using System.Collections.Generic;
using System.Threading.Tasks;
using BotWorker.Modules.AI.Rag;

namespace BotWorker.Modules.AI.Interfaces
{
    public interface IRagStorage
    {
        Task<List<BotWorker.Modules.AI.Rag.Chunk>> SearchAsync(string query, float[]? queryVector, int topK = 5);
        Task SaveChunksAsync(List<BotWorker.Modules.AI.Rag.Chunk> chunks);
    }
}
