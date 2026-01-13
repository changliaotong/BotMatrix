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

        public async Task<List<Chunk>> SearchAsync(string query, long groupId = 0, int topK = 3)
        {
            if (string.IsNullOrWhiteSpace(query)) return new List<Chunk>();

            // --- Agentic RAG: Query Refinement (查询重写) ---
            string refinedQuery = await RefineQueryAsync(query);
            
            var allResults = new List<Chunk>();

            // 1. 从 KnowledgeBaseService 获取外部知识 (Legacy API)
            try
            {
                var kbResults = await _kbService.GetKnowledgesAsync(groupId, refinedQuery);
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

            // --- Agentic RAG: Self-Reflection (检索相关性自检) ---
            var filteredResults = await ReflectResultsAsync(query, distinctResults);

            return filteredResults.Take(topK).ToList();
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

        private async Task<List<Chunk>> ReflectResultsAsync(string query, List<Chunk> chunks)
        {
            if (chunks.Count == 0) return chunks;

            var tasks = chunks.Select(async chunk =>
            {
                try
                {
                    var prompt = $"你是一个知识库评估专家。请判断以下检索到的内容是否能回答或有助于回答用户的问题。只需要返回 YES 或 NO。\n用户提问：{query}\n检索内容：{chunk.Content}\n判断结果：";
                    var decision = await AIService.RawChatAsync(prompt);
                    var isRelevant = decision?.Trim().ToUpper().Contains("YES") ?? true;
                    if (!isRelevant)
                    {
                        Console.WriteLine($"[Agentic RAG] Self-Reflection: Filtered out chunk from {chunk.Source}");
                    }
                    return (Chunk: chunk, IsRelevant: isRelevant);
                }
                catch
                {
                    return (Chunk: chunk, IsRelevant: true); // 报错时默认保留
                }
            });

            var results = await Task.WhenAll(tasks);
            return results.Where(r => r.IsRelevant).Select(r => r.Chunk).ToList();
        }

        public async Task IndexDocumentAsync(string content, string source)
        {
            var splitter = new TextSplitter();
            var textChunks = splitter.Split(content);
            var chunksToSave = new List<Chunk>();

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

                chunksToSave.Add(new Chunk 
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


