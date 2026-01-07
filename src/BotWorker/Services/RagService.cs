using System.Collections.Generic;
using System.Threading.Tasks;

namespace BotWorker.Services.Rag
{
    public interface IRagService
    {
        Task<List<Chunk>> SearchAsync(string query, int topK = 3);
        Task IndexDocumentAsync(string content, string source);
    }

    public class RagService : IRagService
    {
        private readonly List<Chunk> _store = new(); // 简化版内存存储

        public async Task<List<Chunk>> SearchAsync(string query, int topK = 3)
        {
            // 简化实现：返回前 topK 个片段
            return await Task.FromResult(_store.Take(topK).ToList());
        }

        public async Task IndexDocumentAsync(string content, string source)
        {
            var splitter = new TextSplitter();
            var textChunks = splitter.Split(content);
            foreach (var text in textChunks)
            {
                _store.Add(new Chunk { Content = text, Source = source });
            }
            await Task.CompletedTask;
        }
    }
}
