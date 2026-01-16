using System;
using System.Threading.Tasks;
using BotWorker.Domain.Models.BotMessages;
using BotWorker.Domain.Repositories;
using Dapper.Contrib.Extensions;
using Microsoft.Extensions.DependencyInjection;

namespace BotWorker.Domain.Entities
{
    [Table("Event")]
    public class BotEventLog
    {
        private static IBotEventLogRepository Repository => 
            BotMessage.ServiceProvider?.GetRequiredService<IBotEventLogRepository>() 
            ?? throw new InvalidOperationException("IBotEventLogRepository not registered");

        [Key]
        public long Id { get; set; }
        public long BotUin { get; set; }
        public long GroupId { get; set; }
        public string GroupName { get; set; } = string.Empty;
        public long UserId { get; set; }
        public string UserName { get; set; } = string.Empty;
        public string EventName { get; set; } = string.Empty;

        // 记录机器人事件
        public static int Append(BotMessage bm, string eventName)
        {
            return Append(bm.SelfId, eventName, bm.GroupId, bm.GroupName, bm.UserId, bm.Name);
        }

        // 记录机器人事件
        public static int Append(long botUin, string eventName, long groupId, string groupName, long userId, string userName)
        {
            return Repository.AppendAsync(botUin, eventName, groupId, groupName, userId, userName).GetAwaiter().GetResult();
        }
    }
}
