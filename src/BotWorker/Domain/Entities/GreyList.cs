using System;
using System.Collections.Generic;
using System.Threading.Tasks;
using BotWorker.Domain.Repositories;
using Microsoft.Extensions.DependencyInjection;

namespace BotWorker.Domain.Entities
{
    public class GreyList
    {
        private static IGreyListRepository Repository => 
            BotMessage.ServiceProvider?.GetRequiredService<IGreyListRepository>() 
            ?? throw new InvalidOperationException("IGreyListRepository not registered");

        public long BotUin { get; set; }
        public long GroupId { get; set; }
        public string GroupName { get; set; } = string.Empty;
        public long UserId { get; set; }
        public string UserName { get; set; } = string.Empty;
        public long GreyId { get; set; }
        public string GreyInfo { get; set; } = string.Empty;
        public DateTime InsertDate { get; set; }

        // 灰名单指令：灰、加灰、删灰、取消灰、解除灰名单…
        public const string regexGrey = @"^(?<cmdName>(取消|解除|删除)?(灰名单|灰|加灰|删灰))(?<cmdPara>([ ]*(\[?@:?)?[1-9]+\d*(\]?))+)$";

        // 系统灰名单（通用灰名单）
        public static async Task<IEnumerable<long>> GetSystemGreyListAsync()
        {
            return await Repository.GetSystemGreyListAsync();
        }

        public static bool IsSystemGrey(long userId)
        {
            return Repository.IsExistsAsync(BotInfo.GroupIdDef, userId).GetAwaiter().GetResult();
        }

        // 加入灰名单
        public static int AddGreyList(long botUin, long groupId, string groupName, long qq, string name, long greyQQ, string greyInfo)
        {
            return AddGreyListAsync(botUin, groupId, groupName, qq, name, greyQQ, greyInfo).GetAwaiter().GetResult();
        }

        public static async Task<int> AddGreyListAsync(long botUin, long groupId, string groupName, long qq, string name, long greyQQ, string greyInfo)
        {
            if (await Repository.IsExistsAsync(groupId, greyQQ))
                return 0;

            var greyList = new GreyList
            {
                BotUin = botUin,
                GroupId = groupId,
                GroupName = groupName,
                UserId = qq,
                UserName = name,
                GreyId = greyQQ,
                GreyInfo = greyInfo
            };

            return await Repository.AddAsync(greyList);
        }

        public static async Task<bool> ExistsAsync(long groupId, long greyId)
        {
            return await Repository.IsExistsAsync(groupId, greyId);
        }

        public static async Task<int> DeleteAsync(long groupId, long greyId)
        {
            return await Repository.DeleteAsync(groupId, greyId);
        }
    }
}
