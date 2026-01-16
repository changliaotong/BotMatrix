using System;
using System.Data;
using System.Threading.Tasks;
using BotWorker.Domain.Entities;
using BotWorker.Domain.Models.BotMessages;
using BotWorker.Domain.Repositories;
using Dapper.Contrib.Extensions;
using Microsoft.Extensions.DependencyInjection;

namespace BotWorker.Modules.Games.Gift
{
    [Table("GroupMember")]
    public class GroupGift
    {
        private static IGroupGiftRepository Repository => 
            BotMessage.ServiceProvider?.GetRequiredService<IGroupGiftRepository>() 
            ?? throw new InvalidOperationException("IGroupGiftRepository not registered");

        [ExplicitKey]
        public long GroupId { get; set; }
        [ExplicitKey]
        public long UserId { get; set; }
        public long FansValue { get; set; }

        public const string GiftFormat = "格式：赠送 + QQ + 礼物名 + 数量(默认1)\n例如：赠送 {客服QQ} 小心心 10";

        public static async Task<string> GetGiftAsync(long groupId, long userId)
        {
            return $"抽礼物：没有抽到任何礼物\n{userId} {groupId}";
        }

        public static async Task<string> GetGiftResAsync(long botUin, long groupId, string groupName, long userId, string name, long qqGift, string giftName, int giftCount)
        {
            if (giftName == "礼物")
                return $"{GiftFormat}\n\n{await Gift.GetGiftListAsync(botUin, groupId, userId)}";

            long giftId = giftName == "" ? await Gift.GetRandomGiftAsync(botUin, groupId, userId) : await Gift.GetGiftIdAsync(giftName);
            if (giftId == 0)
                return "不存在此礼物";

            var gift = await Gift.GetAsync(giftId);
            if (gift == null) return "不存在此礼物";
            
            long giftCredit = gift.GiftCredit;
            long creditMinus = giftCredit * giftCount;

            long creditAdd = creditMinus / 2;
            long creditAddOwner = creditAdd / 2;

            long credit_value = await UserInfo.GetCreditAsync(botUin, groupId, userId);
            if (credit_value < creditMinus)
                return $"您的积分{credit_value}不足{creditMinus}";

            long robotOwner = await GroupInfo.GetGroupOwnerAsync(groupId);
            string ownerName = await GroupInfo.GetRobotOwnerNameAsync(groupId);
            string creditName = await UserInfo.GetCreditTypeAsync(botUin, groupId, userId);

            await UserInfo.AppendUserAsync(botUin, groupId, qqGift, "");

            using var trans = await Repository.BeginTransactionAsync();
            try
            {
                // 1. 礼物记录
                var log = new GiftLog
                {
                    BotUin = botUin,
                    GroupId = groupId,
                    GroupName = groupName,
                    UserId = userId,
                    UserName = name,
                    RobotOwner = robotOwner,
                    OwnerName = ownerName,
                    GiftUserId = qqGift,
                    GiftUserName = "",
                    GiftId = giftId,
                    GiftName = giftName,
                    GiftCount = giftCount,
                    GiftCredit = giftCredit
                };
                await log.InsertAsync(trans);

                // 2. 扣分 (送礼者)
                var addRes1 = await UserInfo.AddCreditAsync(botUin, groupId, groupName, userId, name, -creditMinus, "礼物扣分", trans);
                if (addRes1.Result == -1) throw new Exception("礼物扣分失败");

                // 3. 对方加分
                var addRes2 = await UserInfo.AddCreditAsync(botUin, groupId, groupName, qqGift, "", creditAdd, "礼物加分", trans);
                if (addRes2.Result == -1) throw new Exception("对方加分失败");

                // 4. 主人加分
                var addRes3 = await UserInfo.AddCreditAsync(botUin, groupId, groupName, robotOwner, ownerName, creditAddOwner, "礼物加分", trans);
                if (addRes3.Result == -1) throw new Exception("主人加分失败");

                // 5. 亲密值
                const string sqlFans = "UPDATE GroupMember SET FansValue = FansValue + @value WHERE GroupId = @groupId AND UserId = @userId";
                await trans.Connection.ExecuteAsync(sqlFans, new { value = creditMinus / 10 / 2, groupId, userId }, trans);

                await trans.CommitAsync();

                // 同步缓存
                UserInfo.SyncCacheField(userId, groupId, "Credit", addRes1.CreditValue);
                UserInfo.SyncCacheField(qqGift, groupId, "Credit", addRes2.CreditValue);
                UserInfo.SyncCacheField(robotOwner, groupId, "Credit", addRes3.CreditValue);
                
                // Get updated FansValue
                const string sqlGetFans = "SELECT FansValue FROM GroupMember WHERE GroupId = @groupId AND UserId = @userId";
                long currentFansValue = await Repository.CreateConnection().ExecuteScalarAsync<long>(sqlGetFans, new { groupId, userId });
                
                // GroupMember.SyncCacheField(userId, groupId, "FansValue", currentFansValue); // Need to check if GroupMember has SyncCacheField

                return $"✅ 送[@:{qqGift}]{giftName}*{giftCount}成功！\n亲密度值：+{creditMinus / 10 / 2}={currentFansValue}\n对方积分：+{creditAdd}={addRes2.CreditValue}\n" +
                       $"{creditName}：-{creditMinus}={addRes1.CreditValue}";
            }
            catch (Exception ex)
            {
                await trans.RollbackAsync();
                return $"❌ 送礼物失败：{ex.Message}";
            }
        }
    }
}
