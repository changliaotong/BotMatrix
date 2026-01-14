using System;
using System.Collections.Generic;
using System.Data;
using System.Linq;
using System.Text.Json;
using System.Threading.Tasks;
using BotWorker.Common;
using BotWorker.Modules.AI.Interfaces;
using BotWorker.Modules.AI.Models;
using Dapper;
using Npgsql;
using Pgvector;
using Pgvector.Dapper;

namespace BotWorker.Modules.AI.Services
{
    public class PostgresRagStorage : IRagStorage
    {
        private readonly string _connectionString;
        private readonly int _vectorSize;
        private readonly NpgsqlDataSource _dataSource;

        public PostgresRagStorage(string? connectionString = null, int vectorSize = 1536)
        {
            _connectionString = connectionString ?? GlobalConfig.KnowledgeBaseConnection;
            _vectorSize = vectorSize;
            
            var dataSourceBuilder = new NpgsqlDataSourceBuilder(_connectionString);
            dataSourceBuilder.UseVector();
            _dataSource = dataSourceBuilder.Build();
            
            // 显式使用命名空间避免 CS0234
            SqlMapper.AddTypeHandler(new Pgvector.Dapper.VectorTypeHandler());
        }

        private async Task<NpgsqlConnection> CreateConnectionAsync()
        {
            return await _dataSource.OpenConnectionAsync();
        }

        public async Task EnsureInitializedAsync()
        {
            using var conn = await CreateConnectionAsync();
            
            // 确保 pgvector 扩展存在
            await conn.ExecuteAsync("CREATE EXTENSION IF NOT EXISTS vector");
            
            // 确保知识库相关表存在 (如果 init 脚本没运行)
            await conn.ExecuteAsync(@"
                CREATE TABLE IF NOT EXISTS knowledge_bases (
                    id BIGSERIAL PRIMARY KEY,
                    name VARCHAR(100) NOT NULL,
                    description TEXT,
                    user_id BIGINT,
                    is_public BOOLEAN DEFAULT FALSE,
                    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
                    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
                )");

            await conn.ExecuteAsync($@"
                CREATE TABLE IF NOT EXISTS knowledge_chunks (
                    id BIGSERIAL PRIMARY KEY,
                    doc_id BIGINT,
                    content TEXT NOT NULL,
                    embedding VECTOR({_vectorSize}),
                    metadata JSONB DEFAULT '{{}}',
                    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
                )");
        }

        public async Task<List<BotWorker.Modules.AI.Rag.Chunk>> SearchAsync(string query, float[]? queryVector, int topK = 5)
        {
            if (queryVector == null || queryVector.Length == 0)
            {
                return new List<BotWorker.Modules.AI.Rag.Chunk>();
            }

            using var conn = await CreateConnectionAsync();
            
            // 使用 pgvector 的余弦距离进行检索 (<=>)
            const string sql = @"
                SELECT id, doc_id as DocId, content, metadata, created_at as CreatedAt,
                       1 - (embedding <=> @vector) as Score
                FROM knowledge_chunks
                ORDER BY embedding <=> @vector
                LIMIT @limit";

            var vector = new Vector(queryVector);
            var results = await conn.QueryAsync<dynamic>(sql, new { vector, limit = topK });
            
            return results.Select(r => new BotWorker.Modules.AI.Rag.Chunk
            {
                Id = r.id.ToString(),
                Content = r.content,
                Embedding = queryVector,
                Source = r.metadata != null ? (JsonSerializer.Deserialize<Dictionary<string, object>>(r.metadata.ToString())?.GetValueOrDefault("source")?.ToString() ?? "") : ""
            }).ToList();
        }

        public async Task SaveChunksAsync(List<BotWorker.Modules.AI.Rag.Chunk> chunks)
        {
            using var conn = await CreateConnectionAsync();
            using var transaction = conn.BeginTransaction();

            try
            {
                const string sql = @"
                    INSERT INTO knowledge_chunks (doc_id, content, embedding, metadata)
                    VALUES (@DocId, @Content, @Embedding, @MetadataJson::jsonb)";

                foreach (var chunk in chunks)
                {
                    if (chunk.Embedding == null || chunk.Embedding.Length == 0)
                    {
                        continue;
                    }

                    var docId = chunk.Metadata.ContainsKey("doc_id") ? Convert.ToInt64(chunk.Metadata["doc_id"]) : 0;
                    var metadataJson = JsonSerializer.Serialize(chunk.Metadata);
                    
                    await conn.ExecuteAsync(sql, new { 
                        DocId = docId, 
                        Content = chunk.Content, 
                        Embedding = new Vector(chunk.Embedding),
                        MetadataJson = metadataJson
                    }, transaction);
                }
                transaction.Commit();
            }
            catch
            {
                transaction.Rollback();
                throw;
            }
        }

        public async Task DeleteChunksAsync(long groupId)
        {
            using var conn = await CreateConnectionAsync();
            await conn.ExecuteAsync("DELETE FROM knowledge_chunks WHERE doc_id = @groupId", new { groupId });
        }
    }
}
