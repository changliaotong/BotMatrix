namespace BotWorker.Domain.Entities
{
    public partial class UserInfo
    {
        private static IIncomeRepository IncomeRepository => 
            BotMessage.ServiceProvider?.GetRequiredService<IIncomeRepository>() 
            ?? throw new InvalidOperationException("IIncomeRepository not registered");

        // 购买积分 (异步事务版)
        public static async Task<int> BuyCreditAsync(long botUin, long groupId, string groupName, long buyerQQ, string buyerName, decimal payMoney, long creditAdd, string payMethod, string trade, string memo, int insertBy)
        {
            using var wrapper = await Repository.BeginTransactionAsync();
            try
            {
                // 1. 记录收入
                await IncomeRepository.AddAsync(new Income
                {
                    GroupId = groupId,
                    GoodsCount = creditAdd,
                    GoodsName = "积分",
                    UserId = buyerQQ,
                    IncomeMoney = payMoney,
                    PayMethod = payMethod,
                    IncomeTrade = trade,
                    IncomeInfo = memo,
                    InsertBy = insertBy,
                    IncomeDate = DateTime.Now
                }, wrapper.Transaction);

                // 2. 通用加积分函数 (含日志记录)
                var res = await AddCreditAsync(botUin, groupId, groupName, buyerQQ, buyerName, creditAdd, "买分", wrapper.Transaction);
                if (res.Result == -1) throw new Exception("更新积分失败");

                await wrapper.CommitAsync();
                
                await SyncCacheFieldAsync(buyerQQ, "Credit", res.CreditValue);
                return 0;
            }
            catch (Exception ex)
            {
                await wrapper.RollbackAsync();
                Console.WriteLine($"[BuyCredit Error] {ex.Message}");
                return -1;
            }
        }

        // 充值余额 (异步事务版)
        public static async Task<int> BuyBalanceAsync(long botUin, long groupId, string groupName, long buyerQQ, string buyerName, decimal payMoney, decimal balanceAdd, string payMethod, string trade, string memo, int insertBy)
        {
            using var wrapper = await Repository.BeginTransactionAsync();
            try
            {
                // 1. 记录收入
                await IncomeRepository.AddAsync(new Income
                {
                    GroupId = groupId,
                    GoodsCount = 1,
                    GoodsName = "余额",
                    UserId = buyerQQ,
                    IncomeMoney = payMoney,
                    PayMethod = payMethod,
                    IncomeTrade = trade,
                    IncomeInfo = memo,
                    InsertBy = insertBy,
                    IncomeDate = DateTime.Now
                }, wrapper.Transaction);

                // 2. 增加余额 (含日志记录)
                var res = await AddBalanceAsync(botUin, groupId, groupName, buyerQQ, buyerName, balanceAdd, "充值余额", wrapper.Transaction);
                if (res.Result == -1) throw new Exception("更新余额失败");

                await wrapper.CommitAsync();

                await SyncCacheFieldAsync(buyerQQ, "Balance", res.BalanceValue);
                return 0;
            }
            catch (Exception ex)
            {
                await wrapper.RollbackAsync();
                Console.WriteLine($"[BuyBalance Error] {ex.Message}");
                return -1;
            }
        }

        // 购买算力 (异步事务版)
        public static async Task<int> BuyTokensAsync(long botUin, long groupId, string groupName, long qqBuyer, string buyerName, decimal payMoney, long tokensAdd, string payMethod, string trade, string memo, int insertBy)
        {
            using var wrapper = await Repository.BeginTransactionAsync();
            try
            {
                // 1. 记录收入
                await IncomeRepository.AddAsync(new Income
                {
                    GroupId = groupId,
                    GoodsCount = tokensAdd,
                    GoodsName = "TOKENS",
                    UserId = qqBuyer,
                    IncomeMoney = payMoney,
                    PayMethod = payMethod,
                    IncomeTrade = trade,
                    IncomeInfo = memo,
                    InsertBy = insertBy,
                    IncomeDate = DateTime.Now
                }, wrapper.Transaction);

                // 2. 增加算力 (含日志记录)
                var res = await AddTokensAsync(botUin, groupId, groupName, qqBuyer, buyerName, tokensAdd, "购买算力", wrapper.Transaction);
                if (res.Result == -1) throw new Exception("更新算力失败");

                await wrapper.CommitAsync();

                await SyncCacheFieldAsync(qqBuyer, "Tokens", res.TokensValue);
                return 0;
            }
            catch (Exception ex)
            {
                await wrapper.RollbackAsync();
                Console.WriteLine($"[BuyTokens Error] {ex.Message}");
                return -1;
            }
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
            decimal balanceValue = await GetBalanceAsync(qq);
            if (balanceMinus < 1)
                return "至少买1元";

            if (balanceMinus > balanceValue)
                return $"您的余额{balanceValue:N}不足{balanceMinus:N}";

            long creditAdd = Convert.ToInt32(balanceMinus * 1200);
            bool isPartner = await Partner.IsPartnerAsync(qq);
            if (isPartner) creditAdd *= 2;

            using var wrapper = await Repository.BeginTransactionAsync();
            try
            {
                // 1. 获取准确余额并锁定
                decimal balanceValueTrans = await GetBalanceForUpdateAsync(qq, wrapper.Transaction);
                if (balanceValueTrans < balanceMinus)
                {
                    await wrapper.RollbackAsync();
                    return $"您的余额{balanceValueTrans:N}不足{balanceMinus:N}";
                }
                decimal balanceNewTrans = balanceValueTrans - balanceMinus;

                // 2. 扣除余额 (含日志记录)
                var resBalance = await AddBalanceAsync(botUin, groupId, groupName, qq, name, -balanceMinus, "买分", wrapper.Transaction);
                if (resBalance.Result == -1) throw new Exception("更新余额失败");

                // 3. 增加积分 (含日志记录)
                var resCredit = await AddCreditAsync(botUin, groupId, groupName, qq, name, creditAdd, "买分", wrapper.Transaction);
                if (resCredit.Result == -1) throw new Exception("更新积分失败");

                await wrapper.CommitAsync();

                await SyncCacheFieldAsync(qq, "Balance", balanceNewTrans);
                await SyncCacheFieldAsync(qq, groupId, "Credit", resCredit.CreditValue);

                return $"✅ 买分成功！\n积分：+{creditAdd}，累计：{resCredit.CreditValue}\n余额：-{balanceMinus:N}，累计：{balanceNewTrans:N}";
            }
            catch (Exception ex)
            {
                await wrapper.RollbackAsync();
                Console.WriteLine($"[GetBuyCredit Error] {ex.Message}");
                return RetryMsg;
            }
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

            if (isPublic && await GetValueAsync("MsgId", qq) == msgId)
                return $"重复消息{RetryMsg}";

            long creditValue = await GetCreditAsync(groupId, buyerQQ);
            long creditAdd = (long)Math.Round(incomeMoney * 1200, 0);
            if (await Partner.IsPartnerAsync(buyerQQ))
            {
                if (await GetIsSuperAsync(buyerQQ))
                    creditAdd *= 2;
                else
                    creditAdd = (long)Math.Round(incomeMoney * 10000, 0);
            }

            return await BuyCreditAsync(botUin, groupId, groupName, buyerQQ, "", incomeMoney, creditAdd, payMethod, "", "", BotInfo.SystemUid) == -1
                ? RetryMsg
                : $"✅ 购买成功！\n{buyerQQ}积分：\n{creditValue}{(creditAdd > 0 ? $"+" : $"")}{creditAdd} = {await GetCreditAsync(groupId, buyerQQ)}";
        }


    }
}
