using System.Collections.Generic;
using System.Threading.Tasks;
using BotWorker.Infrastructure.Persistence.Repositories;
using BotWorker.Modules.AI.Interfaces;
using BotWorker.Modules.AI.Models;
using Dapper;

namespace BotWorker.Modules.AI.Repositories
{
    public class PostgresAgentLogRepository : BasePostgresRepository<AgentLog>, IAgentLogRepository
    {
        public PostgresAgentLogRepository(string? connectionString = null) 
            : base("ai_usage_logs", connectionString)
        {
        }

        public async Task<long> AddAsync(AgentLog log)
        {
            const string sql = @"
                INSERT INTO ai_usage_logs (
                    user_id, agent_id, model_name, input_tokens, output_tokens, 
                    duration_ms, status, error_message, 
                    guid, group_id, group_name, user_name, msg_id, 
                    question, answer, messages, credit, created_at
                ) VALUES (
                    @UserId, @AgentId, @ModelName, @InputTokens, @OutputTokens, 
                    @DurationMs, @Status, @ErrorMessage, 
                    @Guid, @GroupId, @GroupName, @UserName, @MsgId, 
                    @Question, @Answer, @Messages, @Credit, @CreatedAt
                ) RETURNING id";
            
            using var conn = CreateConnection();
            return await conn.ExecuteScalarAsync<long>(sql, log);
        }

        public async Task<IEnumerable<AgentLog>> GetByUserIdAsync(long userId, int limit = 20)
        {
            using var conn = CreateConnection();
            return await conn.QueryAsync<AgentLog>(
                $"SELECT * FROM {_tableName} WHERE user_id = @userId ORDER BY created_at DESC LIMIT @limit",
                new { userId, limit });
        }

        public async Task<IEnumerable<AgentLog>> GetByAgentIdAsync(long agentId, int limit = 20)
        {
            using var conn = CreateConnection();
            return await conn.QueryAsync<AgentLog>(
                $"SELECT * FROM {_tableName} WHERE agent_id = @agentId ORDER BY created_at DESC LIMIT @limit",
                new { agentId, limit });
        }
    }
}
