using System;
using System.Collections.Generic;
using System.Linq;
using System.Threading.Tasks;
using BotWorker.Modules.AI.Interfaces;
using BotWorker.Modules.AI.Rag;

namespace BotWorker.Modules.AI.Services
{
    public class MemoryRagStorage : IRagStorage
    {
        private readonly List<Chunk> _store = new();

        public Task<List<Chunk>> SearchAsync(string query, float[]? queryVector, int topK = 5)
        {
            if (queryVector != null && queryVector.Length > 0)
            {
                // Simple cosine similarity in memory
                var results = _store
                    .Where(c => c.Embedding != null)
                    .Select(c => new
                    {
                        Chunk = c,
                        Similarity = CosineSimilarity(queryVector, c.Embedding!)
                    })
                    .OrderByDescending(x => x.Similarity)
                    .Take(topK)
                    .Select(x => x.Chunk)
                    .ToList();
                
                if (results.Any()) return Task.FromResult(results);
            }

            // Fallback to keyword search
            var keywords = query.ToLower().Split(' ', StringSplitOptions.RemoveEmptyEntries);
            var keywordResults = _store
                .Select(chunk => new
                {
                    Chunk = chunk,
                    Score = keywords.Count(k => chunk.Content.ToLower().Contains(k))
                })
                .Where(x => x.Score > 0)
                .OrderByDescending(x => x.Score)
                .Take(topK)
                .Select(x => x.Chunk)
                .ToList();

            return Task.FromResult(keywordResults);
        }

        public Task SaveChunksAsync(List<Chunk> chunks)
        {
            foreach (var chunk in chunks)
            {
                var existing = _store.FirstOrDefault(c => c.Id == chunk.Id);
                if (existing != null)
                {
                    _store.Remove(existing);
                }
                _store.Add(chunk);
            }
            return Task.CompletedTask;
        }

        private float CosineSimilarity(float[] v1, float[] v2)
        {
            if (v1.Length != v2.Length) return 0;
            float dotProduct = 0;
            float l1 = 0;
            float l2 = 0;
            for (int i = 0; i < v1.Length; i++)
            {
                dotProduct += v1[i] * v2[i];
                l1 += v1[i] * v1[i];
                l2 += v2[i] * v2[i];
            }
            if (l1 == 0 || l2 == 0) return 0;
            return dotProduct / (MathF.Sqrt(l1) * MathF.Sqrt(l2));
        }
    }
}
