using System;
using System.Collections.Generic;
using System.Threading.Tasks;
using BotWorker.Infrastructure.Persistence.Repositories;
using BotWorker.Modules.AI.Interfaces;
using BotWorker.Modules.AI.Models;
using Dapper;

namespace BotWorker.Modules.AI.Repositories
{
    public class PostgresAgentSubscriptionRepository : BasePostgresRepository<AgentSubs>, IAgentSubscriptionRepository
    {
        public PostgresAgentSubscriptionRepository(string? connectionString = null) 
            : base("ai_agent_subscriptions", connectionString)
        {
        }

        public async Task<bool> IsSubscribedAsync(long userId, long agentId)
        {
            const string sql = "SELECT is_sub FROM ai_agent_subscriptions WHERE user_id = @userId AND agent_id = @agentId";
            using var conn = CreateConnection();
            return await conn.ExecuteScalarAsync<bool>(sql, new { userId, agentId });
        }

        public async Task<int> SubscribeAsync(long userId, long agentId, bool isSub = true)
        {
            const string sql = @"
                INSERT INTO ai_agent_subscriptions (user_id, agent_id, is_sub, updated_at)
                VALUES (@userId, @agentId, @isSub, CURRENT_TIMESTAMP)
                ON CONFLICT (user_id, agent_id) 
                DO UPDATE SET is_sub = EXCLUDED.is_sub, updated_at = CURRENT_TIMESTAMP";
            
            using var conn = CreateConnection();
            return await conn.ExecuteAsync(sql, new { userId, agentId, isSub });
        }

        public async Task<IEnumerable<long>> GetSubscribedAgentIdsAsync(long userId)
        {
            const string sql = "SELECT agent_id FROM ai_agent_subscriptions WHERE user_id = @userId AND is_sub = true";
            using var conn = CreateConnection();
            return await conn.QueryAsync<long>(sql, new { userId });
        }
    }
}
