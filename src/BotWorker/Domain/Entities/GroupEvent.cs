using System;
using System.Threading.Tasks;
using BotWorker.Domain.Repositories;
using Microsoft.Extensions.DependencyInjection;

using BotWorker.Domain.Models.BotMessages;
using Dapper.Contrib.Extensions;

namespace BotWorker.Domain.Entities
{
    [Table("group_event")]
    public class GroupEvent
    {
        [Key]
        public int Id { get; set; }
        public long BotUin { get; set; }
        public long GroupId { get; set; }
        public string GroupName { get; set; } = string.Empty;
        public long UserId { get; set; }
        public string UserName { get; set; } = string.Empty;
        public string Message { get; set; } = string.Empty;
        public string EventType { get; set; } = string.Empty;
        public string EventMsg { get; set; } = string.Empty;
        public DateTime InsertDate { get; set; }

        public static async Task<int> AppendAsync(BotMessage bm, string eventType, string eventMsg)
        {
            var groupEvent = new GroupEvent
            {
                BotUin = bm.SelfId,
                GroupId = bm.RealGroupId,
                GroupName = bm.RealGroupName,
                UserId = bm.UserId,
                UserName = bm.Name,
                Message = bm.Message,
                EventType = eventType,
                EventMsg = eventMsg
            };
            return await bm.GroupEventRepository.AddAsync(groupEvent);
        }
    }
}
