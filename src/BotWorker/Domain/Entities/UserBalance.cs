using Microsoft.Data.SqlClient;
using BotWorker.Domain.Entities;
using BotWorker.Common;
using BotWorker.Common.Extensions;
using BotWorker.Infrastructure.Persistence.ORM;

namespace BotWorker.Domain.Entities
{
    public partial class UserInfo : MetaDataGuid<UserInfo>
    {
        public static async Task<decimal> GetBalanceAsync(long qq)
        {
            return await GetAsync<decimal>("balance", qq, null, 0m);
        }

        public static decimal GetBalance(long qq) => GetBalanceAsync(qq).GetAwaiter().GetResult();

        public record AddBalanceResult(int Result, decimal BalanceValue);

        //增加余额
        public static int AddBalance(long botUin, long groupId, string groupName, long qq, string name, decimal balanceAdd, string balanceInfo)
        {
            return AddBalanceAsync(botUin, groupId, groupName, qq, name, balanceAdd, balanceInfo).GetAwaiter().GetResult().Result;
        }

        public static async Task<AddBalanceResult> AddBalanceAsync(long botUin, long groupId, string groupName, long qq, string name, decimal balanceAdd, string balanceInfo, SqlTransaction? trans = null)
        {
            // 如果没有传入事务，则创建一个新事务
            bool isNewTrans = false;
            if (trans == null)
            {
                trans = await BeginTransactionAsync();
                isNewTrans = true;
            }

            try
            {
                // 1. 确保用户存在
                await AppendAsync(botUin, groupId, qq, name, GroupInfo.GetGroupOwner(groupId));
                
                var balanceValue = GetBalance(qq) + balanceAdd;
                var (sql, paras) = SqlAddBalance(qq, balanceAdd);
                var (sql2, paras2) = BalanceLog.SqlLog(botUin, groupId, groupName, qq, name, balanceAdd, balanceInfo, balanceValue);

                await ExecAsync(sql, trans, paras);
                await ExecAsync(sql2, trans, paras2);

                if (isNewTrans)
                    await trans.CommitAsync();

                SyncCacheField(qq, "Balance", balanceValue);
                return new AddBalanceResult(0, balanceValue);
            }
            catch (Exception ex)
            {
                if (isNewTrans)
                    await trans.RollbackAsync();
                Console.WriteLine($"[AddBalance Error] {ex.Message}");
                return new AddBalanceResult(-1, GetBalance(qq));
            }
            finally
            {
                if (isNewTrans)
                {
                    trans.Connection?.Close();
                    await trans.DisposeAsync();
                }
            }
        }

        //减少余额
        public static int MinusBalance(long botUin, long groupId, string groupName, long qq, string name, decimal balanceMinus, string balanceInfo)
        {
            return MinusBalanceAsync(botUin, groupId, groupName, qq, name, balanceMinus, balanceInfo).GetAwaiter().GetResult().Result;
        }

        public static async Task<AddBalanceResult> MinusBalanceAsync(long botUin, long groupId, string groupName, long qq, string name, decimal balanceMinus, string balanceInfo, SqlTransaction? trans = null)
        {
            return await AddBalanceAsync(botUin, groupId, groupName, qq, name, -balanceMinus, balanceInfo, trans);
        }

        //转账 (异步事务版)
        public static async Task<(int Result, decimal SenderBalance, decimal ReceiverBalance)> TransferAsync(long botUin, long groupId, string groupName, long qq, string name, long qqTo, string nameTo, decimal balanceMinus, decimal balanceAdd)
        {
            using var trans = await BeginTransactionAsync();
            try
            {
                var res1 = await MinusBalanceAsync(botUin, groupId, groupName, qq, name, balanceMinus, $"转账给：{qqTo}", trans);
                var res2 = await AddBalanceAsync(botUin, groupId, groupName, qqTo, nameTo, balanceAdd, $"转账来自：{qq}", trans);

                await trans.CommitAsync();
                return (0, res1.BalanceValue, res2.BalanceValue);
            }
            catch (Exception ex)
            {
                await trans.RollbackAsync();
                Console.WriteLine($"[Transfer Error] {ex.Message}");
                return (-1, 0, 0);
            }
        }

        //转账
        public static int Transfer(long botUin, long groupId, string groupName, long qq, string name, long qqTo, string nameTo, decimal balanceMinus, decimal balanceAdd)
        {
            var res = TransferAsync(botUin, groupId, groupName, qq, name, qqTo, nameTo, balanceMinus, balanceAdd).GetAwaiter().GetResult();
            return res.Result;
        }

        //转账操作
        public static string GetTransferBalance(long botUin, long groupId, string groupName, long qq, string name, string cmdPara)
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

            decimal balanceMinus = balanceTransfer;
            if (GetBalance(qq) < balanceMinus)
                return $"余额{GetBalance(qq)}不足{balanceMinus}。";

            var res = TransferAsync(botUin, groupId, groupName, qq, name, qqTransfer, "", balanceMinus, balanceTransfer).GetAwaiter().GetResult();
            return res.Result == -1
                ? RetryMsg
                : $"✅ 成功转出：{balanceTransfer}\n[@:{qqTransfer}] 的余额：{res.ReceiverBalance}\n你的余额：{res.SenderBalance}";
        }

        //冻结余额
        public static decimal GetFreezeBalance(long qq)
        {
            return Get<decimal>("BalanceFreeze", qq);
        }

        public static async Task<decimal> GetFreezeBalanceAsync(long qq)
        {
            return await GetAsync<decimal>("BalanceFreeze", qq, null, 0m);
        }

        //冻结余额 (异步事务版)
        public static async Task<(int Result, decimal BalanceValue, decimal FreezeValue)> FreezeBalanceAsync(long botUin, long groupId, string groupName, long qq, string name, decimal balanceFreeze)
        {
            decimal balanceValue = GetBalance(qq);
            if (balanceValue < balanceFreeze)
                return (-1, balanceValue, 0);

            decimal freezeValue = GetFreezeBalance(qq);

            using var trans = await BeginTransactionAsync();
            try
            {
                var (sql1, paras1) = BalanceLog.SqlLog(botUin, groupId, groupName, qq, name, -balanceFreeze, "冻结余额", balanceValue - balanceFreeze);
                var (sql2, paras2) = SqlSetValues($"balance = balance - ({balanceFreeze}), BalanceFreeze = isnull(BalanceFreeze,0) + ({balanceFreeze})", qq);
                await ExecAsync(sql1, trans, paras1);
                await ExecAsync(sql2, trans, paras2);

                await trans.CommitAsync();

                SyncCacheField(qq, "Balance", balanceValue - balanceFreeze);
                SyncCacheField(qq, "BalanceFreeze", freezeValue + balanceFreeze);
                return (0, balanceValue - balanceFreeze, freezeValue + balanceFreeze);
            }
            catch (Exception ex)
            {
                await trans.RollbackAsync();
                Console.WriteLine($"[FreezeBalance Error] {ex.Message}");
                return (-1, balanceValue, freezeValue);
            }
        }

        //冻结余额
        public static int FreezeBalance(long botUin, long groupId, string groupName, long qq, string name, decimal balanceFreeze)
        {
            return FreezeBalanceAsync(botUin, groupId, groupName, qq, name, balanceFreeze).GetAwaiter().GetResult().Result;
        }

        //解冻余额 (异步事务版)
        public static async Task<(int Result, decimal BalanceValue, decimal FreezeValue)> UnfreezeBalanceAsync(long botUin, long groupId, string groupName, long qq, string name, decimal balanceUnfreeze)
        {
            decimal freezeValue = GetFreezeBalance(qq);
            if (freezeValue < balanceUnfreeze) return (-1, 0, freezeValue);

            decimal balanceValue = GetBalance(qq);

            using var trans = await BeginTransactionAsync();
            try
            {
                var (sql1, paras1) = BalanceLog.SqlLog(botUin, groupId, groupName, qq, name, balanceUnfreeze, "解冻余额", balanceValue + balanceUnfreeze);
                var (sql2, paras2) = SqlSetValues($"balance = balance + ({balanceUnfreeze}), BalanceFreeze = isnull(BalanceFreeze, 0) - ({balanceUnfreeze})", qq);
                await ExecAsync(sql1, trans, paras1);
                await ExecAsync(sql2, trans, paras2);

                await trans.CommitAsync();

                SyncCacheField(qq, "Balance", balanceValue + balanceUnfreeze);
                SyncCacheField(qq, "BalanceFreeze", freezeValue - balanceUnfreeze);
                return (0, balanceValue + balanceUnfreeze, freezeValue - balanceUnfreeze);
            }
            catch (Exception ex)
            {
                await trans.RollbackAsync();
                Console.WriteLine($"[UnfreezeBalance Error] {ex.Message}");
                return (-1, balanceValue, freezeValue);
            }
        }

        //解冻余额
        public static int UnfreezeBalance(long botUin, long groupId, string groupName, long qq, string name, decimal balanceUnfreeze)
        {
            return UnfreezeBalanceAsync(botUin, groupId, groupName, qq, name, balanceUnfreeze).GetAwaiter().GetResult().Result;
        }

        //增加余额sql
        public static (string, SqlParameter[]) SqlAddBalance(long userId, decimal balancePlus)
        {
            return Exists(userId)
                ? SqlPlus("Balance", balancePlus, userId)
                : SqlInsert([
                                new Cov("UserId", userId),
                                new Cov("Balance", balancePlus),
                            ]);
        }

        public static async Task<string> GetBalanceListAsync(long groupId, long qq)
        {
            string res = await QueryResAsync($"select top 10 Id, balance from {FullName} where UserId in " +
                                  $"(select UserId from {CreditLog.FullName} where GroupId = {groupId}) order by balance desc",
                                  "【第{i}名】 [@:{0}] 余额：{1:N}\n");
            return res.Contains(qq.ToString())
                ? res
                : $"{res}{await GetMyBalanceListAsync(groupId, qq)}";
        }

        public static string GetBalanceList(long groupId, long qq) => GetBalanceListAsync(groupId, qq).GetAwaiter().GetResult();

        public static async Task<string> GetMyBalanceListAsync(long groupId, long qq)
        {
            decimal balance = await GetBalanceAsync(qq);
            string res = await QueryAsync($"select count(*)+1 as res from {FullName} where balance > {balance} and UserId in " +
                               $"(select UserId from {CreditLog.FullName} where GroupId = {groupId})");
            return $"【第{res}名】 [@:{qq}] 余额：{balance:N}";
        }

        public static string GetMyBalanceList(long groupId, long qq) => GetMyBalanceListAsync(groupId, qq).GetAwaiter().GetResult();
    }
}
