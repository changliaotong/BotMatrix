using BotWorker.Domain.Models.Messages.BotMessages;
using BotWorker.Domain.Entities;
using BotWorker.Modules.Office;
using BotWorker.Common.Extensions;
using BotWorker.Infrastructure.Persistence.ORM;

namespace BotWorker.Domain.Entities
{
    public partial class UserInfo : MetaDataGuid<UserInfo>
    {
        // 购买积分
        public static int BuyCredit(long botUin, long groupId, string groupName, long buyerQQ, string buyerName, decimal payMoney, long creditAdd, string payMethod, string trade, string memo, int insertBy)
        {
            var sql = Income.SqlInsert(groupId, creditAdd, "积分", payMoney, payMethod, trade, memo, buyerQQ, insertBy);
            var sql2 = SqlAddCredit(botUin, groupId, buyerQQ, creditAdd);
            var sql3 = CreditLog.SqlHistory(botUin, groupId, groupName, buyerQQ, buyerName, creditAdd, "买分");
            int result = ExecTrans(sql, sql2, sql3);
            if (result == 0)
            {
                SyncCacheField(buyerQQ, groupId, "Credit", GetCredit(groupId, buyerQQ) + creditAdd);
            }
            return result;
        }

        // 充值余额
        public static int BuyBalance(long botUin, long groupId, string groupName, long buyerQQ, string buyerName, decimal payMoney, decimal balanceAdd, string payMethod, string trade, string memo, int insertBy)
        {
            var sql = Income.SqlInsert(groupId, 1, "余额", payMoney, payMethod, trade, memo, buyerQQ, insertBy);
            var sql2 = SqlAddBalance(buyerQQ, balanceAdd);
            var sql3 = BalanceLog.SqlLog(botUin, groupId, groupName, buyerQQ, buyerName, balanceAdd, "充值余额");
            int result = ExecTrans(sql, sql2, sql3);
            if (result == 0)
            {
                SyncCacheField(buyerQQ, "Balance", GetBalance(buyerQQ) + balanceAdd);
            }
            return result;
        }

        // 购买算力
        public static int BuyTokens(long botUin, long groupId, string groupName, long qqBuyer, string buyerName, decimal payMoney, long tokensAdd, string payMethod, string trade, string memo, int insertBy)
        {
            var sql = Income.SqlInsert(groupId, tokensAdd, "TOKENS", payMoney, payMethod, trade, memo, qqBuyer, insertBy);
            var sql2 = SqlAddTokens(qqBuyer, tokensAdd);
            var sql3 = TokensLog.SqlLog(botUin, groupId, groupName, qqBuyer, buyerName, tokensAdd, "购买算力");
            int result = ExecTrans(sql, sql2, sql3);
            if (result == 0)
            {
                SyncCacheField(qqBuyer, "Tokens", GetLong("Tokens", qqBuyer) + tokensAdd);
            }
            return result;
        }

        // 使用余额购买积分
        public static string GetBuyCredit(BotMessage context, long botUin, long groupId, string groupName, long qq, string name, string cmdPara)
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
            long creditValue = GetCredit(groupId, qq);
            long creditAdd = Convert.ToInt32(balanceMinus * 1200);
            bool isPartner = Partner.IsPartner(qq);
            if (isPartner) creditAdd *= 2;

            creditValue += creditAdd;
            //扣余额，加分
            var sql = SqlAddBalance(qq, -balanceMinus);
            var sql2 = BalanceLog.SqlLog(botUin, groupId, groupName, qq, name, -balanceMinus, "买分");
            var sql3 = SqlAddCredit(botUin, groupId, qq, creditAdd);
            var sql4 = CreditLog.SqlHistory(botUin, groupId, groupName, qq, name, creditAdd, "买分");

            int i = ExecTrans(sql, sql2, sql3, sql4);
            return i == -1
                ? RetryMsg
                : $"✅ 买分成功！\n积分：+{creditAdd}，累计：{creditValue}\n余额：-{balanceMinus:N}，累计：{balanceNew:N}";
        }

        //客服通过发IM消息给客户充值积分
        public static string GetBuyCredit(long botUin, long groupId, string groupName, long qq, string msgId, long buyerQQ, decimal incomeMoney, string payMethod, bool isPublic = false)
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

            return BuyCredit(botUin, groupId, groupName, buyerQQ, "", incomeMoney, creditAdd, payMethod, "", "", BotInfo.SystemUid) == -1
                ? RetryMsg
                : $"✅ 购买成功！\n{buyerQQ}积分：\n{creditValue}{(creditAdd > 0 ? $"+" : $"")}{creditAdd} = {GetCredit(groupId, buyerQQ)}";
        }


    }
}
