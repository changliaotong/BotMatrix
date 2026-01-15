using System;
using System.Collections.Generic;
using System.Threading.Tasks;
using BotWorker.Infrastructure.Persistence.Repositories;
using BotWorker.Modules.AI.Interfaces;
using BotWorker.Modules.AI.Models;
using Dapper;

namespace BotWorker.Modules.AI.Repositories
{
    public class PostgresKnowledgeFileRepository : BasePostgresRepository<KnowledgeFile>, IKnowledgeFileRepository
    {
        public PostgresKnowledgeFileRepository(string? connectionString = null) 
            : base("ai_knowledge_files", connectionString)
        {
        }

        public async Task<IEnumerable<KnowledgeFile>> GetFilesByGroupAsync(long groupId)
        {
            const string sql = "SELECT * FROM ai_knowledge_files WHERE group_id = @groupId ORDER BY upload_time DESC";
            using var conn = CreateConnection();
            return await conn.QueryAsync<KnowledgeFile>(sql, new { groupId });
        }

        public async Task<long> AddAsync(KnowledgeFile file)
        {
            const string sql = @"
                INSERT INTO ai_knowledge_files (
                    group_id, file_name, display_name, description, storage_path, 
                    enabled, upload_time, file_hash, is_embedded, user_id
                ) VALUES (
                    @GroupId, @FileName, @DisplayName, @Description, @StoragePath, 
                    @Enabled, @UploadTime, @FileHash, @IsEmbedded, @UserId
                ) RETURNING id";
            using var conn = CreateConnection();
            return await conn.ExecuteScalarAsync<long>(sql, file);
        }

        public async Task<bool> UpdateAsync(KnowledgeFile file)
        {
            const string sql = @"
                UPDATE ai_knowledge_files SET 
                    display_name = @DisplayName, description = @Description, 
                    enabled = @Enabled, updated_at = CURRENT_TIMESTAMP
                WHERE id = @Id";
            using var conn = CreateConnection();
            return await conn.ExecuteAsync(sql, file) > 0;
        }

        public async Task<bool> MarkFileEmbeddedAsync(long fileId)
        {
            const string sql = @"
                UPDATE ai_knowledge_files SET 
                    is_embedded = true, embedded_time = CURRENT_TIMESTAMP, 
                    embedding_error = NULL, updated_at = CURRENT_TIMESTAMP 
                WHERE id = @fileId";
            using var conn = CreateConnection();
            return await conn.ExecuteAsync(sql, new { fileId }) > 0;
        }

        public async Task<bool> MarkEmbeddingFailedAsync(long fileId, string error)
        {
            const string sql = @"
                UPDATE ai_knowledge_files SET 
                    is_embedded = false, embedded_time = CURRENT_TIMESTAMP, 
                    embedding_error = @error, updated_at = CURRENT_TIMESTAMP 
                WHERE id = @fileId";
            using var conn = CreateConnection();
            return await conn.ExecuteAsync(sql, new { fileId, error }) > 0;
        }

        public async Task<IEnumerable<KnowledgeFile>> GetPendingEmbeddingFilesAsync(long groupId)
        {
            const string sql = "SELECT * FROM ai_knowledge_files WHERE group_id = @groupId AND is_embedded = false AND embedding_error IS NULL";
            using var conn = CreateConnection();
            return await conn.QueryAsync<KnowledgeFile>(sql, new { groupId });
        }

        public async Task<bool> SetEnabledAsync(long fileId, bool enabled)
        {
            const string sql = "UPDATE ai_knowledge_files SET enabled = @enabled, updated_at = CURRENT_TIMESTAMP WHERE id = @fileId";
            using var conn = CreateConnection();
            return await conn.ExecuteAsync(sql, new { fileId, enabled }) > 0;
        }

        public async Task<bool> DeleteAsync(long fileId)
        {
            using var conn = CreateConnection();
            return await conn.ExecuteAsync("DELETE FROM ai_knowledge_files WHERE id = @fileId", new { fileId }) > 0;
        }
    }
}
