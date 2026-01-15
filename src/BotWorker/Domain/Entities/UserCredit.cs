using System;
using System.Data;
using System.Threading.Tasks;
using BotWorker.Domain.Repositories;
using Microsoft.Extensions.DependencyInjection;

namespace BotWorker.Domain.Entities
{
    public partial class UserInfo
    {
        private static IGroupRepository GroupRepository => 
            BotMessage.ServiceProvider?.GetRequiredService<IGroupRepository>() 
            ?? throw new InvalidOperationException("IGroupRepository not registered");

        private static IBotRepository BotRepository => 
            BotMessage.ServiceProvider?.GetRequiredService<IBotRepository>() 
            ?? throw new InvalidOperationException("IBotRepository not registered");

        private static IGroupMemberRepository GroupMemberRepository => 
            BotMessage.ServiceProvider?.GetRequiredService<IGroupMemberRepository>() 
            ?? throw new InvalidOperationException("IGroupMemberRepository not registered");

        public static async Task<(int Result, long CreditValue, int LogId)> AddCreditAsync(long botUin, long groupId, string groupName, long qq, string name, long creditAdd, string creditInfo, IDbTransaction? trans = null)
        {
            try
            {
                // 1. 确保用户存在
                long ownerId = await GroupRepository.GetGroupOwnerAsync(groupId, 0, trans);
                await GroupMemberRepository.AppendAsync(groupId, qq, name, "", 0, "", trans);

                // 2. 获取当前准确分值（在事务内获取，并加锁防止并发修改）
                var creditValue = await Repository.GetCreditForUpdateAsync(botUin, groupId, qq, trans);

                // 3. 执行积分操作
                await Repository.AddCreditAsync(botUin, groupId, qq, creditAdd, trans);

                // 4. 记录日志
                int logId = await CreditLog.AddLogAsync(botUin, groupId, groupName, qq, name, creditAdd, creditValue, creditInfo, trans);

                long newValue = creditValue + creditAdd;
                return (0, newValue, logId);
            }
            catch (Exception ex)
            {
                Logger.Error($"[AddCredit Error] {ex.Message}");
                if (trans != null) throw; 
                return (-1, 0, 0);
            }
        }

        public static async Task<(int Result, long CreditValue, int LogId)> AddCreditTransAsync(long botUin, long groupId, string groupName, long qq, string name, long creditAdd, string creditInfo, IDbTransaction? trans = null)
        {
            using var wrapper = await Repository.BeginTransactionAsync(trans);
            try
            {
                var res = await AddCreditAsync(botUin, groupId, groupName, qq, name, creditAdd, creditInfo, wrapper.Transaction);
                await wrapper.CommitAsync();

                // 5. 统一同步缓存（仅在自身开启事务时同步）
                if (trans == null)
                {
                    await SyncCreditCacheAsync(botUin, groupId, qq, res.CreditValue);
                }

                return res;
            }
            catch (Exception ex)
            {
                Logger.Error($"[AddCreditTrans Error] {ex.Message}");
                await wrapper.RollbackAsync();
                if (trans != null) throw;
                return (-1, 0, 0);
            }
        }

        public static async Task<int> MinusCreditAsync(long botUin, long groupId, string groupName, long qq, string name, long creditMinus, string creditInfo)
        {
            var res = await AddCreditTransAsync(botUin, groupId, groupName, qq, name, -creditMinus, creditInfo);
            return res.Result;
        }

        public static async Task<int> AppendUserAsync(long botUin, long groupId, long qq, string name, long ownerId, IDbTransaction? trans = null)
        {
            return await Repository.AppendAsync(botUin, groupId, qq, name, ownerId, trans);
        }

        public static async Task<int> SetStateAsync(object state, long qq)
        {
            int stateValue = state is int i ? i : (int)state;
            return await Repository.SetValueAsync("state", stateValue, qq);
        }

        public static async Task SyncCreditCacheAsync(long botUin, long groupId, long qq, long newValue)
        {
            if (await GroupRepository.GetIsCreditAsync(groupId))
                await GroupMember.SyncCacheFieldAsync(groupId, qq, "group_credit", newValue);
            else if (await BotRepository.GetIsCreditAsync(botUin))
                await Friend.SyncCacheFieldAsync(botUin, qq, "credit", newValue);
            else
                await Repository.SyncCacheFieldAsync(qq, "credit", newValue);
        }

        public static async Task SyncSaveCreditCacheAsync(long botUin, long groupId, long qq, long newValue)
        {
            if (await GroupRepository.GetIsCreditAsync(groupId))
                await GroupMember.SyncCacheFieldAsync(groupId, qq, "save_credit", newValue);
            else if (await BotRepository.GetIsCreditAsync(botUin))
                await Friend.SyncCacheFieldAsync(botUin, qq, "save_credit", newValue);
            else
                await Repository.SyncCacheFieldAsync(qq, "save_credit", newValue);
        }

        public static async Task<long> GetCreditAsync(long botUin, long groupId, long qq, IDbTransaction? trans = null)
        {
            return await Repository.GetCreditAsync(botUin, groupId, qq, trans);
        }

        public static async Task<long> GetCreditForUpdateAsync(long botUin, long groupId, long qq, IDbTransaction? trans = null)
        {
            return await Repository.GetCreditForUpdateAsync(botUin, groupId, qq, trans);
        }

        //转账积分 (重构版本：代码极简)
        public static async Task<(int Result, long SenderCredit, long ReceiverCredit)> TransferCreditAsync(
            long botUin, long groupId, string groupName, 
            long senderId, string senderName, 
            long receiverId, string receiverName, 
            long creditMinus, long creditAdd, 
            string transferInfo, IDbTransaction? trans = null)
        {
            // 1. 开启事务 (在事务内检查余额并加锁)
            using var wrapper = await Repository.BeginTransactionAsync(trans);
            try
            {
                // 2. 确保用户存在 (按 ID 从小到大执行，防止 User 表死锁)
                long firstId = senderId < receiverId ? senderId : receiverId;
                long secondId = senderId < receiverId ? receiverId : senderId;
                string firstName = firstId == senderId ? senderName : receiverName;
                string secondName = secondId == senderId ? senderName : receiverName;

                long ownerId = await GroupRepository.GetGroupOwnerAsync(groupId, 0, wrapper.Transaction);
                await Repository.AppendAsync(botUin, groupId, firstId, firstName, ownerId, wrapper.Transaction);
                await Repository.AppendAsync(botUin, groupId, secondId, secondName, ownerId, wrapper.Transaction);

                // 3. 统一加锁顺序，防止 GroupMember 表死锁 (按 ID 从小到大锁定)
                if (firstId == senderId)
                {
                    await Repository.GetCreditForUpdateAsync(botUin, groupId, senderId, wrapper.Transaction);
                    await Repository.GetCreditForUpdateAsync(botUin, groupId, receiverId, wrapper.Transaction);
                }
                else
                {
                    await Repository.GetCreditForUpdateAsync(botUin, groupId, receiverId, wrapper.Transaction);
                    await Repository.GetCreditForUpdateAsync(botUin, groupId, senderId, wrapper.Transaction);
                }

                // 获取发送者当前分值
                long senderCredit = await Repository.GetCreditAsync(botUin, groupId, senderId, wrapper.Transaction);
                if (senderCredit < creditMinus)
                    return (-1, senderCredit, 0);

                // 4. 链式调用业务方法，全部复用同一个 trans
                // 扣除发送者积分 (内部会自动记录日志)
                var res1 = await AddCreditAsync(botUin, groupId, groupName, senderId, senderName, -creditMinus, $"{transferInfo}扣分：{receiverId}", wrapper.Transaction);
                
                // 增加接收者积分 (内部会自动记录日志)
                var res2 = await AddCreditAsync(botUin, groupId, groupName, receiverId, receiverName, creditAdd, $"{transferInfo}加分：{senderId}", wrapper.Transaction);

                // 5. 提交事务
                await wrapper.CommitAsync();

                // 6. 同步缓存 (仅在自身开启事务时同步)
                if (trans == null)
                {
                    await SyncCreditCacheAsync(botUin, groupId, senderId, res1.CreditValue);
                    await SyncCreditCacheAsync(botUin, groupId, receiverId, res2.CreditValue);
                }

                return (0, res1.CreditValue, res2.CreditValue);
            }
            catch (Exception ex)
            {
                await wrapper.RollbackAsync();
                Logger.Error($"[TransferCredit Error] {ex.Message}");
                if (trans != null) throw;
                return (-1, 0, 0);
            }
        }
    }
}
