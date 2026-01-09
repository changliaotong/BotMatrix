using System.Collections.Generic;
using System.Linq;
using System.Threading.Tasks;
using BotWorker.Modules.AI.Interfaces;
using BotWorker.Modules.AI.Rag;

namespace BotWorker.Modules.AI.Services
{
    public interface IRagService
    {
        Task<List<Chunk>> SearchAsync(string query, long groupId = 0, int topK = 3);
        Task<string> GetFormattedKnowledgeAsync(string query, long groupId = 0, int topK = 3);
        Task IndexDocumentAsync(string content, string source);
    }

    public class RagService : IRagService
    {
        private readonly IKnowledgeBaseService _kbService;
        private readonly List<Chunk> _store = new(); // 简化版内存存储

        public RagService(IKnowledgeBaseService kbService)
        {
            _kbService = kbService;
        }

        public async Task<string> GetFormattedKnowledgeAsync(string query, long groupId = 0, int topK = 3)
        {
            var chunks = await SearchAsync(query, groupId, topK);
            if (chunks == null || !chunks.Any()) return string.Empty;

            var sb = new System.Text.StringBuilder();
            sb.AppendLine("【参考知识库信息】:");
            foreach (var chunk in chunks)
            {
                sb.AppendLine($"[来源:{chunk.Source}] {chunk.Content}");
                sb.AppendLine();
            }
            return sb.ToString();
        }

        public async Task<List<Chunk>> SearchAsync(string query, long groupId = 0, int topK = 3)
        {
            if (string.IsNullOrWhiteSpace(query)) return new List<Chunk>();

            var allResults = new List<Chunk>();

            // 1. 从 KnowledgeBaseService 获取知识
            try
            {
                var kbResults = await _kbService.GetKnowledgesAsync(groupId, query);
                if (kbResults != null)
                {
                    allResults.AddRange(kbResults.Select(r => new Chunk
                    {
                        Content = r.Content,
                        Source = r.Source
                    }));
                }
            }
            catch
            {
                // 忽略 KB 服务错误，继续使用本地存储
            }

            // 2. 从本地存储搜索并合并
            var keywords = query.ToLower().Split(' ', System.StringSplitOptions.RemoveEmptyEntries);
            var localResults = _store
                .Select(chunk => new
                {
                    Chunk = chunk,
                    Score = keywords.Count(k => chunk.Content.ToLower().Contains(k))
                })
                .Where(x => x.Score > 0)
                .OrderByDescending(x => x.Score)
                .Select(x => x.Chunk)
                .ToList();

            allResults.AddRange(localResults);

            // 3. 去重并返回 TopK
            return allResults
                .GroupBy(x => x.Content)
                .Select(g => g.First())
                .Take(topK)
                .ToList();
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


