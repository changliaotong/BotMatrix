using System;
using System.Collections.Generic;
using System.Threading.Tasks;
using Microsoft.Extensions.DependencyInjection;
using BotWorker.Modules.AI.Interfaces;
using BotWorker.Domain.Models.BotMessages;

namespace BotWorker.Modules.AI.Models
{
    public class AgentSubs
    {
        public long Id { get; set; }
        public long UserId { get; set; }
        public long AgentId { get; set; }
        public bool IsSub { get; set; }
        public DateTime CreatedAt { get; set; }
        public DateTime UpdatedAt { get; set; }

        private static IAgentSubscriptionRepository Repository => 
            BotMessage.ServiceProvider?.GetRequiredService<IAgentSubscriptionRepository>() 
            ?? throw new InvalidOperationException("IAgentSubscriptionRepository not registered");

        public static async Task<int> AppendAsync(long userId, long id, bool isSub = true)
        {
            return await Repository.SubscribeAsync(userId, id, isSub);
        }

        public static bool IsSub(long userId, Guid guid)
        {
            var agentId = Agent.GetId(guid);
            return Repository.IsSubscribedAsync(userId, agentId).GetAwaiter().GetResult();
        } 

        public static async Task<int> SubAsync(long userId, long id, bool isSub = true)
        {
            int i = await Repository.SubscribeAsync(userId, id, isSub);
            if (i == 0) return i;
            
            // Increment subscription count in agent table
            using var scope = BotMessage.ServiceProvider!.CreateScope();
            var agentRepo = scope.ServiceProvider.GetRequiredService<IAgentRepository>();
            await agentRepo.IncrementSubscriptionCountAsync(id, isSub ? 1 : -1);
            
            return i;
        }
    }
}
