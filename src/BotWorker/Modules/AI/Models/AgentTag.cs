using System;
using System.Collections.Generic;
using System.Threading.Tasks;
using Microsoft.Extensions.DependencyInjection;
using BotWorker.Modules.AI.Interfaces;
using BotWorker.Domain.Models.BotMessages;

namespace BotWorker.Modules.AI.Models
{
    public class AgentTag
    {
        public long Id { get; set; }
        public string Name { get; set; } = string.Empty;
        public string Description { get; set; } = string.Empty;    
        public long UserId { get; set; }
        public DateTime CreatedAt { get; set; } 
        public DateTime UpdatedAt { get; set; } 

        private static IAgentTagRepository Repository => 
            BotMessage.ServiceProvider?.GetRequiredService<IAgentTagRepository>() 
            ?? throw new InvalidOperationException("IAgentTagRepository not registered");

        public static async Task<long> AddAsync(AgentTag tag)
        {
            return await Repository.CreateTagAsync(tag);
        }
    }

    public class AgentTags
    {
        public long AgentId { get; set; } = 0;
        public long TagId { get; set; } = 0;
        public long UserId { get; set; } = 0;
        public DateTime CreatedAt { get; set; } = DateTime.MinValue;

        private static IAgentTagRepository Repository => 
            BotMessage.ServiceProvider?.GetRequiredService<IAgentTagRepository>() 
            ?? throw new InvalidOperationException("IAgentTagRepository not registered");

        public static async Task<bool> AddTagToAgentAsync(long agentId, long tagId)
        {
            return await Repository.AddTagToAgentAsync(agentId, tagId);
        }

        public static async Task<IEnumerable<AgentTag>> GetTagsByAgentIdAsync(long agentId)
        {
            return await Repository.GetTagsByAgentIdAsync(agentId);
        }
    }
}
