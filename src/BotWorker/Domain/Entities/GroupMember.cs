using System.Text.RegularExpressions;
using System.Threading.Tasks;
using Microsoft.Data.SqlClient;

namespace BotWorker.Domain.Entities
{
    public partial class GroupMember : MetaData<GroupMember>
    {
        public override string TableName => "GroupMember";
        public override string KeyField => "GroupId";
        public override string KeyField2 => "UserId";

        //充值/扣除 积分 金币 黑金币 紫币 游戏币等 (异步重构版)
        public static async Task<string> AddCoinsResAsync(long botUin, long groupId, string groupName, long qq, string name, string cmdName, string cmdPara, string cmdPara2, string cmdPara3)
        {
            if (!UserInfo.IsOwner(groupId, qq))
                return $"您无权{cmdName}{cmdPara}";

            if (!cmdPara3.IsNum())
                return "数量不正确";

            long coins_oper = long.Parse(cmdPara3);
            long coins_qq = long.Parse(cmdPara2);

            if ((cmdPara == "本群积分") | (cmdPara == "积分"))
                cmdPara = "群积分";

            int coins_type = CoinsLog.conisNames.IndexOf(cmdPara);
            if (coins_type == -1) return "货币类型不正确";

            long minus_credit = coins_oper;
            long credit_group = groupId;

            if (coins_type == (int)CoinsLog.CoinsType.groupCredit && !await GroupInfo.GetIsCreditAsync(groupId))
                return $"没有开启本群积分";

            long credit_value = await UserInfo.GetCreditAsync(credit_group, qq);

            if (cmdName == "充值")
            {
                if (credit_value < minus_credit)
                    return $"您有{credit_value}分不足{minus_credit}，请先兑换";
            }
            else //扣除
            {
                long coins_value = await GetCoinsAsync(coins_type, groupId, coins_qq);
                if (coins_value < coins_oper)
                    return $"[@:{coins_qq}]{cmdPara}{coins_value}不足{coins_oper}，无法扣除";

                minus_credit = -minus_credit;
                coins_oper = -coins_oper;
            }

            // 使用事务进行异步链式调用
            using var trans = await BeginTransactionAsync();
            try
            {
                // 1. 扣除/增加发送者积分
                var res1 = await UserInfo.AddCreditAsync(botUin, credit_group, groupName, qq, name, -minus_credit, $"{cmdName}{cmdPara}*{coins_oper}", trans);
                
                // 2. 增加/扣除目标用户金币
                var res2 = await AddCoinsAsync(botUin, groupId, groupName, coins_qq, "", coins_type, coins_oper, $"{cmdName}{cmdPara}*{coins_oper}", trans);

                await trans.CommitAsync();

                return $"{cmdName}{cmdPara}：{Math.Abs(coins_oper)}成功！\n[@:{coins_qq}]{cmdPara}:{res2.CoinsValue}\n您：{-minus_credit}分，累计：{res1.CreditValue}";
            }
            catch (Exception ex)
            {
                await trans.RollbackAsync();
                Console.WriteLine($"[AddCoinsRes Error] {ex.Message}");
                return RetryMsg;
            }
        }

        //充值/扣除 积分 金币 黑金币 紫币 游戏币等
        public static string AddCoinsRes(long botUin, long groupId, string groupName, long qq, string name, string cmdName, string cmdPara, string cmdPara2, string cmdPara3)
        {
            return AddCoinsResAsync(botUin, groupId, groupName, qq, name, cmdName, cmdPara, cmdPara2, cmdPara3).GetAwaiter().GetResult();
        }

        public static async Task<string> ExchangeCoinsAsync(long botUin, long groupId, string groupName, long qq, string name, int coins_type, string cmdName, string cmdPara, long minus_credit, long coins_oper, long coins_qq)
        {
            long credit_group = groupId; // 简化处理，通常积分群就是当前群
            long credit_value = await UserInfo.GetCreditAsync(credit_group, qq);
            if (coins_type == (int)CoinsLog.CoinsType.groupCredit)
            { 
                if (!await GroupInfo.GetIsCreditAsync(groupId))
                    return $"没有开启本群积分";
            }

            long minus_value = minus_credit;
            if (cmdName == "充值")
            {
                if (credit_value < minus_value)
                    return $"您有{credit_value}分不足{minus_value}，请先兑换";
            }
            else //扣除
            {
                long coins_value = await GetCoinsAsync(coins_type, groupId, coins_qq);
                if (coins_value < minus_value)
                    return $"[@:{coins_qq}]{cmdPara}{coins_value}不足{coins_oper}，无法扣除";

                minus_credit = -minus_credit;
                coins_oper = -coins_oper;
            }
            credit_value -= minus_credit;

            using var trans = await BeginTransactionAsync();
            try
            {
                // 1. 处理积分
                var addRes = await UserInfo.AddCreditAsync(botUin, credit_group, groupName, qq, name, -minus_credit, $"{cmdName}{cmdPara}*{coins_oper}", trans);
                if (addRes.Result == -1) throw new Exception("积分操作失败");

                // 2. 处理金币
                var (sqlPlusCoins, parasPlus) = SqlPlus(CoinsLog.conisFields[coins_type], coins_oper, groupId, coins_qq);
                await ExecAsync((sqlPlusCoins, parasPlus), trans);

                // 3. 金币日志
                var (sqlCoinsHis, parasHis, coins_last) = CoinsLog.SqlCoins(botUin, groupId, groupName, coins_qq, "", coins_type, coins_oper, $"{cmdName}{cmdPara}*{coins_oper}");
                await ExecAsync((sqlCoinsHis, parasHis), trans);

                await trans.CommitAsync();

                // 同步更新缓存中的积分和对应金币字段
                UserInfo.SyncCacheField(qq, groupId, "Credit", addRes.CreditValue);
                SyncCacheField(coins_qq, groupId, CoinsLog.conisFields[coins_type], coins_last);

                return $"{cmdName}{cmdPara}：{coins_oper}成功！\n[@:{coins_qq}]{cmdPara}:{coins_last}\n您：{-minus_credit}分，累计：{addRes.CreditValue}";
            }
            catch (Exception ex)
            {
                await trans.RollbackAsync();
                Console.WriteLine($"[ExchangeCoins Error] {ex.Message}");
                return RetryMsg;
            }
        }

        // 本群积分
        public static async Task<long> GetGroupCreditAsync(long groupId, long qq)
        {
            return await GetCoinsAsync((int)CoinsLog.CoinsType.groupCredit, groupId, qq);
        }

        // 金币余额
        public static async Task<long> GetCoinsAsync(int coinsType, long groupId, long qq)
        {
            return await GetLongAsync(CoinsLog.conisFields[coinsType], groupId, qq);
        }

        public static long GetCoins(int coinsType, long groupId, long qq)
        {
            return GetCoinsAsync(coinsType, groupId, qq).GetAwaiter().GetResult();
        }

        // 金币余额
        public static async Task<long> GetGoldCoinsAsync(long groupId, long qq)
        {
            return await GetLongAsync("GoldCoins", groupId, qq);
        }

        // 紫币
        public static async Task<long> GetPurpleCoinsAsync(long groupId, long qq)
        {
            return await GetLongAsync("PurpleCoins", groupId, qq);
        }

        // 黑金币
        public static async Task<long> GetBlackCoinsAsync(long groupId, long qq)
        {
            return await GetLongAsync("BlackCoins", groupId, qq);
        }

        // 游戏币

        public static async Task<long> GetGameCoinsAsync(long groupId, long qq)
        {
            return await GetLongAsync("GameCoins", groupId, qq);
        }

        // 加金币/黑金币/紫币/游戏币
        public static int AddCoins(long botUin, long groupId, string groupName, long qq, string name, int coinsType, long coinsAdd, ref long coinsValue, string coinsInfo)
        {
            var res = AddCoinsAsync(botUin, groupId, groupName, qq, name, coinsType, coinsAdd, coinsInfo).GetAwaiter().GetResult();
            if (res.Result == 0)
            {
                coinsValue = res.CoinsValue;
            }
            return res.Result;
        }

        // 加金币/黑金币/紫币/游戏币 (异步事务版)
        public static async Task<(int Result, long CoinsValue)> AddCoinsAsync(long botUin, long groupId, string groupName, long qq, string name, int coinsType, long coinsAdd, string coinsInfo, SqlTransaction? trans = null)
        {
            if (!await ExistsAsync(groupId, qq) && await AppendAsync(groupId, qq, name) == -1)
                return (-1, 0);

            long coinsValue = 0;
            bool isNewTrans = false;
            if (trans == null)
            {
                trans = await BeginTransactionAsync();
                isNewTrans = true;
            }

            try
            {
                var (sql, paras) = SqlPlus(CoinsLog.conisFields[coinsType], coinsAdd, groupId, qq);
                await ExecAsync(sql, trans, paras);
                
                // 注意：CoinsLog.SqlCoins 内部可能包含 ref long coinsValue，这里需要支持异步版本
                await CoinsLog.AddLogAsync(botUin, groupId, groupName, qq, name, coinsType, coinsAdd, coinsInfo, trans);
                
                coinsValue = await GetCoinsAsync(coinsType, groupId, qq); // 事务中获取最新值

                if (isNewTrans) await trans.CommitAsync();
                SyncCacheField(qq, groupId, CoinsLog.conisFields[coinsType], coinsValue);
                return (0, coinsValue);
            }
            catch
            {
                if (isNewTrans) await trans.RollbackAsync();
                return (-1, 0);
            }
        }

        // 虚拟币转账 (异步事务版)
        public static async Task<(int Result, long SenderCoins, long ReceiverCoins)> TransferCoinsAsync(long botUin, long groupId, string groupName, long qq, string name, long qqTo, string nameTo, int coinsType, long coinsMinus, long coinsAdd, string transferInfo)
        {
            using var trans = await BeginTransactionAsync();
            try
            {
                var res1 = await AddCoinsAsync(botUin, groupId, groupName, qq, name, coinsType, -coinsMinus, $"{transferInfo}转出:{qqTo}", trans);
                var res2 = await AddCoinsAsync(botUin, groupId, groupName, qqTo, nameTo, coinsType, coinsAdd, $"{transferInfo}转入:{qq}", trans);

                await trans.CommitAsync();
                return (0, res1.CoinsValue, res2.CoinsValue);
            }
            catch
            {
                await trans.RollbackAsync();
                return (-1, 0, 0);
            }
        }

        // 扣除金币
        public static int MinusCoins(long botUin, long groupId, string groupName, long qq, string name, int coinsType, long coinsMinus, ref long coinsValue, string coinsInfo)
        {
            return AddCoins(botUin, groupId, groupName, qq, name, coinsType, -(coinsMinus), ref coinsValue, coinsInfo);
        }

        // 虚拟币转账
        public static int TransferCoins(long botUin, long groupId, string groupName, long qq, string name, long qqTo, int coinsType, long coinsMinus, long coinsAdd, ref long coinsValue, ref long coinsValue2)
        {
            var res = TransferCoinsAsync(botUin, groupId, groupName, qq, name, qqTo, "", coinsType, coinsMinus, coinsAdd, "").GetAwaiter().GetResult();
            if (res.Result == 0)
            {
                coinsValue = res.SenderCoins;
                coinsValue2 = res.ReceiverCoins;
            }
            return res.Result;
        }

        public static (string, SqlParameter[]) SqlSaveCredit(long groupId, long userId, long creditSave)
        {
            return SqlSetValues($"GroupCredit = GroupCredit - ({creditSave}), SaveCredit = ISNULL(SaveCredit, 0) + ({creditSave})", groupId, userId);
        }

        public static (string, SqlParameter[]) SqlAddCredit(long groupId, long userId, long creditAdd)
            => SqlAddCreditAsync(groupId, userId, creditAdd).GetAwaiter().GetResult();

        public static async Task<(string, SqlParameter[])> SqlAddCreditAsync(long groupId, long userId, long creditAdd)
        {
            return await MetaData<GroupMember>.ExistsAsync(groupId, userId)
                ? SqlPlus("GroupCredit", creditAdd, groupId, userId)
                : SqlInsert([
                                new Cov("GroupId", groupId),
                                new Cov("UserId", userId),
                                new Cov("GroupCredit", creditAdd),
                            ]);
        }

        public static int Append(long groupId, long userId, string name, string displayName = "", long groupCredit = 50, string confirmCode = "")
            => AppendAsync(groupId, userId, name, displayName, groupCredit, confirmCode).GetAwaiter().GetResult();

        // 添加群成员 (异步版本)
        public static async Task<int> AppendAsync(long groupId, long userId, string name, string displayName = "", long groupCredit = 50, string confirmCode = "")
        {
            if (userId.In(2107992324, 3677524472, 3662527857, 2174158062, 2188157235, 3375620034, 1611512438, 3227607419, 3586811032,
                3835195413, 3527470977, 3394199803, 2437953621, 3082166471, 2375832958, 1807139582, 2704647312, 1420694846, 3788007880)) return 0;

            var sql = await MetaData<GroupMember>.ExistsAsync(groupId, userId)
                            ? SqlSetValues($"UserName = {name.Quotes()}, DisplayName = {displayName.Quotes()}, ConfirmCode = {confirmCode.Quotes()}, Status = 1", groupId, userId)
                            : SqlInsert(new List<Cov> {
                                            new Cov("GroupId", groupId),
                                            new Cov("UserId", userId),
                                            new Cov("UserName", name),
                                            new Cov("DisplayName", displayName),
                                            new Cov("GroupCredit", groupCredit),
                                            new Cov("ConfirmCode", confirmCode),
                                        });
            return await ExecAsync(sql);
        }

        //上下分 (异步重构版)
        public static async Task<string> GetShangFenAsync(long botUin, long groupId, string groupName, long userId, string cmdName, string cmdPara)
        {
            if (!await GroupInfo.IsOwnerAsync(groupId, userId) || !await BotInfo.IsAdminAsync(botUin, userId))
                return OwnerOnlyMsg;

            if (await Income.TotalAsync(userId) < 400)            
                return "您无权使用此命令，请联系客服";
            
            if (await BotInfo.GetIsCreditAsync(botUin) || await GroupInfo.GetIsCreditAsync(groupId))
            {
                long creditQQ = 0;
                string regexShangFen;
                if (cmdPara.IsMatch(Regexs.CreditParaAt))
                    regexShangFen = Regexs.CreditParaAt;
                else if (cmdPara.IsMatch(Regexs.CreditParaAt2))
                    regexShangFen = Regexs.CreditParaAt2;
                else if (cmdPara.IsMatch(Regexs.CreditPara))
                    regexShangFen = Regexs.CreditPara;
                else
                    return $"格式：{cmdName} + QQ + 数量\n例如：{cmdName} {{客服QQ}} 5000";

                long creditAdd = 0;

                //分析命令
                foreach (Match match in cmdPara.Matches(regexShangFen))
                {
                    creditQQ = match.Groups["UserId"].Value.AsLong();
                    creditAdd = match.Groups["credit"].Value.AsLong();
                }

                if (creditAdd < 10)
                    return $"至少{(cmdName == "上分" ? "上" : "下")}10分";

                var creditValue = await UserInfo.GetCreditAsync(groupId, creditQQ);

                if (cmdName == "下分")
                {
                    if (creditValue < creditAdd)
                        return $"对方只有{creditValue}分";
                    creditAdd = -creditAdd;
                }

                var res = await UserInfo.AddCreditAsync(botUin, groupId, groupName, creditQQ, "", creditAdd, cmdName);
                
                return res.Result == -1
                    ? RetryMsg
                    : $"[@:{creditQQ}] {cmdName}成功！\n积分：{creditAdd}，累计：{res.CreditValue}";
            }
            else
                return $"此群未开通本群积分，不能上下分";
        }

        // 获取签到次数
        public static async Task<int> GetSignTimesAsync(long groupId, long userId)
        {
            return await GetIntAsync("SignTimes", groupId, userId);
        }

        public static int GetSignTimes(long groupId, long userId)
            => GetSignTimesAsync(groupId, userId).GetAwaiter().GetResult();

        // 获取签到列表
        public static async Task<string> GetSignListAsync(long groupId, int topN = 10)
        {
            return await QueryResAsync($"select top {topN} UserId, SignTimes, SignLevel from {FullName} " +
                                     $"where GroupId = {groupId} and SignTimes > 0 order by SignTimes desc, SignLevel desc",
                                     "【第{i}名】 [@:{0}] 连续签到：{1}天(LV{2})\n");
        }

        public static string GetSignList(long groupId, int topN = 10)
            => GetSignListAsync(groupId, topN).GetAwaiter().GetResult();

        // 积分提现 (异步版)
        public static async Task<string> WithdrawCreditAsync(long botUin, long groupId, string groupName, long userId, string name, long withdrawAmount)
        {
            if (withdrawAmount < 100) return "提现金额不能低于100";

            long creditValue = await UserInfo.GetCreditAsync(groupId, userId);
            if (creditValue < withdrawAmount) return $"您的积分{creditValue}不足{withdrawAmount}";

            using var trans = await BeginTransactionAsync();
            try
            {
                // 1. 扣除积分
                var res = await UserInfo.AddCreditAsync(botUin, groupId, groupName, userId, name, -withdrawAmount, "积分提现", trans);
                if (res.Result == -1) throw new Exception("扣除积分失败");

                // 2. 增加余额 (此处假设 UserInfo.AddBalanceAsync 已存在或通过 SqlPlus 处理)
                var (sqlBalance, parasBalance) = UserInfo.SqlPlus("Balance", (decimal)withdrawAmount / 100, userId);
                await ExecAsync(sqlBalance, trans, parasBalance);

                await trans.CommitAsync();

                UserInfo.SyncCacheField(userId, groupId, "Credit", res.CreditValue);
                // UserInfo.SyncCacheField(userId, "Balance", ...); // 余额同步

                return $"✅ 提现成功！\n提现积分：{withdrawAmount}\n到账余额：{(decimal)withdrawAmount / 100:N2}\n剩余积分：{res.CreditValue}";
            }
            catch (Exception ex)
            {
                await trans.RollbackAsync();
                Console.WriteLine($"[WithdrawCredit Error] {ex.Message}");
                return RetryMsg;
            }
        }

        // 金币排行榜
        public static async Task<string> GetCoinsRankingAsync(long groupId, long userId)
        {
            long order = await CountWhereAsync($"GroupId = {groupId} AND GoldCoins > (select GoldCoins from {FullName} where GroupId = {groupId} and UserId = {userId})") + 1;
            return order.ToString();
        }

        public static long GetCoinsRanking(long groupId, long userId)
            => long.Parse(GetCoinsRankingAsync(groupId, userId).GetAwaiter().GetResult());

        public static async Task<string> GetCoinsRankingAllAsync(long userId)
        {
            long order = await CountWhereAsync($"GoldCoins > (select sum(GoldCoins) from {FullName} where UserId = {userId})") + 1;
            return order.ToString();
        }

        public static long GetCoinsRankingAll(long userId)
            => long.Parse(GetCoinsRankingAllAsync(userId).GetAwaiter().GetResult());

        // 获取签到日期差 (异步版)
        public static async Task<int> GetSignDateDiffAsync(long groupId, long userId)
        {
            return await GetIntAsync("DATEDIFF(day, ISNULL(SignDate, '2000-01-01'), GETDATE())", groupId, userId);
        }

        public static int GetSignDateDiff(long groupId, long userId)
            => GetSignDateDiffAsync(groupId, userId).GetAwaiter().GetResult();

        // 是否签到 (异步版)
        public static async Task<bool> IsSignInAsync(long groupId, long userId)
        {
            return await GetBoolAsync("DATEDIFF(day, ISNULL(SignDate, '2000-01-01'), GETDATE()) = 0", groupId, userId);
        }

        public static bool IsSignIn(long groupId, long userId)
            => IsSignInAsync(groupId, userId).GetAwaiter().GetResult();

        // 更新签到信息 SQL
        public static (string sql, SqlParameter[] parameters) SqlUpdateSignInfo(long groupId, long userId, int signTimes, int signLevel)
        {
            return SqlUpdateWhere($"SignTimes = {signTimes}, SignLevel = {signLevel}, SignDate = GETDATE(), SignTimesAll = ISNULL(SignTimesAll, 0) + 1", $"GroupId = {groupId} AND UserId = {userId}");
        }
    }
}
