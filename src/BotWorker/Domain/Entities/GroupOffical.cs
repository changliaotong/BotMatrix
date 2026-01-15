using System;
using System.Threading.Tasks;
using BotWorker.Domain.Repositories;
using Microsoft.Extensions.DependencyInjection;

namespace BotWorker.Domain.Entities
{
    public class GroupOffical
    {
        private static IGroupOfficalRepository Repository => 
            BotMessage.ServiceProvider?.GetRequiredService<IGroupOfficalRepository>() 
            ?? throw new InvalidOperationException("IGroupOfficalRepository not registered");

        public long GroupId { get; set; }

        public static async Task<bool> IsOfficalAsync(long groupId)
        {
            return await Repository.IsOfficalAsync(groupId);
        }

        public static bool IsOffical(long groupId)
        {
            return IsOfficalAsync(groupId).GetAwaiter().GetResult();
        }
    }
}
