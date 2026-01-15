using System;
using System.Collections.Generic;
using System.Threading.Tasks;
using BotWorker.Modules.AI.Interfaces;
using BotWorker.Modules.AI.Models;
using Dapper;

using BotWorker.Infrastructure.Persistence.Repositories;

namespace BotWorker.Modules.AI.Repositories
{
    public class PostgresLLMCallLogRepository : BasePostgresRepository<LLMCallLog>, ILLMCallLogRepository
    {
        public PostgresLLMCallLogRepository(string? connectionString = null)
            : base("ai_llm_call_logs", connectionString)
        {
        }

        public async Task<IEnumerable<LLMCallLog>> GetByAgentIdAsync(long agentId)
        {
            using var conn = CreateConnection();
            return await conn.QueryAsync<LLMCallLog>(
                $"SELECT * FROM {_tableName} WHERE agent_id = @agentId", new { agentId });
        }

        public async Task<IEnumerable<LLMCallLog>> GetByModelIdAsync(long modelId)
        {
            using var conn = CreateConnection();
            return await conn.QueryAsync<LLMCallLog>(
                $"SELECT * FROM {_tableName} WHERE model_id = @modelId", new { modelId });
        }

        public async Task<long> AddAsync(LLMCallLog entity)
        {
            const string sql = @"
                INSERT INTO ai_llm_call_logs (
                    task_step_id, agent_id, model_id, prompt_tokens, 
                    completion_tokens, total_cost, latency_ms, is_success, 
                    raw_request, raw_response, created_at
                ) VALUES (
                    @TaskStepId, @AgentId, @ModelId, @PromptTokens, 
                    @CompletionTokens, @TotalCost, @LatencyMs, @IsSuccess, 
                    @RawRequest::jsonb, @RawResponse::jsonb, @CreatedAt
                ) RETURNING id";
            using var conn = CreateConnection();
            return await conn.ExecuteScalarAsync<long>(sql, entity);
        }

        public async Task<bool> UpdateAsync(LLMCallLog entity)
        {
            // 日志通常不更新
            return true;
        }
    }
}
