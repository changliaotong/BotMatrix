using Microsoft.Data.SqlClient;
using System.Reflection;
using sz84.Bots.BotMessages;
using sz84.Bots.Entries;
using BotWorker.Common.Exts;
using BotWorker.Infrastructure.Persistence.ORM;
using sz84.Groups;

namespace sz84.Bots.Users
{
    public partial class UserInfo : MetaDataGuid<UserInfo>
    {
        //增加积分
        public static (int, long) AddCredit(long botUin, long groupId, string groupName, long qq, string name, long creditAdd, string creditInfo)
        {
            var creditValue = GetCredit(groupId, qq);
            if (Append(botUin, groupId, qq, name, GroupInfo.GetGroupOwner(groupId)) == -1)
                return (-1, creditValue);

            var sql = SqlAddCredit(botUin, groupId, qq, creditAdd);
            var sql2 = CreditLog.SqlHistory(botUin, groupId, groupName, qq, name, creditAdd, creditInfo);
            int result = ExecTrans(sql, sql2);
            if (result == 0)
            {
                SyncCacheField(qq, groupId, "Credit", creditValue + creditAdd);
            }
            return (result, creditValue + creditAdd);
        }

        public static (int, long) MinusCredit(long botUin, long groupId, string groupName, long qq, string name, long creditMinus, string creditInfo)
            => AddCredit(botUin, groupId, groupName, qq, name, -creditMinus, creditInfo);


        //增加积分sql
        public static (string, SqlParameter[]) SqlAddCredit(long botUin, long groupId, long userId, long creditPlus)
        {
            if (GroupInfo.GetIsCredit(groupId))
            {
                return GroupMember.SqlAddCredit(groupId, userId, creditPlus);
            }
            else if (BotInfo.GetIsCredit(botUin))
            {
                return Friend.SqlAddCredit(botUin, userId, creditPlus);
            }
            else
            {
                if (Exists(userId))
                    return SqlPlus("Credit", creditPlus, userId);
                else
                    return SqlInsert([
                        new Cov("BotUin", botUin),
                        new Cov("GroupId", groupId),
                        new Cov("Id", userId),
                        new Cov("Credit", creditPlus),
                    ]);
            }
        }

        //转账积分
        public static int TransferCredit(long botUin, long groupId, string groupName, long qq, string name, long qqTo, string nameTo, long creditMinus, long creditAdd, ref long creditValue, ref long creditValue2, string transferInfo)
        {
            int i = AppendUser(botUin, groupId, qqTo, nameTo);
            if (i == -1)
                return i;

            creditValue = GetCredit(groupId, qq);
            if (creditValue < creditMinus)
                return -1;

            creditValue -= creditMinus;
            creditValue2 = GetCredit(groupId, qqTo) + creditAdd;

            var sql = SqlAddCredit(botUin, groupId, qq, -creditMinus);
            var sql2 = SqlAddCredit(botUin, groupId, qqTo, creditAdd);
            var sql3 = CreditLog.SqlHistory(botUin, groupId, groupName, qq, name, -creditMinus, $"{transferInfo}扣分：{qqTo}");
            var sql4 = CreditLog.SqlHistory(botUin, groupId, groupName, qqTo, nameTo, creditAdd, $"{transferInfo}加分：{qq}");

            int result = ExecTrans(sql, sql3, sql2, sql4);
            if (result == 0)
            {
                SyncCacheField(qq, groupId, "Credit", creditValue);
                SyncCacheField(qqTo, groupId, "Credit", creditValue2);
            }
            return result;
        }


        //读取积分
        public static long GetCredit(long botUin, long groupId, long qq)
        {
            return groupId != 0 && GroupInfo.GetIsCredit(groupId)
                ? GroupMember.GetGroupCredit(groupId, qq)
                : GetCredit(botUin, qq);
        }

        public static long GetCredit(long userId)
        {
            return GetLong("Credit", userId);
        }

        //读取积分
        public static long GetCredit(long botUin, long userId)
        {
            return BotInfo.GetIsCredit(botUin) ? Friend.GetCredit(botUin, userId) : GetLong("credit", userId);
        }

        //积分总额
        public static long GetTotalCredit(long userId) => GetCredit(userId) + GetSaveCredit(userId);

        public static long GetTotalCredit(long groupId, long qq) => GetCredit(groupId, qq) + GetSaveCredit(groupId, qq);

        public static long GetSaveCredit(long botUin, long userId)
        {
            return BotInfo.GetIsCredit(botUin)
                ? Friend.GetSaveCredit(botUin, userId)
                : GetSaveCredit(userId);
        }

        public static long GetSaveCredit(long botUin, long groupId, long qq)
        {
            return GroupInfo.GetIsCredit(groupId)
                ? GroupMember.GetLong("SaveCredit", groupId, qq)
                : GetSaveCredit(qq);
        }

        public static long GetSaveCredit(long userId)
        {
            return GetLong("SaveCredit", userId);
        }

        public static (string, SqlParameter[]) SqlSaveCredit(long botUin, long groupId, long userId, long creditSave)
        {
            return GroupInfo.GetIsCredit(groupId)
                ? GroupMember.SqlSaveCredit(groupId, userId, creditSave)
                : BotInfo.GetIsCredit(botUin) ? Friend.SqlSaveCredit(botUin, userId, creditSave)
                                 : SqlSetValues($"Credit = Credit - ({creditSave}), SaveCredit = isnull(SaveCredit, 0) + ({creditSave})", userId);
        }

        public static (string, SqlParameter[]) SqlFreezeCredit(long userId, long creditFreeze)
        {
            return SqlSetValues($"Credit = Credit - ({creditFreeze}), FreezeCredit = isnull(FreezeCredit, 0) + ({creditFreeze})", userId);
        }

        public static int DoFreezeCredit(long botUin, long groupId, string groupName, long qq, string name, long creditFreeze)
        {
            long creditValue = GetCredit(groupId, qq);
            if (creditValue < creditFreeze)
                return -1;

            var sql = SqlFreezeCredit(qq, creditFreeze);
            var sql2 = CreditLog.SqlHistory(botUin, groupId, groupName, qq, name, -creditFreeze, "冻结积分");
            int result = ExecTrans(sql, sql2);
            if (result == 0)
            {
                SyncCacheField(qq, groupId, "Credit", creditValue - creditFreeze);
            }
            return result;
        }

        public static long GetFreezeCredit(long qq) => GetLong("FreezeCredit", qq);

        public static int UnfreezeCredit(long botUin, long groupId, string groupName, long qq, string name, long creditUnfreeze)
        {
            long creditValue = GetFreezeCredit(qq);
            if (creditValue < creditUnfreeze)
                return -1;

            var sql = SqlFreezeCredit(qq, -creditUnfreeze);
            var sql2 = CreditLog.SqlHistory(botUin, groupId, groupName, qq, name, creditUnfreeze, "解冻积分");
            int result = ExecTrans(sql, sql2);
            if (result == 0)
            {
                SyncCacheField(qq, groupId, "Credit", GetCredit(groupId, qq) + creditUnfreeze);
            }
            return result;
        }

        public static long GetCreditRanking(long botUin, long groupId, long qq)
        {
            long credit_value = GetCredit(groupId, qq);
            return GroupInfo.GetIsCredit(groupId)
                ? GroupMember.CountWhere($"GroupId = {groupId} and Credit > {credit_value}") + 1
                : BotInfo.GetBool("IsCredit", botUin)
                    ? Friend.CountWhere($"BotUin = {botUin} and Credit > {credit_value} and Id in (select UserId from {GroupMember.FullName} where GroupId = {groupId})") + 1
                    : CountWhere($"Credit > {credit_value} and Id in (select UserId from {GroupMember.FullName} where GroupId = {groupId})") + 1;
        }

        public static long GetCreditRankingAll(long qq)
        {
            return CountWhere($"Credit + SaveCredit > {GetTotalCredit(qq)}") + 1;
        }
    }
}
