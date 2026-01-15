using System;
using System.Collections.Generic;
using System.Threading.Tasks;
using BotWorker.Domain.Repositories;
using Microsoft.Extensions.DependencyInjection;

namespace BotWorker.Domain.Entities
{
    public class WhiteList
    {
        private static IWhiteListRepository Repository => 
            BotMessage.ServiceProvider?.GetRequiredService<IWhiteListRepository>() 
            ?? throw new InvalidOperationException("IWhiteListRepository not registered");

        public long BotUin { get; set; }
        public long GroupId { get; set; }
        public string GroupName { get; set; } = string.Empty;
        public long UserId { get; set; }
        public string UserName { get; set; } = string.Empty;
        public long WhiteId { get; set; }
        public DateTime InsertDate { get; set; }

        // 加入白名单
        public static async Task<int> AppendWhiteListAsync(long botUin, long groupId, string groupName, long qq, string name, long qqWhite)
        {
            if (await Repository.IsExistsAsync(groupId, qqWhite))
                return 0;

            var whiteList = new WhiteList
            {
                BotUin = botUin,
                GroupId = groupId,
                GroupName = groupName,
                UserId = qq,
                UserName = name,
                WhiteId = qqWhite
            };

            return await Repository.AddAsync(whiteList);
        }

        public static int AppendWhiteList(long botUin, long groupId, string groupName, long qq, string name, long qqWhite)
        {
            return AppendWhiteListAsync(botUin, groupId, groupName, qq, name, qqWhite).GetAwaiter().GetResult();
        }

        public static async Task<bool> ExistsAsync(long groupId, long whiteId)
        {
            return await Repository.IsExistsAsync(groupId, whiteId);
        }

        public static async Task<int> DeleteAsync(long groupId, long whiteId)
        {
            return await Repository.DeleteAsync(groupId, whiteId);
        }
    }
}
