using System;
using System.Threading.Tasks;
using BotWorker.Domain.Repositories;
using Dapper.Contrib.Extensions;
using Microsoft.Extensions.DependencyInjection;

namespace BotWorker.Domain.Entities
{
    [Table("msgcount")]
    public class GroupMsgCount
    {
        private static IGroupMsgCountRepository Repository => 
            BotMessage.ServiceProvider?.GetRequiredService<IGroupMsgCountRepository>() 
            ?? throw new InvalidOperationException("IGroupMsgCountRepository not registered");

        private static IGroupRepository GroupRepository => 
            BotMessage.ServiceProvider?.GetRequiredService<IGroupRepository>() 
            ?? throw new InvalidOperationException("IGroupRepository not registered");

        [Key]
        public long Id { get; set; }
        public long BotUin { get; set; }
        public long GroupId { get; set; }
        public string GroupName { get; set; } = "";
        public long UserId { get; set; }
        public string UserName { get; set; } = "";
        public DateTime CDate { get; set; }
        public DateTime MsgDate { get; set; }
        public int CMsg { get; set; }

        public static async Task<bool> ExistTodayAsync(long groupId, long userId)
        {
            return await Repository.ExistTodayAsync(groupId, userId);
        }

        public static async Task<int> AppendAsync(long botUin, long groupId, string groupName, long userId, string name)
        {
            return await Repository.AppendAsync(botUin, groupId, groupName, userId, name);
        }

        public static int Update(long botUin, long groupId, string groupName, long userId, string name)
            => UpdateAsync(botUin, groupId, groupName, userId, name).GetAwaiter().GetResult();

        public static async Task<int> UpdateAsync(long botUin, long groupId, string groupName, long userId, string name)
        {
            return await Repository.UpdateAsync(botUin, groupId, groupName, userId, name);
        }

        // 今日发言次数
        public static async Task<int> GetMsgCountAsync(long groupId, long qq)
        {
            return await Repository.GetMsgCountAsync(groupId, qq);
        }

        // 昨日发言次数
        public static async Task<int> GetMsgCountYAsync(long groupId, long qq)
        {
            return await Repository.GetMsgCountAsync(groupId, qq, true);
        }

        // 今日发言排名
        public static async Task<int> GetCountOrderAsync(long groupId, long userId)
        {
            return await Repository.GetCountOrderAsync(groupId, userId);
        }

        /// 昨日发言排名
        public static async Task<int> GetCountOrderYAsync(long groupId, long userId)
        {
            return await Repository.GetCountOrderAsync(groupId, userId, true);
        }

        // 今日发言榜前N名
        public static async Task<string> GetCountListAsync(long botUin, long groupId, long userId, long top)
        {
            if (!await GroupRepository.IsOwnerAsync(groupId, userId) && !BotInfo.IsAdmin(botUin, userId))
                return "此命令仅限管理员使用";

            string res = await Repository.GetCountListAsync(groupId, false, top);
            if (!res.Contains(userId.ToString()))
            {
                int order = await GetCountOrderAsync(groupId, userId);
                int count = await GetMsgCountAsync(groupId, userId);
                res += $"【第{order}名】 你 发言：{count}";
            }
            res += "\n进入 后台 查看更多内容";
            return res;
        }

        // 昨日发言榜前N名
        public static async Task<string> GetCountListYAsync(long botUin, long groupId, long userId, long top)
        {
            if (!await GroupRepository.IsOwnerAsync(groupId, userId) && !BotInfo.IsAdmin(botUin, userId))
                return "此命令仅限管理员使用";

            string res = await Repository.GetCountListAsync(groupId, true, top);
            if (!res.Contains(userId.ToString()))
            {
                int order = await GetCountOrderYAsync(groupId, userId);
                int count = await GetMsgCountYAsync(groupId, userId);
                res += $"【第{order}名】 你 发言：{count}";
            }
            res += "\n进入 后台 查看更多内容";
            return res;
        }
    }
}
