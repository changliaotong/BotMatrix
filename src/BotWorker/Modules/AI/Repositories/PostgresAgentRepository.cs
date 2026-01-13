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

namespace BotWorker.Modules.AI.Repositories
{
    public class PostgresAgentRepository : BasePostgresRepository<Agent>, IAgentRepository
    {
        public PostgresAgentRepository(string? connectionString = null)
            : base("ai_agents", connectionString)
        {
        }

        public async Task<Agent?> GetByGuidAsync(Guid guid)
        {
            using var conn = CreateConnection();
            var agent = await conn.QueryFirstOrDefaultAsync<Agent>(
                $"SELECT * FROM {_tableName} WHERE guid = @guid", new { guid });
            return MapAgent(agent);
        }

        public async Task<Agent?> GetByNameAsync(string name)
        {
            using var conn = CreateConnection();
            var agent = await conn.QueryFirstOrDefaultAsync<Agent>(
                $"SELECT * FROM {_tableName} WHERE name = @name AND is_public = true ORDER BY id DESC LIMIT 1", new { name });
            return MapAgent(agent);
        }

        public async Task<IEnumerable<Agent>> GetPublicAgentsAsync()
        {
            using var conn = CreateConnection();
            var agents = await conn.QueryAsync<Agent>($"SELECT * FROM {_tableName} WHERE is_public = true ORDER BY id DESC");
            return agents.Select(MapAgent).Where(a => a != null)!;
        }

        public override async Task<Agent?> GetByIdAsync(long id)
        {
            var agent = await base.GetByIdAsync(id);
            return MapAgent(agent);
        }

        public override async Task<IEnumerable<Agent>> GetAllAsync()
        {
            var agents = await base.GetAllAsync();
            return agents.Select(MapAgent).Where(a => a != null)!;
        }

        public async Task<long> AddAsync(Agent agent)
        {
            const string sql = @"
                INSERT INTO ai_agents (
                    guid, name, description, system_prompt, user_prompt_template, tags, model_id, config, owner_id, is_public
                ) VALUES (
                    @Guid, @Name, @Description, @SystemPrompt, @UserPromptTemplate, @Tags::jsonb, @ModelId, @Config::jsonb, @OwnerId, @IsPublic
                ) RETURNING id";
            
            using var conn = CreateConnection();
            return await conn.ExecuteScalarAsync<long>(sql, agent);
        }

        public async Task<bool> UpdateAsync(Agent agent)
        {
            const string sql = @"
                UPDATE ai_agents SET 
                    name = @Name, description = @Description, system_prompt = @SystemPrompt, 
                    user_prompt_template = @UserPromptTemplate, tags = @Tags::jsonb, 
                    model_id = @ModelId, config = @Config::jsonb, owner_id = @OwnerId, is_public = @IsPublic
                WHERE id = @Id";
            
            using var conn = CreateConnection();
            return await conn.ExecuteAsync(sql, agent) > 0;
        }

        public async Task<bool> IncrementUsedTimesAsync(long id, int increment = 1)
        {
            // 可以在这里实现，或者如果表中没有这个字段，则记录到统计表
            return true; 
        }

        public async Task<bool> IncrementSubscriptionCountAsync(long id, int increment = 1)
        {
            return true;
        }

        public async Task<bool> ExistsByNameAndUserAsync(string name, long ownerId)
        {
            using var conn = CreateConnection();
            return await conn.ExecuteScalarAsync<int>(
                $"SELECT COUNT(1) FROM {_tableName} WHERE name = @name AND owner_id = @ownerId", 
                new { name, ownerId }) > 0;
        }

        private Agent? MapAgent(Agent? agent)
        {
            return agent;
        }
    }
}
