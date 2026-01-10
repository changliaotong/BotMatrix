using System.Data;
using BotWorker.Domain.Entities;
using BotWorker.Common.Extensions;
using BotWorker.Infrastructure.Persistence.ORM;

namespace BotWorker.Modules.Games.Gift
{
    public class GroupGift : MetaData<GroupGift>
    {
        //粉丝团/粉丝灯牌/送礼物等功能
        public override string TableName => "GroupMember";
        public override string KeyField => "GroupId";
        public override string KeyField2 => "UserId";
               

        //抽礼物
        public static string GetGift(long groupId, long userId)
            => GetGiftAsync(groupId, userId).GetAwaiter().GetResult();

        public static async Task<string> GetGiftAsync(long groupId, long userId)
        {
            //todo 抽礼物
            return $"抽礼物：没有抽到任何礼物\n{userId} {groupId}";
        }

        public const string GiftFormat = "格式：赠送 + QQ + 礼物名 + 数量(默认1)\n例如：赠送 {客服QQ} 小心心 10";

        // 送礼物命令+参数
        public static string GetGiftRes(long botUin, long groupId, string groupName, long userId, string name, long qqGift, string giftName, int giftCount)
        {
            return GetGiftResAsync(botUin, groupId, groupName, userId, name, qqGift, giftName, giftCount).GetAwaiter().GetResult();
        }

        public static async Task<string> GetGiftResAsync(long botUin, long groupId, string groupName, long userId, string name, long qqGift, string giftName, int giftCount)
        {
            if (giftName == "")
                return $"{GiftFormat}\n\n{await Gift.GetGiftListAsync(groupId, userId)}";

            long giftId = giftName == "" ? await Gift.GetRandomGiftAsync(groupId, userId) : await Gift.GetGiftIdAsync(giftName);
            if (giftId == 0)
                return "不存在此礼物";

            long giftCredit = await Gift.GetLongAsync("GiftCredit", giftId);
            long creditMinus = giftCredit * giftCount;

            long creditAdd = creditMinus / 2;
            long creditAddOwner = creditAdd / 2;

            long credit_value = await UserInfo.GetCreditAsync(groupId, userId);
            if (credit_value < creditMinus)
                return $"您的积分{credit_value}不足{creditMinus}";

            long robotOwner = await GroupInfo.GetGroupOwnerAsync(groupId);
            string ownerName = await GroupInfo.GetRobotOwnerNameAsync(groupId);
            string creditName = await UserInfo.GetCreditTypeAsync(groupId, userId);

            await UserInfo.AppendUserAsync(botUin, groupId, qqGift, "");

            using var trans = await BeginTransactionAsync();
            try
            {
                // 1. 礼物记录
                var (sqlGift, parasGift) = GiftLog.SqlAppend(botUin, groupId, groupName, userId, name, robotOwner, ownerName, qqGift, "", giftId, giftName, giftCount, giftCredit);
                await ExecAsync(sqlGift, trans, parasGift);

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
                var (sqlFans, parasFans) = SqlPlus("FansValue", creditMinus / 10 / 2, groupId, userId);
                await ExecAsync(sqlFans, trans, parasFans);

                await trans.CommitAsync();

                // 同步缓存
                UserInfo.SyncCacheField(userId, groupId, "Credit", addRes1.CreditValue);
                UserInfo.SyncCacheField(qqGift, groupId, "Credit", addRes2.CreditValue);
                UserInfo.SyncCacheField(robotOwner, groupId, "Credit", addRes3.CreditValue);
                
                long currentFansValue = await GetFansValueAsync(groupId, userId);
                SyncCacheField(userId, groupId, "FansValue", currentFansValue);

                long fansOrder = await GetFansOrderAsync(groupId, userId);
                int fansLevel = await GetFansLevelAsync(groupId, userId);

                return $"✅ 送[@:{qqGift}]{giftName}*{giftCount}成功！\n亲密度值：+{creditMinus / 10 / 2}={currentFansValue}\n对方积分：+{creditAdd}={addRes2.CreditValue}\n" +
                       $"粉丝排名：第{fansOrder}名 LV{fansLevel}\n{creditName}：-{creditMinus}={addRes1.CreditValue}";
            }
            catch (Exception ex)
            {
                await trans.RollbackAsync();
                Console.WriteLine($"[GetGiftRes Error] {ex.Message}");
                return RetryMsg;
            }
        }

        // 粉丝排名
        public static string GetFansList(long groupId, long qq, int topN = 10)
            => GetFansListAsync(groupId, qq, topN).GetAwaiter().GetResult();

        public static async Task<string> GetFansListAsync(long groupId, long qq, int topN = 10)
        {
            string sql = $"select {SqlTop(topN)} UserId, FansValue, FansLevel from {FullName} " +
                                  $"where GroupId = {groupId} and IsFans = 1 order by FansValue desc {SqlLimit(topN)}";
            string res = await QueryResAsync(sql, "【第{i}名】 [@:{0}] 亲密度：{1}\n");
            if (!res.Contains(qq.ToString()))
                res += $"【第{{粉丝排名}}名】 {qq} 亲密度：{await GetIntAsync("FansValue", groupId, qq)}";
            return $"{res}\n👪 粉丝团成员：{await GetFansCountAsync(groupId)}人";
        }

        // 加入粉丝团
        public static (string, IDataParameter[]) SqlBingFans(long groupId, long UserId)
            => SqlBingFansAsync(groupId, UserId).GetAwaiter().GetResult();

        public static async Task<(string, IDataParameter[])> SqlBingFansAsync(long groupId, long UserId)
        {
            return await ExistsAsync(groupId, UserId)
                ? SqlUpdateWhere($"IsFans=1, FansDate={SqlDateTime}, FansLevel=1, FansValue=100", $"GroupID = {groupId} and UserId = {UserId}")
                : SqlInsert([
                                new Cov("GroupId", groupId),
                                new Cov("UserId", UserId),
                                new Cov("IsFans", 1),
                                new Cov("FansDate", DateTime.MinValue),
                                new Cov("FansLevel", 1),
                                new Cov("FansValue", 100)
                            ]);
        }

        // 点亮灯牌sql
        public static (string, IDataParameter[]) SqlLightLamp(long groupId, long UserId)
        {
            return SqlUpdateWhere($"LampDate={SqlDateTime}, FansValue = FansValue + 10", $"GroupId = {groupId} and UserId = {UserId}");
        }

        // 是否点亮灯牌
        public static int LampMinutes(long groupId, long userId)
            => LampMinutesAsync(groupId, userId).GetAwaiter().GetResult();

        public static async Task<int> LampMinutesAsync(long groupId, long userId)
        {
            string sql = SqlDateDiff("MINUTE", SqlIsNull("LampDate", SqlDateAdd("day", -1, SqlDateTime)), SqlDateTime);
            return await GetIntAsync(sql, groupId, userId);
        }

        //是否粉丝团成员
        public static bool IsFans(long groupId, long userId)
            => IsFansAsync(groupId, userId).GetAwaiter().GetResult();

        public static async Task<bool> IsFansAsync(long groupId, long userId)
        {
            return await GetBoolAsync("IsFans", groupId, userId);
        }

        // 亲密值 fans_value
        public static long GetFansValue(long groupId, long userId)
            => GetFansValueAsync(groupId, userId).GetAwaiter().GetResult();

        public static async Task<long> GetFansValueAsync(long groupId, long userId)
        {
            return await GetIntAsync("FansValue", groupId, userId);
        }

        // 粉丝等级
        public static int GetFansLevel(long groupId, long userId)
            => GetFansLevelAsync(groupId, userId).GetAwaiter().GetResult();

        public static async Task<int> GetFansLevelAsync(long groupId, long userId)
        {
            string func = IsPostgreSql ? "get_fans_level" : $"{DbName}.dbo.get_fans_level";
            return await GetIntAsync($"{func}({SqlIsNull("FansValue", "0")})", groupId, userId);
        }

        // 粉丝团人数
        public static long GetFansCount(long groupId)
            => GetFansCountAsync(groupId).GetAwaiter().GetResult();

        public static async Task<long> GetFansCountAsync(long groupId)
        {
            return await CountWhereAsync($"GroupId = {groupId} AND IsFans = 1");
        }

        // 粉丝团排名
        public static long GetFansOrder(long groupId, long userId)
            => GetFansOrderAsync(groupId, userId).GetAwaiter().GetResult();

        public static async Task<long> GetFansOrderAsync(long groupId, long userId)
        {
            return await CountWhereAsync($"GroupId = {groupId} AND IsFans = 1 AND FansValue > {await GetFansValueAsync(groupId, userId)}") + 1;
        }
    }
}
