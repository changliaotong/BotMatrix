using System.Data;
using BotWorker.Infrastructure.Persistence.ORM;

namespace BotWorker.Domain.Entities
{

    public class GroupVip : MetaData<GroupVip>
    {
        public override string TableName => "VIP";
        public override string KeyField => "GroupId";

        // 购买机器人
        public static async Task<int> BuyRobotAsync(long botUin, long groupId, string groupName, long qqBuyer, string buyerName, long month, decimal payMoney, string payMethod, string trade, string memo, int insertBy)
        {
            await GroupInfo.AppendAsync(groupId, groupName, BotInfo.BotUinDef, BotInfo.BotNameDef);
            await UserInfo.AppendUserAsync(botUin, groupId, qqBuyer, buyerName);
            var sqlIncome = Income.SqlInsert(groupId, month, "机器人", payMoney, payMethod, trade, memo, qqBuyer, insertBy);
            var sqlBuyVip = await SqlBuyVipAsync(groupId, groupName, qqBuyer, month, payMoney, memo);

            using var trans = await BeginTransactionAsync();
            try
            {
                var (sql1, paras1) = sqlIncome;
                await ExecAsync(sql1, trans, paras1);
                
                var (sql2, paras2) = sqlBuyVip;
                await ExecAsync(sql2, trans, paras2);
                
                await trans.CommitAsync();
                return 0;
            }
            catch (Exception ex)
            {
                Console.WriteLine($"BuyRobotAsync error: {ex.Message}");
                await trans.RollbackAsync();
                return -1;
            }
        }

        // 购买机器人
        public static int BuyRobot(long botUin, long groupId, string groupName, long qqBuyer, string buyerName, long month, decimal payMoney, string payMethod, string trade, string memo, int insertBy)
        {
            return BuyRobotAsync(botUin, groupId, groupName, qqBuyer, buyerName, month, payMoney, payMethod, trade, memo, insertBy).GetAwaiter().GetResult();
        }

        // 购买、续费机器人
        public static (string, IDataParameter[]) SqlBuyVip(long groupId, string groupName, long userId, long month, decimal payMoney, string vipInfo, int insertBy = BotInfo.SystemUid)
            => SqlBuyVipAsync(groupId, groupName, userId, month, payMoney, vipInfo, insertBy).GetAwaiter().GetResult();

        public static async Task<(string, IDataParameter[])> SqlBuyVipAsync(long groupId, string groupName, long userId, long month, decimal payMoney, string vipInfo, int insertBy = BotInfo.SystemUid)
        {
            int is_year_vip = await IsYearVIPAsync(groupId) || await RestMonthsAsync(groupId) + month >= 12 ? 1 : 0;       
            return await IsVipAsync(groupId)
                ? ($"update {FullName} set EndDate = {SqlDateAdd("month", month, "EndDate")}, UserId = {userId}, " +
                  $"IncomeDay = (IncomeDay * {SqlDateDiff("MONTH", SqlDateTime, "EndDate")} + {payMoney}) /(({SqlDateDiff("MONTH", SqlDateTime, "EndDate")} + {month} + 0.0000001)*1.0)," +
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
            => ChangeGroupAsync(groupId, newGroupId, qq).GetAwaiter().GetResult();

        public static async Task<int> ChangeGroupAsync(long groupId, long newGroupId, long qq)
        {
            return await ExecAsync($"exec sz84_robot..sp_ChangeVIP {groupId}, {newGroupId}, {qq}, {BotInfo.SystemUid}");
        }

        public static int RestDays(long groupId)
            => RestDaysAsync(groupId).GetAwaiter().GetResult();

        public static async Task<int> RestDaysAsync(long groupId)
        {
            return await GetIntAsync(SqlDateDiff("day", SqlDateTime, "EndDate"), groupId);
        }
 
        // 是否年费版
        public static bool IsYearVIP(long groupId)
            => IsYearVIPAsync(groupId).GetAwaiter().GetResult();

        public static async Task<bool> IsYearVIPAsync(long groupId)
        {
            return await GetBoolAsync("IsYearVip", groupId);
        }

        public static bool IsVip(long groupId)
            => IsVipAsync(groupId).GetAwaiter().GetResult();

        public static async Task<bool> IsVipAsync(long groupId)
        {
            return await ExistsAsync(groupId);
        }

        public static bool IsForever(long groupId)
            => IsForeverAsync(groupId).GetAwaiter().GetResult();

        public static async Task<bool> IsForeverAsync(long groupId)
        {
            return await RestDaysAsync(groupId) > 3650;
        }

        //是否开通过VIP
        public static bool IsVipOnce(long groupId)
            => IsVipOnceAsync(groupId).GetAwaiter().GetResult();

        public static async Task<bool> IsVipOnceAsync(long groupId)
        {
            return await Income.IsVipOnceAsync(groupId);
        }

        public static bool IsClientVip(long qq)
            => IsClientVipAsync(qq).GetAwaiter().GetResult();

        public static async Task<bool> IsClientVipAsync(long qq)
        {
            return await ExistsFieldAsync("UserId", qq.ToString());
        }

        public static int RestMonths(long groupId)
            => RestMonthsAsync(groupId).GetAwaiter().GetResult();

        public static async Task<int> RestMonthsAsync(long groupId)
        {
            return await GetIntAsync(SqlDateDiff("MONTH", SqlDateTime, "EndDate"), groupId);
        }
    }
}
