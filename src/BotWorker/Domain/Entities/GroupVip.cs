using Microsoft.Data.SqlClient;
using BotWorker.Bots.Entries;
using BotWorker.Bots.Models.Office;
using BotWorker.BotWorker.Common.Exts;
using BotWorker.Infrastructure.Persistence.ORM;
using BotWorker.Bots.Users;

namespace BotWorker.Domain.Entities
{

    public class GroupVip : MetaData<GroupVip>
    {
        public override string TableName => "VIP";
        public override string KeyField => "GroupId";

        // 购买机器人
        public static int BuyRobot(long botUin, long groupId, string groupName, long qqBuyer, string buyerName, long month, decimal payMoney, string payMethod, string trade, string memo, int insertBy)
        {
            GroupInfo.Append(groupId, groupName, BotInfo.BotUinDef, BotInfo.BotNameDef);
            UserInfo.AppendUser(botUin, groupId, qqBuyer, buyerName);
            var sqlIncome = Income.SqlInsert(groupId, month, "机器人", payMoney, payMethod, trade, memo, qqBuyer, insertBy);
            var sqlBuyVip = SqlBuyVip(groupId, groupName, qqBuyer, month, payMoney, memo);
            return ExecTrans(sqlIncome, sqlBuyVip);
        }

        // 购买、续费机器人
        public static (string, SqlParameter[]) SqlBuyVip(long groupId, string groupName, long userId, long month, decimal payMoney, string vipInfo, int insertBy = BotInfo.SystemUid)
        {
            int is_year_vip = IsYearVIP(groupId) || RestMonths(groupId) + month >= 12 ? 1 : 0;       
            return IsVip(groupId)
                ? ($"update {FullName} set EndDate = dateadd(month, {month}, EndDate), UserId = {userId}, " +
                  $"IncomeDay = (IncomeDay * DATEDIFF(MONTH,GETDATE(),EndDate) + {payMoney}) /((DATEDIFF(MONTH,GETDATE(),EndDate) + {month} + 0.0000001)*1.0)," +
                  $"IsYearVip = {is_year_vip}, InsertBy = {insertBy}, IsGoon = null where GroupId = {groupId}" , [])

                : SqlInsert([
                                new Cov("GroupId", groupId),
                                new Cov("GroupName", groupName),
                                new Cov("FirstPay", payMoney),
                                new Cov("StartDate", DateTime.MinValue),
                                new Cov("EndDate", GetDate().ToDateTime().AddMonths((int)month)),
                                new Cov("VipInfo", vipInfo),
                                new Cov("UserId", userId),
                                new Cov("IncomeDay", payMoney / month),
                                new Cov("IsYearVip", is_year_vip),
                                new Cov("InsertBy", insertBy),
                            ]);
        }


        // 换群
        public static int ChangeGroup(long groupId, long newGroupId, long qq)
        {
            return Exec($"exec sz84_robot..sp_ChangeVIP {groupId}, {newGroupId}, {qq}, {BotInfo.SystemUid}");
        }

        public static int RestDays(long groupId)
        {
            return GetInt("datediff(day,getdate(),EndDate)", groupId);
        }
 
        // 是否年费版
        public static bool IsYearVIP(long groupId)
        {
            return GetBool("IsYearVip", groupId);
        }

        public static bool IsVip(long groupId)
        {
            return Exists(groupId);
        }

        public static bool IsForever(long groupId)
        {
            return RestDays(groupId) > 3650;
        }

        //是否开通过VIP
        public static bool IsVipOnce(long groupId)
        {
            return Income.IsVipOnce(groupId);
        }

        public static bool IsClientVip(long qq)
        {
            return ExistsField("UserId", qq.ToString());
        }

        public static int RestMonths(long groupId)
        {
            return GetInt("DATEDIFF(MONTH,GETDATE(), EndDate)", groupId);
        }
    }
}
