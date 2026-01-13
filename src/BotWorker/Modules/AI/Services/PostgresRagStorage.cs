using System;
using System.Collections.Generic;
using System.Data;
using System.Linq;
using System.Threading.Tasks;
using BotWorker.Common;
using BotWorker.Modules.AI.Interfaces;
using BotWorker.Modules.AI.Rag;
using Dapper;
using Npgsql;
using Pgvector;
using Pgvector.Dapper;
using Newtonsoft.Json;

namespace BotWorker.Modules.AI.Services
{
    public class PostgresRagStorage : IRagStorage
    {
        private class ChunkDto
        {
            public string id { get; set; } = string.Empty;
            public string content { get; set; } = string.Empty;
            public Vector? embedding { get; set; }
            public string? source { get; set; }
            public string? metadata { get; set; }
        }
        private readonly string _connectionString;
        private readonly int _vectorSize;
        private bool _isInitialized = false;

        public PostgresRagStorage(string? connectionString = null, int vectorSize = 1536)
        {
            _connectionString = connectionString ?? GlobalConfig.KnowledgeBaseConnection;
            _vectorSize = vectorSize;
            
            // Register Pgvector Type Handler for Dapper
            SqlMapper.AddTypeHandler(new VectorTypeHandler());
        }

        private async Task EnsureInitializedAsync()
        {
            if (_isInitialized) return;

            using var conn = new NpgsqlConnection(_connectionString);
            await conn.OpenAsync();

            // 1. Enable pgvector extension
            await conn.ExecuteAsync("CREATE EXTENSION IF NOT EXISTS vector");
            conn.ReloadTypes();

            // 2. Create knowledge_chunks table
            string createTableSql = $@"
                CREATE TABLE IF NOT EXISTS knowledge_chunks (
                    id TEXT PRIMARY KEY,
                    content TEXT NOT NULL,
                    embedding VECTOR({_vectorSize}),
                    source TEXT,
                    metadata JSONB,
                    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
                );
                
                CREATE INDEX IF NOT EXISTS idx_chunks_embedding ON knowledge_chunks 
                USING hnsw (embedding vector_cosine_ops);
            ";
            await conn.ExecuteAsync(createTableSql);

            _isInitialized = true;
        }

        public async Task<List<Chunk>> SearchAsync(string query, float[]? queryVector, int topK = 5)
        {
            await EnsureInitializedAsync();

            if (queryVector == null || queryVector.Length == 0)
            {
                // Fallback to keyword search if no vector provided
                return await KeywordSearchAsync(query, topK);
            }

            using var conn = new NpgsqlConnection(_connectionString);
            await conn.OpenAsync();

            var vector = new Vector(queryVector);
            
            // Cosine distance search (<=>)
            string sql = @"
                SELECT id, content, embedding, source, metadata 
                FROM knowledge_chunks 
                ORDER BY embedding <=> @vector 
                LIMIT @topK";

            var results = await conn.QueryAsync<ChunkDto>(sql, new { vector, topK });
            
            return results.Select(r => new Chunk
            {
                Id = r.id,
                Content = r.content,
                Embedding = r.embedding?.ToArray(),
                Source = r.source ?? string.Empty,
                Metadata = JsonConvert.DeserializeObject<Dictionary<string, object>>(r.metadata ?? "{}") ?? new Dictionary<string, object>()
            }).ToList();
        }

        private async Task<List<Chunk>> KeywordSearchAsync(string query, int topK)
        {
            using var conn = new NpgsqlConnection(_connectionString);
            await conn.OpenAsync();

            string sql = @"
                SELECT id, content, embedding, source, metadata 
                FROM knowledge_chunks 
                WHERE content ILIKE @query 
                LIMIT @topK";

            var results = await conn.QueryAsync<ChunkDto>(sql, new { query = $"%{query}%", topK });

            return results.Select(r => new Chunk
            {
                Id = r.id,
                Content = r.content,
                Embedding = r.embedding?.ToArray(),
                Source = r.source ?? string.Empty,
                Metadata = JsonConvert.DeserializeObject<Dictionary<string, object>>(r.metadata ?? "{}") ?? new Dictionary<string, object>()
            }).ToList();
        }

        public async Task SaveChunksAsync(List<Chunk> chunks)
        {
            await EnsureInitializedAsync();

            using var conn = new NpgsqlConnection(_connectionString);
            await conn.OpenAsync();

            using var trans = await conn.BeginTransactionAsync();

            try
            {
                string sql = @"
                    INSERT INTO knowledge_chunks (id, content, embedding, source, metadata)
                    VALUES (@Id, @Content, @Embedding, @Source, @Metadata)
                    ON CONFLICT (id) DO UPDATE SET
                        content = EXCLUDED.content,
                        embedding = EXCLUDED.embedding,
                        source = EXCLUDED.source,
                        metadata = EXCLUDED.metadata";

                foreach (var chunk in chunks)
                {
                    await conn.ExecuteAsync(sql, new
                    {
                        chunk.Id,
                        chunk.Content,
                        Embedding = chunk.Embedding != null ? new Vector(chunk.Embedding) : null,
                        chunk.Source,
                        Metadata = JsonConvert.SerializeObject(chunk.Metadata)
                    }, trans);
                }

                await trans.CommitAsync();
            }
            catch
            {
                await trans.RollbackAsync();
                throw;
            }
        }
    }
}
