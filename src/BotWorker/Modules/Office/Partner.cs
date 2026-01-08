using System.Data;
using BotWorker.Common.Extensions;
using BotWorker.Domain.Entities;
using BotWorker.Infrastructure.Persistence.Database;

namespace BotWorker.Modules.Office
{
    public class Partner : MetaData<Partner>
    {
        public override string TableName => "Partner";
        public override string KeyField => "UserID";


        public static async Task<string> BecomePartnerAsync(long userId)
        {
            var incomeTotle = Income.Total(userId);

            if (IsPartner(userId))
                return "您已经是我们尊贵的合伙人";

            if (incomeTotle < 1000)
                return $"您的总消费金额{incomeTotle}不足1000元";

            int i = await AppendAsync(userId);
            if (i == -1)
                return RetryMsg;

            return $"恭喜你已经成为我司尊贵的合伙人。";
        }

        public static string BecomePartner(long userId)
        {
            return BecomePartnerAsync(userId).GetAwaiter().GetResult();
        }

        public static async Task<int> AppendAsync(long userId, long refUserId = 0)
        {
            return await ExecAsync(SqlInsert([
                                    new Cov("UserId", userId),
                                    new Cov("refUserId", refUserId)
                                  ]));
        }

        public static int Append(long userId, long refUserId = 0)
        {
            return AppendAsync(userId, refUserId).GetAwaiter().GetResult();
        }


        public static async Task<string> GetCreditTodayAsync(long qq)
        {
            string sql = $"select top 10 a.UserId, SUM(abs(CreditAdd)) * 6 /1000 as partner_credit from {CreditLog.FullName} a inner join {UserInfo.FullName} b on a.UserId = b.UserId "
                         + $"where datediff(day, a.InsertDate, getdate()) = 0 and b.credit_freeze = 1 and partner_qq = {qq} and a.InsertDate > b.BindDateHome "
                         + $"and (CreditInfo like '%猜大小%' or CreditInfo like '%三公%' or CreditInfo like '%抽奖%' or CreditInfo like '%猜拳%' or CreditInfo like '%猜数字%')"
                         + $"group by a.UserId order by partner_credit desc";
            string res = await QueryResAsync(sql, "{i} {0} {1}\n");
            sql = $"select SUM(abs(CreditAdd)) * 6 / 1000 as partner_credit from {CreditLog.FullName} a inner join  {UserInfo.FullName} b on a.UserId = b.UserId "
                  + $"where datediff(day, a.InsertDate, getdate()) = 0 and b.IsSuper = 1 and partner_qq = {qq} and a.InsertDate > b.BinDate "
                  + $"and (CreditInfo like '%猜大小%' or CreditInfo like '%三公%' or CreditInfo like '%抽奖%' or CreditInfo like '%猜拳%' or CreditInfo like '%猜数字%')";
            res += $"今日合计：{await QueryScalarAsync<string>(sql)}";
            return res;
        }

        public static string GetCreditToday(long qq) => GetCreditTodayAsync(qq).GetAwaiter().GetResult();

        public static async Task<string> GetSettleResAsync(long botUin, long groupId, string groupName, long qq, string name)
        {
            if (!IsPartner(qq))
                return "此功能仅合伙人可用";

            string res = QueryScalar<string>($"select sum(partner_credit) as res from sz84_robot..robot_credit_day where partner_qq = {qq} and is_settle = 0") ?? "";
            if (res == "")
                return "没有需要结算的流水";

            long partnerCredit = long.Parse(res);
            
            using var trans = await BeginTransactionAsync();
            try
            {
                var addRes = await UserInfo.AddCreditAsync(botUin, groupId, groupName, qq, name, partnerCredit, "流水结算", trans);
                if (addRes.Result == -1) throw new Exception("结算积分失败");

                var (sql3, paras3) = GetSettleSql(qq, addRes.LogId);
                await ExecAsync(sql3, trans, paras3);

                await trans.CommitAsync();

                UserInfo.SyncCacheField(qq, groupId, "Credit", addRes.CreditValue);
                return $"结算成功\n+{partnerCredit}分，累计：{addRes.CreditValue}";
            }
            catch (Exception ex)
            {
                await trans.RollbackAsync();
                Console.WriteLine($"[GetSettleRes Error] {ex.Message}");
                return RetryMsg;
            }
        }

        public static (string, IDataParameter[]) GetSettleSql(long qq, int settle_id = 0)
        {
            string sql = $"update sz84_robot..robot_credit_day set settle_date = getdate(), is_settle = 1, settle_id = @settle_id where is_settle = 0 and partner_qq = @qq";
            IDataParameter[] parameters = [
                DbProviderFactory.CreateParameter("@qq", qq),
                DbProviderFactory.CreateParameter("@settle_id", settle_id == 0 ? (object)qq : settle_id)
            ];
            return (sql, parameters);
        }

        public static async Task<bool> IsPartnerAsync(long userId)
        {
            if (userId == 0) return false;
            return await CountWhereAsync($"UserId = {userId} and IsValid = 1") > 0;
        }

        public static bool IsPartner(long userId) => IsPartnerAsync(userId).GetAwaiter().GetResult();

        public static async Task<bool> IsNotPartnerAsync(long qq) => !await IsPartnerAsync(qq);

        public static bool IsNotPartner(long qq) => IsNotPartnerAsync(qq).GetAwaiter().GetResult();

        // 流水 只能查询未结算的
        public static async Task<string> GetCreditListAsync(long qq)
        {
            if (!await IsPartnerAsync(qq))
                return "此功能仅合伙人可用";

            string res = await QueryResAsync($"select top 7 MONTH(credit_day)*100 + DAY(credit_day)  as c_day, count(UserId) as c_client, SUM(partner_credit) as partner_credit " +
                                  $"from sz84_robot..robot_credit_day where partner_qq = {qq} and is_settle = 0 group by credit_day order by credit_day desc", "{0} {1}人 {2}分\n");
            res += await QueryResAsync($"select '合计' as c_day,count(distinct UserId) as c_client, SUM(partner_credit) as partner_credit from sz84_robot..robot_credit_day " +
                            $"where partner_qq = {qq} and is_settle = 0",
                            "{0} {1}人 {2}分");
            return res;
        }

        public static string GetCreditList(long qq) => GetCreditListAsync(qq).GetAwaiter().GetResult();
    }
}
