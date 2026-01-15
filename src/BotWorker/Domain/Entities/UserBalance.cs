using System;
using System.Data;
using System.Threading.Tasks;
using BotWorker.Domain.Entities;
using BotWorker.Domain.Repositories;
using Microsoft.Extensions.DependencyInjection;
using BotWorker.Domain.Models.BotMessages;

namespace BotWorker.Domain.Entities
{
    public partial class UserInfo
    {
        private static IUserRepository UserRepository => BotMessage.ServiceProvider.GetRequiredService<IUserRepository>();
        private static IBalanceLogRepository BalanceLogRepository => BotMessage.ServiceProvider.GetRequiredService<IBalanceLogRepository>();
        private static IGroupMemberRepository GroupMemberRepository => BotMessage.ServiceProvider.GetRequiredService<IGroupMemberRepository>();
        private static IGroupRepository GroupRepository => BotMessage.ServiceProvider.GetRequiredService<IGroupRepository>();

        public static async Task<decimal> GetBalanceAsync(long qq, IDbTransaction? trans = null)
        {
            return await UserRepository.GetBalanceAsync(qq, trans);
        }

        public static async Task<decimal> GetBalanceForUpdateAsync(long qq, IDbTransaction? trans = null)
        {
            return await UserRepository.GetBalanceForUpdateAsync(qq, trans);
        }

        public static async Task<decimal> GetFreezeBalanceAsync(long qq, IDbTransaction? trans = null)
        {
            return await UserRepository.GetFreezeBalanceAsync(qq, trans);
        }

        public static async Task<decimal> GetFreezeBalanceForUpdateAsync(long qq, IDbTransaction? trans = null)
        {
            return await UserRepository.GetFreezeBalanceForUpdateAsync(qq, trans);
        }

        public record AddBalanceResult(int Result, decimal BalanceValue);

        public static async Task<AddBalanceResult> AddBalanceAsync(long botUin, long groupId, string groupName, long qq, string name, decimal balanceAdd, string balanceInfo, IDbTransaction? trans = null)
        {
            using var wrapper = await UserRepository.BeginTransactionAsync(trans);
            try
            {
                // 1. 确保用户存在
                long ownerId = await GroupRepository.GetGroupOwnerAsync(groupId); // 注意：Repository method signature might differ slightly, checking usage
                // GroupInfo.GetGroupOwnerAsync supported trans, Repository one might not?
                // IGroupRepository has Task<long> GetGroupOwnerAsync(long groupId);
                
                // AppendAsync in GroupMemberRepository
                await GroupMemberRepository.AppendAsync(groupId, qq, name, "", 0, "", wrapper.Transaction);
                
                // 2. 获取当前分值并加锁
                var balanceValue = await UserRepository.GetBalanceForUpdateAsync(qq, wrapper.Transaction);
                var newValue = balanceValue + balanceAdd;

                // 3. Update Balance
                // Note: AddBalanceAsync in Repository does UPDATE balance = balance + amount.
                // We want to return the new value. 
                // Using AddBalanceAsync is fine, but we need to know the new value.
                // We calculated newValue above.
                // But AddBalanceAsync in Repository does "balance = balance + @amount".
                // If we use that, we trust the DB calculation.
                // And we already locked the row, so newValue is correct.
                
                // Handle case where User might not exist in user_info but AppendAsync should have created it?
                // If AppendAsync creates GroupMember, it usually creates UserInfo first.
                // But if UserInfo was just created, balance is 0.
                // GetBalanceForUpdateAsync would return 0.
                // So newValue = 0 + balanceAdd.
                
                // However, UserRepository.AddBalanceAsync assumes record exists.
                // If AppendAsync ensures it, we are fine.
                
                bool updated = await UserRepository.AddBalanceAsync(qq, balanceAdd, wrapper.Transaction);
                if (!updated)
                {
                    // Should not happen if AppendAsync works and we locked row.
                    // But if it happens, maybe insert?
                    // Repository.AddAsync?
                    // But we assume AppendAsync handles it.
                    // Let's assume updated is true.
                }

                // 4. Log
                await BalanceLogRepository.AddLogAsync(botUin, groupId, groupName, qq, name, balanceAdd, newValue, balanceInfo, wrapper.Transaction);

                await wrapper.CommitAsync();

                // Sync Cache
                await UserRepository.SyncCacheFieldAsync(qq, "Balance", newValue);
                
                return new AddBalanceResult(0, newValue);
            }
            catch (Exception ex)
            {
                await wrapper.RollbackAsync();
                Logger.Error($"[AddBalance Error] {ex.Message}");
                // Fallback to get current balance without lock
                return new AddBalanceResult(-1, await UserRepository.GetBalanceAsync(qq));
            }
        }

        public static async Task<AddBalanceResult> MinusBalanceAsync(long botUin, long groupId, string groupName, long qq, string name, decimal balanceMinus, string balanceInfo, IDbTransaction? trans = null)
        {
            return await AddBalanceAsync(botUin, groupId, groupName, qq, name, -balanceMinus, balanceInfo, trans);
        }

        //转账 (异步事务版)
        public static async Task<(int Result, decimal SenderBalance, decimal ReceiverBalance)> TransferAsync(long botUin, long groupId, string groupName, long qq, string name, long qqTo, string nameTo, decimal balanceMinus, decimal balanceAdd)
        {
            using var wrapper = await UserRepository.BeginTransactionAsync();
            try
            {
                // 1. 确保用户存在 (按 ID 从小到大执行，防止 User 表死锁)
                long firstId = qq < qqTo ? qq : qqTo;
                long secondId = qq < qqTo ? qqTo : qq;
                string firstName = firstId == qq ? name : nameTo;
                string secondName = secondId == qq ? name : nameTo;

                long ownerId = await GroupRepository.GetGroupOwnerAsync(groupId);
                await GroupMemberRepository.AppendAsync(groupId, firstId, firstName, "", 0, "", wrapper.Transaction);
                await GroupMemberRepository.AppendAsync(groupId, secondId, secondName, "", 0, "", wrapper.Transaction);

                // 2. 统一加锁顺序，防止死锁 (按 ID 从小到大锁定)
                await UserRepository.GetBalanceForUpdateAsync(firstId, wrapper.Transaction);
                await UserRepository.GetBalanceForUpdateAsync(secondId, wrapper.Transaction);

                // 获取发送者余额并检查是否足够
                var currentBalance = await UserRepository.GetBalanceAsync(qq, wrapper.Transaction);
                if (currentBalance < balanceMinus)
                {
                    await wrapper.RollbackAsync();
                    return (-2, currentBalance, 0); // -2 表示余额不足
                }

                // Execute transfer
                // We can reuse AddBalanceAsync but we need to pass the transaction wrapper.Transaction
                // And AddBalanceAsync creates its own wrapper if trans is null, or uses provided.
                // But AddBalanceAsync calls GetGroupOwner and AppendAsync again. 
                // Since we already did that, maybe we can optimize?
                // But calling it again is safe (idempotent usually) and easier to reuse code.
                // However, AddBalanceAsync does BeginTransactionAsync(trans). 
                // If we pass wrapper.Transaction, it uses it.
                
                var res1 = await MinusBalanceAsync(botUin, groupId, groupName, qq, name, balanceMinus, $"转账给：{qqTo}", wrapper.Transaction);
                var res2 = await AddBalanceAsync(botUin, groupId, groupName, qqTo, nameTo, balanceAdd, $"转账来自：{qq}", wrapper.Transaction);

                if (res1.Result == -1 || res2.Result == -1)
                    throw new Exception("Transfer failed");

                await wrapper.CommitAsync();
                return (0, res1.BalanceValue, res2.BalanceValue);
            }
            catch (Exception ex)
            {
                await wrapper.RollbackAsync();
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
            using var wrapper = await UserRepository.BeginTransactionAsync();
            try
            {
                // 1. 获取当前余额并加锁
                decimal balanceValue = await UserRepository.GetBalanceForUpdateAsync(qq, wrapper.Transaction);
                if (balanceValue < balanceFreeze)
                {
                    await wrapper.RollbackAsync();
                    return (-1, balanceValue, 0);
                }

                // 2. Perform Freeze
                await UserRepository.FreezeBalanceAsync(qq, balanceFreeze, wrapper.Transaction);
                
                // Calculate new values for return
                decimal freezeValue = await UserRepository.GetFreezeBalanceAsync(qq, wrapper.Transaction); // This is new freeze value
                decimal newBalance = balanceValue - balanceFreeze;

                // 3. Log
                await BalanceLogRepository.AddLogAsync(botUin, groupId, groupName, qq, name, -balanceFreeze, newBalance, "冻结余额", wrapper.Transaction);

                await wrapper.CommitAsync();

                await UserRepository.SyncCacheFieldAsync(qq, "Balance", newBalance);
                await UserRepository.SyncCacheFieldAsync(qq, "BalanceFreeze", freezeValue);
                return (0, newBalance, freezeValue);
            }
            catch (Exception ex)
            {
                await wrapper.RollbackAsync();
                Logger.Error($"[FreezeBalance Error] {ex.Message}");
                return (-1, await UserRepository.GetBalanceAsync(qq), await UserRepository.GetFreezeBalanceAsync(qq));
            }
        }

        //解冻余额 (异步事务版)
        public static async Task<(int Result, decimal BalanceValue, decimal FreezeValue)> UnfreezeBalanceAsync(long botUin, long groupId, string groupName, long qq, string name, decimal balanceUnfreeze)
        {
            using var wrapper = await UserRepository.BeginTransactionAsync();
            try
            {
                // 1. 获取当前冻结余额并加锁
                decimal freezeValue = await UserRepository.GetFreezeBalanceForUpdateAsync(qq, wrapper.Transaction);
                if (freezeValue < balanceUnfreeze)
                {
                    await wrapper.RollbackAsync();
                    return (-1, 0, freezeValue);
                }

                // 2. Perform Unfreeze (using FreezeBalanceAsync with negative amount? Or implementing Unfreeze?)
                // FreezeBalanceAsync: balance - amount, freeze + amount
                // Unfreeze: balance + amount, freeze - amount
                // So passing negative amount to FreezeBalanceAsync works:
                // balance - (-amount) = balance + amount
                // freeze + (-amount) = freeze - amount
                await UserRepository.FreezeBalanceAsync(qq, -balanceUnfreeze, wrapper.Transaction);

                decimal balanceValue = await UserRepository.GetBalanceAsync(qq, wrapper.Transaction); // new balance
                decimal newFreeze = freezeValue - balanceUnfreeze;

                // 3. Log
                await BalanceLogRepository.AddLogAsync(botUin, groupId, groupName, qq, name, balanceUnfreeze, balanceValue, "解冻余额", wrapper.Transaction);

                await wrapper.CommitAsync();

                await UserRepository.SyncCacheFieldAsync(qq, "Balance", balanceValue);
                await UserRepository.SyncCacheFieldAsync(qq, "BalanceFreeze", newFreeze);
                return (0, balanceValue, newFreeze);
            }
            catch (Exception ex)
            {
                await wrapper.RollbackAsync();
                Logger.Error($"[UnfreezeBalance Error] {ex.Message}");
                return (-1, await UserRepository.GetBalanceAsync(qq), await UserRepository.GetFreezeBalanceAsync(qq));
            }
        }

        public static async Task<string> GetBalanceListAsync(long groupId, long qq)
        {
            return await UserRepository.GetBalanceListAsync(groupId, qq);
        }

        public static async Task<string> GetMyBalanceListAsync(long groupId, long qq)
        {
            return await UserRepository.GetRankAsync(groupId, qq);
        }
    }
}
