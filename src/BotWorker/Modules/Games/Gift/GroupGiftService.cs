using System;
using System.Data;
using System.Linq;
using System.Threading.Tasks;
using BotWorker.Domain.Entities;
using BotWorker.Domain.Interfaces;
using BotWorker.Domain.Repositories;
using Dapper;
using Microsoft.Extensions.Logging;

namespace BotWorker.Modules.Games.Gift
{
    public class GroupGiftService : IGroupGiftService
    {
        private readonly IGroupMemberRepository _groupMemberRepo;
        private readonly IGiftRepository _giftRepo;
        private readonly IGiftLogRepository _giftLogRepo;
        private readonly IUserRepository _userRepo;
        private readonly IGroupRepository _groupRepo;
        private readonly ILogger<GroupGiftService> _logger;

        public GroupGiftService(
            IGroupMemberRepository groupMemberRepo,
            IGiftRepository giftRepo,
            IGiftLogRepository giftLogRepo,
            IUserRepository userRepo,
            IGroupRepository groupRepo,
            ILogger<GroupGiftService> logger)
        {
            _groupMemberRepo = groupMemberRepo;
            _giftRepo = giftRepo;
            _giftLogRepo = giftLogRepo;
            _userRepo = userRepo;
            _groupRepo = groupRepo;
            _logger = logger;
        }

        public async Task<string> GetGiftAsync(long groupId, long userId)
        {
            return $"抽礼物：没有抽到任何礼物\n{userId} {groupId}";
        }

        public async Task<string> GetGiftResAsync(long botUin, long groupId, string groupName, long userId, string name, long qqGift, string giftName, int giftCount)
        {
            if (giftName == "礼物")
                return $"{GroupGift.GiftFormat}\n\n{await _giftRepo.GetGiftListAsync(botUin, groupId, userId)}";

            long giftId = giftName == "" ? await _giftRepo.GetRandomGiftAsync(botUin, groupId, userId) : await _giftRepo.GetGiftIdAsync(giftName);
            if (giftId == 0)
                return "不存在此礼物";

            var gift = await _giftRepo.GetByIdAsync(giftId);
            if (gift == null) return "不存在此礼物";
            
            long giftCredit = gift.GiftCredit;
            long creditMinus = giftCredit * giftCount;

            long creditAdd = creditMinus / 2;
            long creditAddOwner = creditAdd / 2;

            long credit_value = await _userRepo.GetCreditAsync(botUin, groupId, userId);
            if (credit_value < creditMinus)
                return $"您的积分{credit_value}不足{creditMinus}";

            long robotOwner = await _groupRepo.GetGroupOwnerAsync(groupId);
            string ownerName = await _userRepo.GetRobotOwnerNameAsync(groupId);
            string creditName = await _userRepo.GetCreditTypeAsync(botUin, groupId, userId);

            await _userRepo.AppendUserAsync(botUin, groupId, qqGift, "");

            using var trans = await _groupMemberRepo.BeginTransactionAsync();
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
                await _giftLogRepo.InsertAsync(log, trans);

                // 2. 扣分 (送礼者)
                var addRes1 = await _userRepo.AddCreditAsync(botUin, groupId, groupName, userId, name, -creditMinus, "礼物扣分", trans);
                if (addRes1.Result == -1) throw new Exception("礼物扣分失败");

                // 3. 对方加分
                var addRes2 = await _userRepo.AddCreditAsync(botUin, groupId, groupName, qqGift, "", creditAdd, "礼物加分", trans);
                if (addRes2.Result == -1) throw new Exception("对方加分失败");

                // 4. 主人加分
                var addRes3 = await _userRepo.AddCreditAsync(botUin, groupId, groupName, robotOwner, ownerName, creditAddOwner, "礼物加分", trans);
                if (addRes3.Result == -1) throw new Exception("主人加分失败");

                // 5. 亲密值
                await _groupMemberRepo.IncrementValueAsync("FansValue", creditMinus / 10 / 2, groupId, userId, trans);

                await trans.CommitAsync();

                // 同步缓存
                await _userRepo.SyncCacheFieldAsync(userId, groupId, "Credit", addRes1.CreditValue);
                await _userRepo.SyncCacheFieldAsync(qqGift, groupId, "Credit", addRes2.CreditValue);
                await _userRepo.SyncCacheFieldAsync(robotOwner, groupId, "Credit", addRes3.CreditValue);
                
                // Get updated FansValue
                long currentFansValue = await _groupMemberRepo.GetLongAsync("FansValue", groupId, userId);
                await _groupMemberRepo.SyncCacheFieldAsync(groupId, userId, "FansValue", currentFansValue);

                return $"✅ 送[@:{qqGift}]{giftName}*{giftCount}成功！\n亲密度值：+{creditMinus / 10 / 2}={currentFansValue}\n对方积分：+{creditAdd}={addRes2.CreditValue}\n" +
                       $"{creditName}：-{creditMinus}={addRes1.CreditValue}";
            }
            catch (Exception ex)
            {
                await trans.RollbackAsync();
                _logger.LogError(ex, "送礼物失败");
                return $"❌ 送礼物失败：{ex.Message}";
            }
        }

        public bool IsFans(long groupId, long userId)
        {
            return IsFansAsync(groupId, userId).GetAwaiter().GetResult();
        }

        public async Task<bool> IsFansAsync(long groupId, long userId)
        {
            return await _groupMemberRepo.GetValueAsync<int>("IsFans", groupId, userId) == 1;
        }

        public async Task<long> GetFansValueAsync(long groupId, long userId)
        {
            return await _groupMemberRepo.GetValueAsync<long>("FansValue", groupId, userId);
        }

        public async Task<long> GetFansRankingAsync(long groupId, long userId)
        {
            string sql = "SELECT COUNT(1) + 1 FROM group_member WHERE group_id = @groupId AND fans_value > (SELECT fans_value FROM group_member WHERE group_id = @groupId AND user_id = @userId)";
            return await _groupMemberRepo.ExecuteScalarAsync<long>(sql, new { groupId, userId });
        }

        public async Task<int> GetFansLevelAsync(long groupId, long userId)
        {
            return await _groupMemberRepo.GetValueAsync<int>("FansLevel", groupId, userId);
        }

        public int LampMinutes(long groupId, long userId)
        {
            // Implementation of LampMinutes for SQLite
            // In SQLite: (julianday('now') - julianday(LampDate)) * 24 * 60
            // But we can just fetch the date and calculate in C#
            var member = _groupMemberRepo.GetAsync(groupId, userId).Result;
            if (member == null || member.LampDate == default) return 99999;
            return (int)(DateTime.Now - member.LampDate).TotalMinutes;
        }

        public (string sql, object paras) SqlLightLamp(long groupId, long userId)
        {
            return ("UPDATE group_member SET lamp_date = @now, fans_value = fans_value + 10 WHERE group_id = @groupId AND user_id = @userId", 
                    new { now = DateTime.Now, groupId, userId });
        }

        public (string sql, object paras) SqlBingFans(long groupId, long userId)
        {
            // Check if exists first is handled by the caller or we can use UPSERT
            // For now, let's keep it simple as the original code
            return ("UPDATE group_member SET is_fans = 1, fans_date = @now, fans_level = 1, fans_value = 100 WHERE group_id = @groupId AND user_id = @userId",
                    new { now = DateTime.Now, groupId, userId });
        }

        public int GetFansCount(long groupId)
        {
            // Implementation of GetFansCount
            using var conn = _groupMemberRepo.CreateConnection();
            return conn.ExecuteScalar<int>("SELECT COUNT(1) FROM group_member WHERE group_id = @groupId AND is_fans = 1", new { groupId });
        }
    }
}
