using System.Data;

namespace BotWorker.Domain.Entities
{
    public partial class UserInfo : MetaDataGuid<UserInfo>
    {
        public static async Task<decimal> GetBalanceAsync(long qq, IDbTransaction? trans = null)
        {
            return await GetAsync<decimal>("balance", qq, null, 0m, trans);
        }

        public static async Task<decimal> GetBalanceForUpdateAsync(long qq, IDbTransaction? trans = null)
        {
            return await GetForUpdateAsync<decimal>("balance", qq, null, 0m, trans);
        }

        public static async Task<decimal> GetFreezeBalanceAsync(long qq, IDbTransaction? trans = null)
        {
            return await GetAsync<decimal>("BalanceFreeze", qq, null, 0m, trans);
        }

        public static async Task<decimal> GetFreezeBalanceForUpdateAsync(long qq, IDbTransaction? trans = null)
        {
            return await GetForUpdateAsync<decimal>("BalanceFreeze", qq, null, 0m, trans);
        }

        public record AddBalanceResult(int Result, decimal BalanceValue);

        public static async Task<AddBalanceResult> AddBalanceAsync(long botUin, long groupId, string groupName, long qq, string name, decimal balanceAdd, string balanceInfo, IDbTransaction? trans = null)
        {
            using var wrapper = await BeginTransactionAsync(trans);
            try
            {
                // 1. 确保用户存在
                await AppendAsync(botUin, groupId, qq, name, await GroupInfo.GetGroupOwnerAsync(groupId), trans: wrapper.Transaction);
                
                // 2. 获取当前分值并加锁
                var balanceValue = await GetBalanceForUpdateAsync(qq, wrapper.Transaction);
                var newValue = balanceValue + balanceAdd;

                var (sql, paras) = await SqlAddBalanceAsync(qq, balanceAdd, wrapper.Transaction);
                var (sql2, paras2) = BalanceLog.SqlLog(botUin, groupId, groupName, qq, name, balanceAdd, balanceInfo, newValue);

                await ExecAsync(sql, wrapper.Transaction, paras);
                await ExecAsync(sql2, wrapper.Transaction, paras2);

                wrapper.Commit();

                SyncCacheField(qq, "Balance", newValue);
                return new AddBalanceResult(0, newValue);
            }
            catch (Exception ex)
            {
                wrapper.Rollback();
                Logger.Error($"[AddBalance Error] {ex.Message}");
                return new AddBalanceResult(-1, await GetBalanceAsync(qq));
            }
        }

        public static async Task<AddBalanceResult> MinusBalanceAsync(long botUin, long groupId, string groupName, long qq, string name, decimal balanceMinus, string balanceInfo, IDbTransaction? trans = null)
        {
            return await AddBalanceAsync(botUin, groupId, groupName, qq, name, -balanceMinus, balanceInfo, trans);
        }

        //转账 (异步事务版)
        public static async Task<(int Result, decimal SenderBalance, decimal ReceiverBalance)> TransferAsync(long botUin, long groupId, string groupName, long qq, string name, long qqTo, string nameTo, decimal balanceMinus, decimal balanceAdd)
        {
            using var wrapper = await BeginTransactionAsync();
            try
            {
                // 1. 获取发送者余额并加锁，检查是否足够
                var currentBalance = await GetBalanceForUpdateAsync(qq, wrapper.Transaction);
                if (currentBalance < balanceMinus)
                {
                    wrapper.Rollback();
                    return (-2, currentBalance, 0); // -2 表示余额不足
                }

                var res1 = await MinusBalanceAsync(botUin, groupId, groupName, qq, name, balanceMinus, $"转账给：{qqTo}", wrapper.Transaction);
                var res2 = await AddBalanceAsync(botUin, groupId, groupName, qqTo, nameTo, balanceAdd, $"转账来自：{qq}", wrapper.Transaction);

                wrapper.Commit();
                return (0, res1.BalanceValue, res2.BalanceValue);
            }
            catch (Exception ex)
            {
                wrapper.Rollback();
                Logger.Error($"[Transfer Error] {ex.Message}");
                return (-1, 0, 0);
            }
        }

        //转账操作
        public static string GetTransferBalance(long botUin, long groupId, string groupName, long qq, string name, string cmdPara) => GetTransferBalanceAsync(botUin, groupId, groupName, qq, name, cmdPara).GetAwaiter().GetResult();

        public static async Task<string> GetTransferBalanceAsync(long botUin, long groupId, string groupName, long qq, string name, string cmdPara)
        {
            cmdPara = cmdPara.Trim();
            if (cmdPara.NotMatch(Regexs.Transfer))
                return "格式：\n转账 + QQ + 余额\n例如：\n转账 {客服QQ} 9.99";
            long qqTransfer = cmdPara.RegexGetValue(Regexs.Transfer, "UserId").AsLong();
            decimal balanceTransfer = cmdPara.RegexGetValue(Regexs.Transfer, "balance").AsDecimal();
            if (qqTransfer == qq)
                return "不能转给自己";

            if (balanceTransfer < 1)
                return "至少转1.00R";

            var res = await TransferAsync(botUin, groupId, groupName, qq, name, qqTransfer, "", balanceTransfer, balanceTransfer);
            if (res.Result == -2)
                return $"余额{res.SenderBalance}不足{balanceTransfer}。";

            return res.Result == -1
                ? RetryMsg
                : $"✅ 成功转出：{balanceTransfer}\n[@:{qqTransfer}] 的余额：{res.ReceiverBalance}\n你的余额：{res.SenderBalance}";
        }

        //冻结余额 (异步事务版)
        public static async Task<(int Result, decimal BalanceValue, decimal FreezeValue)> FreezeBalanceAsync(long botUin, long groupId, string groupName, long qq, string name, decimal balanceFreeze)
        {
            using var wrapper = await BeginTransactionAsync();
            try
            {
                // 1. 获取当前余额并加锁
                decimal balanceValue = await GetBalanceForUpdateAsync(qq, wrapper.Transaction);
                if (balanceValue < balanceFreeze)
                {
                    wrapper.Rollback();
                    return (-1, balanceValue, 0);
                }

                decimal freezeValue = await GetFreezeBalanceForUpdateAsync(qq, wrapper.Transaction);
                decimal newBalance = balanceValue - balanceFreeze;
                decimal newFreeze = freezeValue + balanceFreeze;

                var (sql1, paras1) = BalanceLog.SqlLog(botUin, groupId, groupName, qq, name, -balanceFreeze, "冻结余额", newBalance);
                var (sql2, paras2) = SqlSetValues($"balance = balance - ({balanceFreeze}), BalanceFreeze = {SqlIsNull("BalanceFreeze", "0")} + ({balanceFreeze})", qq);
                await ExecAsync(sql1, wrapper.Transaction, paras1);
                await ExecAsync(sql2, wrapper.Transaction, paras2);

                wrapper.Commit();

                SyncCacheField(qq, "Balance", newBalance);
                SyncCacheField(qq, "BalanceFreeze", newFreeze);
                return (0, newBalance, newFreeze);
            }
            catch (Exception ex)
            {
                wrapper.Rollback();
                Logger.Error($"[FreezeBalance Error] {ex.Message}");
                return (-1, await GetBalanceAsync(qq), await GetFreezeBalanceAsync(qq));
            }
        }

        //解冻余额 (异步事务版)
        public static async Task<(int Result, decimal BalanceValue, decimal FreezeValue)> UnfreezeBalanceAsync(long botUin, long groupId, string groupName, long qq, string name, decimal balanceUnfreeze)
        {
            using var wrapper = await BeginTransactionAsync();
            try
            {
                // 1. 获取当前冻结余额并加锁
                decimal freezeValue = await GetFreezeBalanceForUpdateAsync(qq, wrapper.Transaction);
                if (freezeValue < balanceUnfreeze)
                {
                    wrapper.Rollback();
                    return (-1, 0, freezeValue);
                }

                decimal balanceValue = await GetBalanceForUpdateAsync(qq, wrapper.Transaction);
                decimal newBalance = balanceValue + balanceUnfreeze;
                decimal newFreeze = freezeValue - balanceUnfreeze;

                var (sql1, paras1) = BalanceLog.SqlLog(botUin, groupId, groupName, qq, name, balanceUnfreeze, "解冻余额", newBalance);
                var (sql2, paras2) = SqlSetValues($"balance = balance + ({balanceUnfreeze}), BalanceFreeze = {SqlIsNull("BalanceFreeze", "0")} - ({balanceUnfreeze})", qq);
                await ExecAsync(sql1, wrapper.Transaction, paras1);
                await ExecAsync(sql2, wrapper.Transaction, paras2);

                wrapper.Commit();

                SyncCacheField(qq, "Balance", newBalance);
                SyncCacheField(qq, "BalanceFreeze", newFreeze);
                return (0, newBalance, newFreeze);
            }
            catch (Exception ex)
            {
                wrapper.Rollback();
                Logger.Error($"[UnfreezeBalance Error] {ex.Message}");
                return (-1, await GetBalanceAsync(qq), await GetFreezeBalanceAsync(qq));
            }
        }

        //增加余额sql
        public static async Task<(string, IDataParameter[])> SqlAddBalanceAsync(long userId, decimal balancePlus, IDbTransaction? trans = null)
        {
            return await ExistsAsync(userId, trans)
                ? SqlPlus("Balance", balancePlus, userId)
                : SqlInsert(new { UserId = userId, Balance = balancePlus });
        }

        public static async Task<string> GetBalanceListAsync(long groupId, long qq)
        {
            string res = await QueryResAsync($"select {SqlTop(10)} [Id], [balance] from {FullName} where [UserId] in " +
                                  $"(select [UserId] from {CreditLog.FullName} where [GroupId] = {{0}}) order by [balance] desc {SqlLimit(10)}",
                                  "【第{i}名】 [@:{0}] 余额：{1:N}\n",
                                  groupId);
            return res.Contains(qq.ToString())
                ? res
                : $"{res}{await GetMyBalanceListAsync(groupId, qq)}";
        }

        public static async Task<string> GetMyBalanceListAsync(long groupId, long qq)
        {
            decimal balance = await GetBalanceAsync(qq);
            long res = await QueryScalarAsync<long>($"select count(*)+1 as res from {FullName} where [balance] > {{0}} and [UserId] in " +
                               $"(select [UserId] from {CreditLog.FullName} where [GroupId] = {{1}})",
                               balance, groupId);
            return $"【第{res}名】 [@:{qq}] 余额：{balance:N}";
        }
    }
}
