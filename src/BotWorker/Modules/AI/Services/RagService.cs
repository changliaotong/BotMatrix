using System.Collections.Generic;
using System.Linq;
using System.Threading.Tasks;
using BotWorker.Modules.AI.Interfaces;
using BotWorker.Modules.AI.Rag;

namespace BotWorker.Modules.AI.Services
{
    public interface IRagService
    {
        Task<List<BotWorker.Modules.AI.Rag.Chunk>> SearchAsync(string query, long groupId = 0, int topK = 3);
        Task<string> GetFormattedKnowledgeAsync(string query, long groupId = 0, int topK = 3);
        Task IndexDocumentAsync(string content, string source);
    }

    public class RagService : IRagService
    {
        private readonly IKnowledgeBaseService _kbService;
        private readonly IServiceProvider _serviceProvider;
        private readonly IRagStorage _storage;

        public RagService(IKnowledgeBaseService kbService, IServiceProvider serviceProvider)
        {
            _kbService = kbService;
            _serviceProvider = serviceProvider;
            
            // 优先使用 PostgreSQL 存储，如果没有配置则回退到内存存储
            if (!string.IsNullOrEmpty(GlobalConfig.KnowledgeBaseConnection))
            {
                _storage = new PostgresRagStorage();
            }
            else
            {
                _storage = new MemoryRagStorage();
            }
        }

        private IAIService AIService => _serviceProvider.GetRequiredService<IAIService>();

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

        public async Task<List<BotWorker.Modules.AI.Rag.Chunk>> SearchAsync(string query, long groupId = 0, int topK = 3)
        {
            if (string.IsNullOrWhiteSpace(query)) return new List<BotWorker.Modules.AI.Rag.Chunk>();

            // --- Agentic RAG: Query Refinement (查询重写) ---
            string refinedQuery = await RefineQueryAsync(query);
            
            var allResults = new List<BotWorker.Modules.AI.Rag.Chunk>();

            // 1. 从 KnowledgeBaseService 获取外部知识 (Legacy API)
            try
            {
                var kbResults = await _kbService.GetKnowledgesAsync(groupId, refinedQuery);
                if (kbResults != null)
                {
                    allResults.AddRange(kbResults.Select(r => new BotWorker.Modules.AI.Rag.Chunk
                    {
                        Content = r.Content,
                        Source = r.Source
                    }));
                }
            }
            catch
            {
                // 忽略 KB 服务错误
            }

            // 2. 从本地存储搜索 (向量搜索 + 关键词搜索)
            float[]? queryVector = null;
            try
            {
                // 尝试生成查询向量
                queryVector = await AIService.GenerateEmbeddingAsync(refinedQuery);
            }
            catch (Exception ex)
            {
                Console.WriteLine($"[RagService] Failed to generate embedding: {ex.Message}");
            }

            var localResults = await _storage.SearchAsync(refinedQuery, queryVector, topK * 2);
            allResults.AddRange(localResults);

            // 3. 去重
            var distinctResults = allResults
                .GroupBy(x => x.Content)
                .Select(g => g.First())
                .ToList();

            // 4. --- Agentic RAG: Re-ranking (重排序) ---
            var rerankedResults = await RerankResultsAsync(refinedQuery, distinctResults, topK);

            return rerankedResults;
        }

        private async Task<string> RefineQueryAsync(string query)
        {
            try
            {
                var prompt = $"你是一个搜索专家。请将以下用户提问改写为一个或多个最适合在知识库中进行向量搜索 and 关键词搜索的关键词或短语。只需要返回改写后的内容，不要有任何解释。\n提问：{query}\n改写结果：";
                var refined = await AIService.RawChatAsync(prompt);
                if (!string.IsNullOrWhiteSpace(refined))
                {
                    Console.WriteLine($"[Agentic RAG] Query Refined: '{query}' -> '{refined.Trim()}'");
                    return refined.Trim();
                }
            }
            catch (Exception ex)
            {
                Console.WriteLine($"[Agentic RAG] Query Refinement failed: {ex.Message}");
            }
            return query;
        }

        private async Task<List<BotWorker.Modules.AI.Rag.Chunk>> RerankResultsAsync(string query, List<BotWorker.Modules.AI.Rag.Chunk> chunks, int topK)
        {
            if (chunks.Count <= topK) return chunks;

            // 这里可以调用 LLM 进行重排序，或者使用简单的评分机制
            // 目前先返回前 topK 个
            return chunks.Take(topK).ToList();
        }

        public async Task IndexDocumentAsync(string content, string source)
        {
            var splitter = new TextSplitter();
            var textChunks = splitter.Split(content);
            var chunksToSave = new List<BotWorker.Modules.AI.Rag.Chunk>();

            foreach (var text in textChunks)
            {
                float[]? embedding = null;
                try
                {
                    embedding = await AIService.GenerateEmbeddingAsync(text);
                }
                catch (Exception ex)
                {
                    Console.WriteLine($"[RagService] Failed to generate embedding for indexing: {ex.Message}");
                }

                chunksToSave.Add(new BotWorker.Modules.AI.Rag.Chunk 
                { 
                    Content = text, 
                    Source = source,
                    Embedding = embedding,
                    Metadata = new Dictionary<string, object> { { "indexed_at", DateTime.UtcNow } }
                });
            }

            await _storage.SaveChunksAsync(chunksToSave);
        }
    }
}


