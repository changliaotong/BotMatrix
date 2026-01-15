using System;
using System.Collections.Generic;
using System.Threading.Tasks;
using BotWorker.Infrastructure.Persistence.Repositories;
using BotWorker.Modules.AI.Interfaces;
using BotWorker.Modules.AI.Models;
using Dapper;

namespace BotWorker.Modules.AI.Repositories
{
    public class PostgresAgentTagRepository : BasePostgresRepository<AgentTag>, IAgentTagRepository
    {
        public PostgresAgentTagRepository(string? connectionString = null) 
            : base("ai_agent_tags", connectionString)
        {
        }

        public async Task<long> CreateTagAsync(AgentTag tag)
        {
            const string sql = @"
                INSERT INTO ai_agent_tags (name, description, owner_id, created_at, updated_at)
                VALUES (@Name, @Description, @UserId, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
                RETURNING id";
            using var conn = CreateConnection();
            return await conn.ExecuteScalarAsync<long>(sql, tag);
        }

        public async Task<bool> UpdateTagAsync(AgentTag tag)
        {
            const string sql = @"
                UPDATE ai_agent_tags SET 
                    name = @Name, description = @Description, updated_at = CURRENT_TIMESTAMP
                WHERE id = @Id";
            using var conn = CreateConnection();
            return await conn.ExecuteAsync(sql, tag) > 0;
        }

        public async Task<bool> DeleteTagAsync(long tagId)
        {
            using var conn = CreateConnection();
            return await conn.ExecuteAsync("DELETE FROM ai_agent_tags WHERE id = @tagId", new { tagId }) > 0;
        }

        public async Task<AgentTag?> GetTagByIdAsync(long tagId)
        {
            return await GetByIdAsync(tagId);
        }

        public async Task<AgentTag?> GetTagByNameAsync(string name)
        {
            using var conn = CreateConnection();
            return await conn.QueryFirstOrDefaultAsync<AgentTag>(
                "SELECT * FROM ai_agent_tags WHERE name = @name", new { name });
        }

        public async Task<IEnumerable<AgentTag>> GetAllTagsAsync()
        {
            return await GetAllAsync();
        }

        public async Task<bool> AddTagToAgentAsync(long agentId, long tagId)
        {
            const string sql = "INSERT INTO ai_agent_tag_relations (agent_id, tag_id) VALUES (@agentId, @tagId) ON CONFLICT DO NOTHING";
            using var conn = CreateConnection();
            return await conn.ExecuteAsync(sql, new { agentId, tagId }) > 0;
        }

        public async Task<bool> RemoveTagFromAgentAsync(long agentId, long tagId)
        {
            const string sql = "DELETE FROM ai_agent_tag_relations WHERE agent_id = @agentId AND tag_id = @tagId";
            using var conn = CreateConnection();
            return await conn.ExecuteAsync(sql, new { agentId, tagId }) > 0;
        }

        public async Task<IEnumerable<AgentTag>> GetTagsByAgentIdAsync(long agentId)
        {
            const string sql = @"
                SELECT t.* FROM ai_agent_tags t
                JOIN ai_agent_tag_relations r ON t.id = r.tag_id
                WHERE r.agent_id = @agentId";
            using var conn = CreateConnection();
            return await conn.QueryAsync<AgentTag>(sql, new { agentId });
        }
    }
}
