using BotWorker.Domain.Models.Messages.BotMessages;
using BotWorker.Domain.Entities;
using BotWorker.Modules.Office;
using BotWorker.Common.Extensions;
using BotWorker.Infrastructure.Persistence.ORM;

namespace BotWorker.Domain.Entities
{
    public partial class UserInfo : MetaDataGuid<UserInfo>
    {
        // 购买积分 (异步事务版)
        public static async Task<int> BuyCreditAsync(long botUin, long groupId, string groupName, long buyerQQ, string buyerName, decimal payMoney, long creditAdd, string payMethod, string trade, string memo, int insertBy)
        {
            using var trans = await BeginTransactionAsync();
            try
            {
                // 1. 记录收入
                var (sql, paras) = Income.SqlInsert(groupId, creditAdd, "积分", payMoney, payMethod, trade, memo, buyerQQ, insertBy);
                await ExecAsync(sql, trans, paras);

                // 2. 通用加积分函数 (含日志记录)
                var res = await UserInfo.AddCreditAsync(botUin, groupId, groupName, buyerQQ, buyerName, creditAdd, "买分", trans);
                if (res.Result == -1) throw new Exception("更新积分失败");

                await trans.CommitAsync();
                
                SyncCacheField(buyerQQ, groupId, "Credit", res.CreditValue);
                return 0;
            }
            catch (Exception ex)
            {
                await trans.RollbackAsync();
                Console.WriteLine($"[BuyCredit Error] {ex.Message}");
                return -1;
            }
        }

        // 购买积分
        public static int BuyCredit(long botUin, long groupId, string groupName, long buyerQQ, string buyerName, decimal payMoney, long creditAdd, string payMethod, string trade, string memo, int insertBy)
        {
            // 为了安全，同步版本也应该使用事务。这里使用 Wait() 简化重构。
            return BuyCreditAsync(botUin, groupId, groupName, buyerQQ, buyerName, payMoney, creditAdd, payMethod, trade, memo, insertBy).GetAwaiter().GetResult();
        }

        // 充值余额 (异步事务版)
        public static async Task<int> BuyBalanceAsync(long botUin, long groupId, string groupName, long buyerQQ, string buyerName, decimal payMoney, decimal balanceAdd, string payMethod, string trade, string memo, int insertBy)
        {
            using var trans = await BeginTransactionAsync();
            try
            {
                // 1. 记录收入
                var (sql1, paras1) = Income.SqlInsert(groupId, 1, "余额", payMoney, payMethod, trade, memo, buyerQQ, insertBy);
                await ExecAsync(sql1, trans, paras1);

                // 2. 增加余额 (含日志记录)
                var res = await UserInfo.AddBalanceAsync(botUin, groupId, groupName, buyerQQ, buyerName, balanceAdd, "充值余额", trans);
                if (res.Result == -1) throw new Exception("更新余额失败");

                await trans.CommitAsync();

                SyncCacheField(buyerQQ, "Balance", res.BalanceValue);
                return 0;
            }
            catch (Exception ex)
            {
                await trans.RollbackAsync();
                Console.WriteLine($"[BuyBalance Error] {ex.Message}");
                return -1;
            }
        }

        // 充值余额
        public static int BuyBalance(long botUin, long groupId, string groupName, long buyerQQ, string buyerName, decimal payMoney, decimal balanceAdd, string payMethod, string trade, string memo, int insertBy)
        {
            return BuyBalanceAsync(botUin, groupId, groupName, buyerQQ, buyerName, payMoney, balanceAdd, payMethod, trade, memo, insertBy).GetAwaiter().GetResult();
        }

        // 购买算力 (异步事务版)
        public static async Task<int> BuyTokensAsync(long botUin, long groupId, string groupName, long qqBuyer, string buyerName, decimal payMoney, long tokensAdd, string payMethod, string trade, string memo, int insertBy)
        {
            using var trans = await BeginTransactionAsync();
            try
            {
                // 1. 记录收入
                var (sql1, paras1) = Income.SqlInsert(groupId, tokensAdd, "TOKENS", payMoney, payMethod, trade, memo, qqBuyer, insertBy);
                await ExecAsync(sql1, trans, paras1);

                // 2. 增加算力 (含日志记录)
                var (sql2, paras2) = SqlAddTokens(qqBuyer, tokensAdd);
                var (sql3, paras3) = TokensLog.SqlLog(botUin, groupId, groupName, qqBuyer, buyerName, tokensAdd, "购买算力");
                await ExecAsync(sql2, trans, paras2);
                await ExecAsync(sql3, trans, paras3);

                await trans.CommitAsync();

                long tokensValue = GetLong("Tokens", qqBuyer) + tokensAdd;
                SyncCacheField(qqBuyer, "Tokens", tokensValue);
                return 0;
            }
            catch (Exception ex)
            {
                await trans.RollbackAsync();
                Console.WriteLine($"[BuyTokens Error] {ex.Message}");
                return -1;
            }
        }

        // 购买算力
        public static int BuyTokens(long botUin, long groupId, string groupName, long qqBuyer, string buyerName, decimal payMoney, long tokensAdd, string payMethod, string trade, string memo, int insertBy)
        {
            return BuyTokensAsync(botUin, groupId, groupName, qqBuyer, buyerName, payMoney, tokensAdd, payMethod, trade, memo, insertBy).GetAwaiter().GetResult();
        }

        // 使用余额购买积分 (异步事务版)
        public static async Task<string> GetBuyCreditAsync(BotMessage context, long botUin, long groupId, string groupName, long qq, string name, string cmdPara)
        {
            if (!context.Group.IsCreditSystem)
                return CreditSystemClosed;

            if (cmdPara == "")
                return "格式：买分 + 金额\n例如：买分 16.88\n价格：1R=1200分\n您的余额：{余额}";

            if (!cmdPara.IsDouble())
                return "金额不正确";

            decimal balanceMinus = cmdPara.AsDecimal();
            decimal balanceValue = GetBalance(qq);
            if (balanceMinus < 1)
                return "至少买1元";

            if (balanceMinus > balanceValue)
                return $"您的余额{balanceValue:N}不足{balanceMinus:N}";

            decimal balanceNew = balanceValue - balanceMinus;
            long creditAdd = Convert.ToInt32(balanceMinus * 1200);
            bool isPartner = Partner.IsPartner(qq);
            if (isPartner) creditAdd *= 2;

            using var trans = await BeginTransactionAsync();
            try
            {
                // 1. 扣除余额 (含日志记录)
                var (sql1, paras1) = SqlAddBalance(qq, -balanceMinus);
                var (sql2, paras2) = BalanceLog.SqlLog(botUin, groupId, groupName, qq, name, -balanceMinus, "买分");
                await ExecAsync(sql1, trans, paras1);
                await ExecAsync(sql2, trans, paras2);

                // 2. 通用加积分函数 (含日志记录)
                var res = await UserInfo.AddCreditAsync(botUin, groupId, groupName, qq, name, creditAdd, "买分", trans);
                if (res.Result == -1) throw new Exception("更新积分失败");

                await trans.CommitAsync();

                SyncCacheField(qq, "Balance", balanceNew);
                SyncCacheField(qq, groupId, "Credit", res.CreditValue);

                return $"✅ 买分成功！\n积分：+{creditAdd}，累计：{res.CreditValue}\n余额：-{balanceMinus:N}，累计：{balanceNew:N}";
            }
            catch (Exception ex)
            {
                await trans.RollbackAsync();
                Console.WriteLine($"[GetBuyCredit Error] {ex.Message}");
                return RetryMsg;
            }
        }

        // 使用余额购买积分
        public static string GetBuyCredit(BotMessage context, long botUin, long groupId, string groupName, long qq, string name, string cmdPara)
        {
            return GetBuyCreditAsync(context, botUin, groupId, groupName, qq, name, cmdPara).GetAwaiter().GetResult();
        }

        // 客服通过发IM消息给客户充值积分 (异步版)
        public static async Task<string> GetBuyCreditAsync(long botUin, long groupId, string groupName, long qq, string msgId, long buyerQQ, decimal incomeMoney, string payMethod, bool isPublic = false)
        {
            if (!BotInfo.IsSuperAdmin(qq))
                return "您不是管理员，无权充值积分";
            payMethod = payMethod switch
            {
                "qq" => "QQ红包",
                "wx" => "微信支付",
                "zfb" => "支付宝",
                "微信" => "微信支付",
                _ => "QQ红包"
            };

            if (isPublic && GetValue("MsgId", qq) == msgId)
                return $"重复消息{RetryMsg}";

            long creditValue = GetCredit(groupId, buyerQQ);
            long creditAdd = (long)Math.Round(incomeMoney * 1200, 0);
            if (Partner.IsPartner(buyerQQ))
            {
                if (GetIsSuper(buyerQQ))
                    creditAdd *= 2;
                else
                    creditAdd = (long)Math.Round(incomeMoney * 10000, 0);
            }

            return await BuyCreditAsync(botUin, groupId, groupName, buyerQQ, "", incomeMoney, creditAdd, payMethod, "", "", BotInfo.SystemUid) == -1
                ? RetryMsg
                : $"✅ 购买成功！\n{buyerQQ}积分：\n{creditValue}{(creditAdd > 0 ? $"+" : $"")}{creditAdd} = {GetCredit(groupId, buyerQQ)}";
        }

        //客服通过发IM消息给客户充值积分
        public static string GetBuyCredit(long botUin, long groupId, string groupName, long qq, string msgId, long buyerQQ, decimal incomeMoney, string payMethod, bool isPublic = false)
        {
            return GetBuyCreditAsync(botUin, groupId, groupName, qq, msgId, buyerQQ, incomeMoney, payMethod, isPublic).GetAwaiter().GetResult();
        }


    }
}
