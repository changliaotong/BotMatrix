using Microsoft.Data.SqlClient;
using sz84.Bots.Entries;
using sz84.Bots.Users;
using sz84.Core.MetaDatas;

namespace sz84.Bots.Games.Gift
{
    public class GroupGift : MetaData<GroupGift>
    {
        //粉丝团/粉丝灯牌/送礼物等功能
        public override string TableName => "GroupMember";
        public override string KeyField => "GroupId";
        public override string KeyField2 => "UserId";
               

        //抽礼物
        public static string GetGift(long groupId, long userId)
        {
            //todo 抽礼物
            return $"抽礼物：没有抽到任何礼物\n{userId} {groupId}";
        }

        public const string GiftFormat = "格式：赠送 + QQ + 礼物名 + 数量(默认1)\n例如：赠送 {客服QQ} 小心心 10";

        // 送礼物命令+参数
        public static string GetGiftRes(long botUin, long groupId, string groupName, long userId, string name, long qqGift, string giftName, int giftCount)
        {
            if (giftName == "")
                return $"{GiftFormat}\n\n{Gift.GetGiftList(groupId, userId)}";

            long giftId = giftName == "" ? Gift.GetRandomGift(groupId, userId) : Gift.GetGiftId(giftName);
            if (giftId == 0)
                return "不存在此礼物";

            long giftCredit = Gift.GetLong("GiftCredit", giftId);
            long creditMinus = giftCredit * giftCount;

            long creditAdd = creditMinus / 2;
            long creditAddOwner = creditAdd / 2;

            long credit_value = UserInfo.GetCredit(groupId, userId);
            if (credit_value < creditMinus)
                return $"您的积分{credit_value}不足{creditMinus}";

            long robotOwner = GroupInfo.GetGroupOwner(groupId);
            string ownerName = GroupInfo.GetRobotOwnerName(groupId);
            long credit_owner = UserInfo.GetCredit(groupId, robotOwner);

            UserInfo.AppendUser(botUin, groupId, qqGift, "");
            long creditOther = UserInfo.GetCredit(groupId, qqGift);
            creditOther += creditAdd;

            //更新亲密值 积分记录 更新记录
            if (qqGift == userId)
                creditOther -= creditMinus;

            if (robotOwner == userId)
                credit_owner -= creditMinus;

            //礼物记录
            var sql = GiftLog.SqlAppend(botUin, groupId, groupName, userId, name, robotOwner, ownerName, qqGift, "", giftId, giftName, giftCount, giftCredit);
            //扣分
            var sql2 = CreditLog.SqlHistory(botUin, groupId, groupName, userId, name, -creditMinus, "礼物扣分");
            var sql3 = UserInfo.SqlAddCredit(botUin, groupId, userId, -creditMinus);
            //对方加分
            var sql4 = CreditLog.SqlHistory(botUin, groupId, groupName, qqGift, "", creditAdd, "礼物加分");
            var sql5 = UserInfo.SqlAddCredit(botUin, groupId, qqGift, creditAdd);
            //主人加分
            var sql6 = CreditLog.SqlHistory(botUin, groupId, groupName, robotOwner, ownerName, creditAddOwner, "礼物加分");
            var sql7 = UserInfo.SqlAddCredit(botUin, groupId, robotOwner, creditAddOwner);
            //亲密值
            var sql8 = SqlPlus("FansValue", creditMinus / 10 / 2, groupId, userId);

            return ExecTrans(sql, sql2, sql3, sql4, sql5, sql6, sql7, sql8) == -1
                ? RetryMsg
                : $"✅ 送[@:{qqGift}]{giftName}*{giftCount}成功！\n亲密度值：+{creditMinus / 10 / 2}={{亲密度值}}\n对方积分：+{creditAdd}={UserInfo.GetCredit(groupId, qqGift)}\n" +
                  $"粉丝排名：第{{粉丝排名}}名 LV{{粉丝等级}}\n{{积分类型}}：-{creditMinus}={{积分}}";
        }

        // 粉丝排名
        public static string GetFansList(long groupId, long qq, int topN = 10)
        {
            string res = QueryRes($"select top {topN} UserId, FansValue, FansLevel from {FullName} " +
                                  $"where GroupId = {groupId} and IsFans = 1 order by FansValue desc",
                                      "【第{i}名】 [@:{0}] 亲密度：{1}\n");
            if (!res.Contains(qq.ToString()))
                res += $"【第{{粉丝排名}}名】 {qq} 亲密度：{GetInt("FansValue", groupId, qq)}";
            return $"{res}\n👪 粉丝团成员：{GetFansCount(groupId)}人";
        }

        // 加入粉丝团
        public static (string, SqlParameter[]) SqlBingFans(long groupId, long UserId)
        {
            return Exists(groupId, UserId)
                ? SqlUpdateWhere($"IsFans=1, FansDate=GETDATE(), FansLevel=1, FansValue=100", $"GroupID = {groupId} and UserId = {UserId}")
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
        public static (string, SqlParameter[]) SqlLightLamp(long groupId, long UserId)
        {
            return SqlUpdateWhere($"LampDate=GETDATE(), FansValue = FansValue + 10", $"GroupId = {groupId} and UserId = {UserId}");
        }

        // 是否点亮灯牌
        public static int LampMinutes(long groupId, long userId)
        {
            return GetInt("DATEDIFF(MINUTE, ISNULL(LampDate, GETDATE()-1), GETDATE())", groupId, userId);
        }

        //是否粉丝团成员
        public static bool IsFans(long groupId, long userId)
        {
            return GetBool("IsFans", groupId, userId);
        }

        // 亲密值 fans_value
        public static long GetFansValue(long groupId, long userId)
        {
            return GetInt("FansValue", groupId, userId);
        }

        // 粉丝等级
        public static int GetFansLevel(long groupId, long userId)
        {
            return GetInt($"{DbName}.dbo.get_fans_level(isnull(FansValue, 0))", groupId, userId);
        }

        // 粉丝团人数
        public static long GetFansCount(long groupId)
        {
            return CountWhere($"GroupId = {groupId} AND IsFans = 1");
        }

        // 粉丝团排名
        public static long GetFansOrder(long groupId, long userId)
        {
            return CountWhere($"GroupId = {groupId} AND IsFans = 1 AND FansValue > {GetFansValue(groupId, userId)}") + 1;
        }
    }
}
