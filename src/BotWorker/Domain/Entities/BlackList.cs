using System;
using System.Collections.Generic;
using System.Threading.Tasks;
using BotWorker.Domain.Repositories;
using Microsoft.Extensions.DependencyInjection;

namespace BotWorker.Domain.Entities
{
    public class BlackList
    {
        private static IBlackListRepository Repository => 
            BotMessage.ServiceProvider?.GetRequiredService<IBlackListRepository>() 
            ?? throw new InvalidOperationException("IBlackListRepository not registered");

        public long BotUin { get; set; }
        public long GroupId { get; set; }
        public string GroupName { get; set; } = string.Empty;
        public long UserId { get; set; }
        public string UserName { get; set; } = string.Empty;
        public long BlackId { get; set; }
        public string BlackInfo { get; set; } = string.Empty;
        public DateTime InsertDate { get; set; }

        public const string regexBlack = @"^(?<cmdName>(取消|解除|删除)?(黑名单|拉黑|加黑|删黑))(?<cmdPara>([ ]*(\[?@:?)?[1-9]+\d*(\]?))+)$";

        public static async Task<IEnumerable<long>> GetSystemBlackListAsync()
        {
            return await Repository.GetSystemBlackListAsync();
        }

        public static bool IsSystemBlack(long userId) => IsSystemBlackAsync(userId).GetAwaiter().GetResult();

        public static async Task<bool> IsSystemBlackAsync(long userId)
        {
            return await Repository.IsExistsAsync(BotInfo.GroupIdDef, userId);
        }

        public static int AddBlackList(long botUin, long groupId, string groupName, long qq, string name, long blackQQ, string blackInfo)
            => AddBlackListAsync(botUin, groupId, groupName, qq, name, blackQQ, blackInfo).GetAwaiter().GetResult();

        // 加入黑名单
        public static async Task<int> AddBlackListAsync(long botUin, long groupId, string groupName, long qq, string name, long blackQQ, string blackInfo)
        {
            if (await Repository.IsExistsAsync(groupId, blackQQ))
                return 0;

            var blackList = new BlackList
            {
                BotUin = botUin,
                GroupId = groupId,
                GroupName = groupName,
                UserId = qq,
                UserName = name,
                BlackId = blackQQ,
                BlackInfo = blackInfo
            };

            return await Repository.AddAsync(blackList);
        }

        public static async Task<bool> ExistsAsync(long groupId, long blackId)
        {
            return await Repository.IsExistsAsync(groupId, blackId);
        }

        public static async Task<int> DeleteAsync(long groupId, long blackId)
        {
            return await Repository.DeleteAsync(groupId, blackId);
        }

        /// <summary>
        /// 清空指定群组的黑名单
        /// </summary>
        public static int ClearGroupBlacklist(long groupId) => ClearGroupBlacklistAsync(groupId).GetAwaiter().GetResult();

        public static async Task<int> ClearGroupBlacklistAsync(long groupId)
        {
            return await Repository.ClearGroupAsync(groupId);
        }
    }
}
