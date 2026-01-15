using System;
using System.Threading.Tasks;
using BotWorker.Domain.Repositories;
using Microsoft.Extensions.DependencyInjection;

namespace BotWorker.Domain.Entities
{
    public class GroupEvent
    {
        private static IGroupEventRepository Repository => 
            BotMessage.ServiceProvider?.GetRequiredService<IGroupEventRepository>() 
            ?? throw new InvalidOperationException("IGroupEventRepository not registered");

        public int Id { get; set; }
        public long BotUin { get; set; }
        public long GroupId { get; set; }
        public string EventType { get; set; } = string.Empty;
        public string EventMsg { get; set; } = string.Empty;
        public DateTime InsertDate { get; set; }

        public static async Task<int> InsertEventAsync(long botUin, long groupId, string eventType, string eventMsg)
        {
            var groupEvent = new GroupEvent
            {
                BotUin = botUin,
                GroupId = groupId,
                EventType = eventType,
                EventMsg = eventMsg
            };
            return await Repository.AddAsync(groupEvent);
        }
    }
}
