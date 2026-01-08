using Microsoft.Data.SqlClient;
using System.Reflection;
using BotWorker.Bots.BotMessages;
using BotWorker.Bots.Entries;
using BotWorker.Core.MetaDatas;
using BotWorker.Groups;

namespace BotWorker.Bots.Users
{
    public partial class UserInfo : MetaDataGuid<UserInfo>
    {
        //读取积分（直读数据库，用于逻辑判断）
        public static long GetCreditNoCache(long botUin, long groupId, long qq)
        {
            if (groupId != 0 && GroupInfo.GetIsCredit(groupId))
                return GroupMember.GetFieldNoCache<long>("GroupCredit", groupId, qq);
            
            return GetFieldNoCache<long>("Credit", qq);
        }

        //增加积分
        public static (int, long) AddCredit(long botUin, long groupId, string groupName, long qq, string name, long creditAdd, string creditInfo)
        {
            // 1. 使用直读数据库获取当前余额，确保准确
            var creditValue = GetCreditNoCache(botUin, groupId, qq);
            
            if (Append(botUin, groupId, qq, name, GroupInfo.GetGroupOwner(groupId)) == -1)
                return (-1, creditValue);

            var sql = SqlAddCredit(botUin, groupId, qq, creditAdd);
            var sql2 = CreditLog.SqlHistory(botUin, groupId, groupName, qq, name, creditAdd, creditInfo);
            
            int result = ExecTrans(sql, sql2);
            
            // 2. 写入成功后，立即同步缓存
            if (result != -1)
            {
                var newValue = creditValue + creditAdd;
                if (groupId != 0 && GroupInfo.GetIsCredit(groupId))
                    GroupMember.SyncCacheField(groupId, qq, "GroupCredit", newValue);
                else
                    SyncCacheField(qq, "Credit", newValue);
            }

            return (result, creditValue + creditAdd);
        }

        public static (int, long) MinusCredit(long botUin, long groupId, string groupName, long qq, string name, long creditMinus, string creditInfo)
            => AddCredit(botUin, groupId, groupName, qq, name, -creditMinus, creditInfo);


        //增加积分sql
        public static SqlTask TaskAddCredit(long botUin, long groupId, long userId, long creditPlus)
        {
            if (GroupInfo.GetIsCredit(groupId))
            {
                return GroupMember.TaskAddCredit(groupId, userId, creditPlus);
            }
            
            if (Exists(userId))
                return TaskPlus("Credit", creditPlus, userId);
            
            var (sql, parameters) = SqlInsert([
                new Cov("BotUin", botUin),
                new Cov("GroupId", groupId),
                new Cov("Id", userId),
                new Cov("Credit", creditPlus),
            ]);
            // 插入操作也可以利用 ReSync 确保缓存被填充
            return new SqlTask(sql, parameters, userId, "Credit", true);
        }

        //转账积分
        public static int TransferCredit(long botUin, long groupId, string groupName, long qq, string name, long qqTo, string nameTo, long creditMinus, long creditAdd, ref long creditValue, ref long creditValue2, string transferInfo)
        {
            int i = AppendUser(botUin, groupId, qqTo, nameTo);
            if (i == -1)
                return i;

            creditValue = GetCreditNoCache(botUin, groupId, qq);
            if (creditValue < creditMinus)
                return -1;

            creditValue2 = GetCreditNoCache(botUin, groupId, qqTo) + creditAdd;

            return ExecTrans(
                TaskAddCredit(botUin, groupId, qq, -creditMinus),
                TaskAddCredit(botUin, groupId, qqTo, creditAdd),
                CreditLog.SqlHistory(botUin, groupId, groupName, qq, name, -creditMinus, $"{transferInfo}扣分：{qqTo}"),
                CreditLog.SqlHistory(botUin, groupId, groupName, qqTo, nameTo, creditAdd, $"{transferInfo}加分：{qq}")
            );
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

        public static SqlTask TaskSaveCredit(long botUin, long groupId, long userId, long creditSave)
        {
            if (GroupInfo.GetIsCredit(groupId))
                return GroupMember.TaskSaveCredit(groupId, userId, creditSave);
            
            if (BotInfo.GetIsCredit(botUin))
                return Friend.TaskSaveCredit(botUin, userId, creditSave);

            var (sql, parameters) = SqlSetValues($"Credit = Credit - @creditSave, SaveCredit = isnull(SaveCredit, 0) + @creditSave", userId);
            var paramList = parameters.ToList();
            paramList.Add(new SqlParameter("@creditSave", creditSave));

            // 因为一次更新了两个字段，且是原子计算，所以标记 NeedsReSync = true
            // 在 ExecTrans 中，我们会自动重新从数据库读取这两个字段并同步到缓存
            return new SqlTask(sql, [.. paramList], userId, null, "Credit", true);
        }

        public static SqlTask TaskFreezeCredit(long userId, long creditFreeze)
        {
            var (sql, parameters) = SqlSetValues($"Credit = Credit - @creditFreeze, FreezeCredit = isnull(FreezeCredit, 0) + @creditFreeze", userId);
            var paramList = parameters.ToList();
            paramList.Add(new SqlParameter("@creditFreeze", creditFreeze));

            return new SqlTask(sql, [.. paramList], userId, null, "Credit", true);
        }

        public static int DoFreezeCredit(long botUin, long groupId, string groupName, long qq, string name, long creditFreeze)
        {
            long creditValue = GetCreditNoCache(botUin, groupId, qq);
            if (creditValue < creditFreeze)
                return -1;

            return ExecTrans(
                TaskFreezeCredit(qq, creditFreeze),
                CreditLog.SqlHistory(botUin, groupId, groupName, qq, name, -creditFreeze, "冻结积分")
            );
        }

        public static long GetFreezeCredit(long qq) => GetLong("FreezeCredit", qq);

        public static int UnfreezeCredit(long botUin, long groupId, string groupName, long qq, string name, long creditUnfreeze)
        {
            long creditValue = GetFreezeCredit(qq);
            if (creditValue < creditUnfreeze)
                return -1;

            return ExecTrans(
                TaskFreezeCredit(qq, -creditUnfreeze),
                CreditLog.SqlHistory(botUin, groupId, groupName, qq, name, creditUnfreeze, "解冻积分")
            );
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
