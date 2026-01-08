using Microsoft.Data.SqlClient;
using BotWorker.BotWorker.Common.Exts;
using BotWorker.Core.MetaDatas;
using BotWorker.Bots.Users;
using BotWorker.Bots.Entries;

namespace BotWorker.Domain.Entities.Office
{
    public class Partner : MetaData<Partner>
    {
        public override string TableName => "Partner";
        public override string KeyField => "UserID";


        public static string BecomePartner(long userId)
        {
            var incomeTotle = Income.Total(userId);

            if (IsPartner(userId))
                return "您已经是我们尊贵的合伙人";

            if (incomeTotle < 1000)
                return $"您的总消费金额{incomeTotle}不足1000元";

            int i = Append(userId);
            if (i == -1)
                return RetryMsg;

            return $"恭喜你已经成为我司尊贵的合伙人。";
        }

        public static int Append(long userId, long refUserId = 0)
        {
            return Exec(SqlInsert([
                                    new Cov("UserId", userId),
                                    new Cov("refUserId", refUserId)
                                  ]));
        }


        public static string GetCreditToday(long qq)
        {
            string sql = $"select top 10 a.UserId, SUM(abs(CreditAdd)) * 6 /1000 as partner_credit from {CreditLog.FullName} a inner join {UserInfo.FullName} b on a.UserId = b.UserId "
                         + $"where datediff(day, a.InsertDate, getdate()) = 0 and b.credit_freeze = 1 and partner_qq = {qq} and a.InsertDate > b.BindDateHome "
                         + $"and (CreditInfo like '%猜大小%' or CreditInfo like '%三公%' or CreditInfo like '%抽奖%' or CreditInfo like '%猜拳%' or CreditInfo like '%猜数字%')"
                         + $"group by a.UserId order by partner_credit desc";
            string res = QueryRes(sql, "{i} {0} {1}\n");
            sql = $"select SUM(abs(CreditAdd)) * 6 / 1000 as partner_credit from {CreditLog.FullName} a inner join  {UserInfo.FullName} b on a.UserId = b.UserId "
                  + $"where datediff(day, a.InsertDate, getdate()) = 0 and b.IsSuper = 1 and partner_qq = {qq} and a.InsertDate > b.BinDate "
                  + $"and (CreditInfo like '%猜大小%' or CreditInfo like '%三公%' or CreditInfo like '%抽奖%' or CreditInfo like '%猜拳%' or CreditInfo like '%猜数字%')";
            res += $"今日合计：{Query(sql)}";
            return res;
        }

        public static string GetSettleRes(long botUin, long groupId, string groupName, long qq, string name)
        {
            if (!IsPartner(qq))
                return "此功能仅合伙人可用";

            string res = Query($"select sum(partner_credit) as res from sz84_robot..robot_credit_day where partner_qq = {qq} and is_settle = 0");
            if (res == "")
                return "没有需要结算的流水";

            long partnerCredit = long.Parse(res);
            long credit_value = UserInfo.GetCredit(qq);
            var sql = UserInfo.SqlAddCredit(botUin, groupId, qq, partnerCredit);
            var sql2 = CreditLog.SqlHistory(botUin, groupId, groupName, qq, name, partnerCredit, "流水结算");
            var sql3 = GetSettleSql(qq);
            int i = ExecTrans(sql, sql2, sql3); 
            if (i == -1)
                return RetryMsg;

            int settle_id = Query($"select max(Id) as res from {CreditLog.FullName} where userId = {qq} and CreditInfo = '流水结算'").AsInt();
            i = Exec($"update sz84_robot..robot_credit_day set settle_id = {settle_id} where settle_id = {qq}");
            if (i == -1)
                ErrorMessage("更新结算ID出错");
            return $"结算成功\n+{partnerCredit}分，累计：{credit_value}";
        }

        public static (string, SqlParameter[]) GetSettleSql(long qq)
        {
            string sql = $"update sz84_robot..robot_credit_day set settle_date = getdate(), is_settle = 1, settle_id = @qq where is_settle = 0 and partner_qq = @qq";
            SqlParameter[] parameters = [new("@qq", qq)];
            return (sql, parameters);
        }

        public static bool IsPartner(long userId)
        {
            if (userId == 0) return false;
            return CountWhere($"UserId = {userId} and IsValid = 1") > 0;
        }

        public static bool IsNotPartner(long qq) => !IsPartner(qq);

        // 流水 只能查询未结算的
        public static string GetCreditList(long qq)
        {
            if (!IsPartner(qq))
                return "此功能仅合伙人可用";

            string res = QueryRes($"select top 7 MONTH(credit_day)*100 + DAY(credit_day)  as c_day, count(UserId) as c_client, SUM(partner_credit) as partner_credit " +
                                  $"from sz84_robot..robot_credit_day where partner_qq = {qq} and is_settle = 0 group by credit_day order by credit_day desc", "{0} {1}人 {2}分\n");
            res += QueryRes($"select '合计' as c_day,count(distinct UserId) as c_client, SUM(partner_credit) as partner_credit from sz84_robot..robot_credit_day " +
                            $"where partner_qq = {qq} and is_settle = 0",
                            "{0} {1}人 {2}分");
            return res;
        }
    }
}
