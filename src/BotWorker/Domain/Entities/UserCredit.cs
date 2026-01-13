using System.Data;

namespace BotWorker.Domain.Entities
{
    public partial class UserInfo : MetaDataGuid<UserInfo>
    {
        public static async Task<(int Result, long CreditValue, int LogId)> AddCreditAsync(long botUin, long groupId, string groupName, long qq, string name, long creditAdd, string creditInfo, IDbTransaction? trans = null)
        {
            // Logger.Debug($"[AddCredit Start] QQ:{qq}, Add:{creditAdd}, Info:{creditInfo}, ExistingTrans:{(trans != null)}");
            try
            {
                // 1. 确保用户存在
                long ownerId = await GroupInfo.GetGroupOwnerAsync(groupId, 0, trans);
                await AppendAsync(botUin, groupId, qq, name, ownerId, trans: trans);

                // 2. 获取当前准确分值（在事务内获取，并加锁防止并发修改）
                var creditValue = await GetCreditForUpdateAsync(botUin, groupId, qq, trans);
                // Logger.Debug($"[AddCredit Current] QQ:{qq}, Current:{creditValue}");

                // 3. 执行积分操作
                var (sql, paras) = await SqlAddCreditAsync(botUin, groupId, qq, creditAdd, trans);
                // Logger.Debug($"[AddCredit SQL] {sql}");
                await ExecAsync(sql, trans, paras);

                // 4. 记录日志
                int logId = await CreditLog.AddLogAsync(botUin, groupId, groupName, qq, name, creditAdd, creditValue, creditInfo, trans);

                long newValue = creditValue + creditAdd;
                // Logger.Debug($"[AddCredit Success] QQ:{qq}, Add:{creditAdd}, NewValue:{newValue}");
                return (0, newValue, logId);
            }
            catch (Exception ex)
            {
                Logger.Error($"[AddCredit Error] {ex.Message}");
                if (trans != null) throw; // 事务嵌套时抛出异常，由外层事务处理回滚
                return (-1, 0, 0);
            }
        }

        public static async Task<(int Result, long CreditValue, int LogId)> AddCreditTransAsync(long botUin, long groupId, string groupName, long qq, string name, long creditAdd, string creditInfo, IDbTransaction? trans = null)
        {
            using var wrapper = await BeginTransactionAsync(trans);
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
            return await AppendAsync(botUin, groupId, qq, name, ownerId, trans: trans);
        }

        public static int SetState(object state, long qq)
        {
            int stateValue = state is int i ? i : (int)state;
            return SetValue("State", stateValue, qq);
        }

        public static async Task SyncCreditCacheAsync(long botUin, long groupId, long qq, long newValue)
        {
            if (await GroupInfo.GetIsCreditAsync(groupId))
                GroupMember.SyncCacheField(groupId, qq, "GroupCredit", newValue);
            else if (await BotInfo.GetIsCreditAsync(botUin))
                Friend.SyncCacheField(botUin, qq, "Credit", newValue);
            else
                SyncCacheField(qq, "Credit", newValue);
        }

        public static async Task SyncSaveCreditCacheAsync(long botUin, long groupId, long qq, long newValue)
        {
            if (await GroupInfo.GetIsCreditAsync(groupId))
                GroupMember.SyncCacheField(groupId, qq, "SaveCredit", newValue);
            else if (await BotInfo.GetIsCreditAsync(botUin))
                Friend.SyncCacheField(botUin, qq, "SaveCredit", newValue);
            else
                SyncCacheField(qq, "SaveCredit", newValue);
        }

        public static async Task<(string, IDataParameter[])> SqlAddCreditAsync(long botUin, long groupId, long userId, long creditPlus, IDbTransaction? trans = null)
        {
            if (await GroupInfo.GetIsCreditAsync(groupId, trans))
            {
                return await GroupMember.SqlAddCreditAsync(groupId, userId, creditPlus, trans);
            }
            else if (await BotInfo.GetIsCreditAsync(botUin, trans))
            {
                return await Friend.SqlAddCreditAsync(botUin, userId, creditPlus, trans);
            }
            else
            {
                // 暂时保持同步 Exists，因为它调用频繁且是内存/主键检查
                if (await ExistsAsync(userId, null, trans))
                    return SqlPlus("Credit", creditPlus, userId);
                else
                    return SqlInsert(new
                    {
                        BotUin = botUin,
                        GroupId = groupId,
                        Id = userId,
                        Credit = creditPlus,
                    });
            }
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
            using var wrapper = await BeginTransactionAsync(trans);
            try
            {
                // 2. 确保用户存在 (按 ID 从小到大执行，防止 User 表死锁)
                long firstId = senderId < receiverId ? senderId : receiverId;
                long secondId = senderId < receiverId ? receiverId : senderId;
                string firstName = firstId == senderId ? senderName : receiverName;
                string secondName = secondId == senderId ? senderName : receiverName;

                long ownerId = await GroupInfo.GetGroupOwnerAsync(groupId, 0, wrapper.Transaction);
                await AppendAsync(botUin, groupId, firstId, firstName, ownerId, trans: wrapper.Transaction);
                await AppendAsync(botUin, groupId, secondId, secondName, ownerId, trans: wrapper.Transaction);

                // 3. 统一加锁顺序，防止 GroupMember 表死锁 (按 ID 从小到大锁定)
                if (firstId == senderId)
                {
                    await GetCreditForUpdateAsync(botUin, groupId, senderId, wrapper.Transaction);
                    await GetCreditForUpdateAsync(botUin, groupId, receiverId, wrapper.Transaction);
                }
                else
                {
                    await GetCreditForUpdateAsync(botUin, groupId, receiverId, wrapper.Transaction);
                    await GetCreditForUpdateAsync(botUin, groupId, senderId, wrapper.Transaction);
                }

                // 获取发送者当前分值
                long senderCredit = await GetCreditAsync(botUin, groupId, senderId, wrapper.Transaction);
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

        public static async Task<long> GetCreditAsync(long botUin, long groupId, long qq, IDbTransaction? trans = null)
        {
            if (groupId != 0 && await GroupInfo.GetIsCreditAsync(groupId, trans))
            {
                // Logger.Debug($"[UserInfo] GetCredit - Type: Group, GroupId: {groupId}, QQ: {qq}");
                return await GroupMember.GetGroupCreditAsync(groupId, qq, trans);
            }
            else if (await BotInfo.GetIsCreditAsync(botUin, trans))
            {
                // Logger.Debug($"[UserInfo] GetCredit - Type: Friend, BotUin: {botUin}, QQ: {qq}");
                return await Friend.GetCreditAsync(botUin, qq, trans);
            }
            else
            {
                // Logger.Debug($"[UserInfo] GetCredit - Type: General, QQ: {qq}");
                return await GetCreditAsync(botUin, qq, trans);
            }
        }

        public static async Task<long> GetCreditForUpdateAsync(long botUin, long groupId, long qq, IDbTransaction? trans = null)
        {
            if (groupId != 0 && await GroupInfo.GetIsCreditAsync(groupId, trans))
            {
                // Logger.Debug($"[UserInfo] GetCreditForUpdate - Type: Group, GroupId: {groupId}, QQ: {qq}");
                return await GroupMember.GetGroupCreditForUpdateAsync(groupId, qq, trans);
            }
            else if (await BotInfo.GetIsCreditAsync(botUin, trans))
            {
                // Logger.Debug($"[UserInfo] GetCreditForUpdate - Type: Friend, BotUin: {botUin}, QQ: {qq}");
                return await Friend.GetCreditForUpdateAsync(botUin, qq, trans);
            }
            else
            {
                // Logger.Debug($"[UserInfo] GetCreditForUpdate - Type: General, QQ: {qq}");
                return await GetForUpdateAsync<long>("Credit", qq, null, 0, trans);
            }
        }

        public static async Task<long> GetCreditAsync(long userId, IDbTransaction? trans = null)
        {
            return await GetLongAsync("Credit", userId, null, trans);
        }

        //读取积分
        public static async Task<long> GetCreditAsync(long botUin, long userId, IDbTransaction? trans = null)
        {
            return await BotInfo.GetIsCreditAsync(botUin) ? await Friend.GetCreditAsync(botUin, userId, trans) : await GetLongAsync("Credit", userId, null, trans);
        }

        //积分总额
        public static async Task<long> GetTotalCreditAsync(long botUin, long userId) => await GetCreditAsync(botUin, userId) + await GetSaveCreditAsync(botUin, userId);
        public static async Task<long> GetTotalCreditAsync(long botUin, long groupId, long userId) => await GetCreditAsync(botUin, groupId, userId) + await GetSaveCreditAsync(botUin, groupId, userId);

        public static async Task<long> GetSaveCreditAsync(long botUin, long groupId, long qq, IDbTransaction? trans = null)
        {
            return await GroupInfo.GetIsCreditAsync(groupId)
                ? await GroupMember.GetLongAsync("SaveCredit", groupId, qq, trans)
                : await GetSaveCreditAsync(botUin, qq, trans);
        }

        public static async Task<long> GetSaveCreditAsync(long botUin, long userId, IDbTransaction? trans = null)
        {
            return await BotInfo.GetIsCreditAsync(botUin)
                ? await Friend.GetSaveCreditAsync(botUin, userId, trans)
                : await GetSaveCreditAsync(userId, trans);
        }

        public static async Task<long> GetSaveCreditAsync(long userId, IDbTransaction? trans = null)
        {
            return await GetLongAsync("SaveCredit", userId, null, trans);
        }

        public static async Task<long> GetSaveCreditForUpdateAsync(long botUin, long groupId, long qq, IDbTransaction? trans = null)
        {
            if (await GroupInfo.GetIsCreditAsync(groupId, trans))
            {
                return await GroupMember.GetSaveCreditForUpdateAsync(groupId, qq, trans);
            }
            else if (await BotInfo.GetIsCreditAsync(botUin, trans))
            {
                return await Friend.GetForUpdateAsync<long>("SaveCredit", botUin, qq, 0, trans);
            }
            else
            {
                return await GetForUpdateAsync<long>("SaveCredit", qq, null, 0, trans);
            }
        }

        public static async Task<(string, IDataParameter[])> SqlSaveCreditAsync(long botUin, long groupId, long userId, long creditSave, IDbTransaction? trans = null)
        {
            return await GroupInfo.GetIsCreditAsync(groupId, trans)
                ? GroupMember.SqlSaveCredit(groupId, userId, creditSave)
                : await BotInfo.GetIsCreditAsync(botUin, trans) ? Friend.SqlSaveCredit(botUin, userId, creditSave)
                                 : SqlSetValues($"Credit = Credit - ({creditSave}), SaveCredit = {SqlIsNull("SaveCredit", "0")} + ({creditSave})", userId);
        }

        public static (string, IDataParameter[]) SqlSaveCredit(long botUin, long groupId, long userId, long creditSave)
            => SqlSaveCreditAsync(botUin, groupId, userId, creditSave).GetAwaiter().GetResult();

        public static (string, IDataParameter[]) SqlFreezeCredit(long userId, long creditFreeze)
        {
            return SqlSetValues($"Credit = Credit - ({creditFreeze}), FreezeCredit = {SqlIsNull("FreezeCredit", "0")} + ({creditFreeze})", userId);
        }

        public static async Task<long> GetFreezeCreditAsync(long qq, IDbTransaction? trans = null) => await GetLongAsync("FreezeCredit", qq, null, trans);

        public static async Task<long> GetFreezeCreditForUpdateAsync(long qq, IDbTransaction? trans = null)
        {
            return await GetForUpdateAsync<long>("FreezeCredit", qq, null, 0, trans);
        }

        //冻结积分 (重构异步版)
        public static async Task<int> FreezeCreditAsync(long botUin, long groupId, string groupName, long qq, string name, long creditFreeze)
        {
            using var wrapper = await BeginTransactionAsync();
            try
            {
                // 1. 获取当前积分并加锁
                long creditValue = await GetCreditForUpdateAsync(botUin, groupId, qq, wrapper.Transaction);
                if (creditValue < creditFreeze)
                {
                    await wrapper.RollbackAsync();
                    return -1;
                }

                var (sql, paras) = SqlFreezeCredit(qq, creditFreeze);
                await ExecAsync(sql, wrapper.Transaction, paras);
                await CreditLog.AddLogAsync(botUin, groupId, groupName, qq, name, -creditFreeze, creditValue, "冻结积分", wrapper.Transaction);
                
                await wrapper.CommitAsync();

                await SyncCacheFieldAsync(qq, groupId, "Credit", creditValue - creditFreeze);
                return 0;
            }
            catch (Exception ex)
            {
                await wrapper.RollbackAsync();
                Logger.Error($"[FreezeCredit Error] {ex.Message}");
                return -1;
            }
        }

        //解冻积分 (重构异步版)
        public static async Task<int> UnfreezeCreditAsync(long botUin, long groupId, string groupName, long qq, string name, long creditUnfreeze)
        {
            using var wrapper = await BeginTransactionAsync();
            try
            {
                // 1. 获取当前冻结积分并加锁
                long freezeValue = await GetFreezeCreditForUpdateAsync(qq, wrapper.Transaction);
                long creditValue = await GetCreditForUpdateAsync(botUin, groupId, qq, wrapper.Transaction);
                if (freezeValue < creditUnfreeze)
                {
                    await wrapper.RollbackAsync();
                    return -1;
                }

                var (sql, paras) = SqlFreezeCredit(qq, -creditUnfreeze);
                await ExecAsync(sql, wrapper.Transaction, paras);
                await CreditLog.AddLogAsync(botUin, groupId, groupName, qq, name, creditUnfreeze, creditValue, "解冻积分", wrapper.Transaction);

                await wrapper.CommitAsync();

                await SyncCacheFieldAsync(qq, groupId, "FreezeCredit", freezeValue - creditUnfreeze);
                return 0;
            }
            catch (Exception ex)
            {
                await wrapper.RollbackAsync();
                Logger.Error($"[UnfreezeCredit Error] {ex.Message}");
                return -1;
            }
        }

        public static async Task<long> GetCreditRankingAsync(long botUin, long groupId, long qq)
        {
            long creditValue = await GetCreditAsync(botUin, groupId, qq);
            
            // 1. 优先检查本群积分
            if (await GroupInfo.GetIsCreditAsync(groupId))
            {
                return await GroupMember.CountWhereAsync(
                    "[GroupId] = {0} AND [Credit] > {1}", 
                    groupId, creditValue) + 1;
            }

            // 2. 检查机器人全局积分
            if (await BotInfo.GetBoolAsync("IsCredit", botUin))
            {
                return await Friend.CountWhereAsync(
                    "[BotUin] = {0} AND [Credit] > {1} AND [Id] IN (SELECT [UserId] FROM " + GroupMember.FullName + " WHERE [GroupId] = {2})",
                    botUin, creditValue, groupId) + 1;
            }

            // 3. 默认全局排名 (仅限该群成员)
            return await CountWhereAsync(
                "[Credit] > {0} AND [Id] IN (SELECT [UserId] FROM " + GroupMember.FullName + " WHERE [GroupId] = {1})",
                creditValue, groupId) + 1;
        }

        public static async Task<long> GetCreditRankingAllAsync(long botUin, long qq)
        {
            long totalCredit = await GetTotalCreditAsync(botUin, qq);
            return await CountWhereAsync(
                "[Credit] + [SaveCredit] > {0}",
                totalCredit) + 1;
        }

        public static async Task<string> GetCreditListAsync(long groupId)
        {
            return await QueryResAsync(
                $"SELECT {SqlTop(10)} [UserId], [Credit] FROM {GroupMember.FullName} " +
                $"WHERE [GroupId] = {0} ORDER BY [Credit] DESC {SqlLimit(10)}",
                "【第{i}名】 [@:{0}] 积分：{1}\n",
                groupId);
        }

        public static async Task<string> GetCreditListAllAsync()
        {
            return await QueryResAsync(
                $"SELECT {SqlTop(10)} [Id], [Credit] + [SaveCredit] AS TotalCredit FROM {FullName} " +
                $"ORDER BY TotalCredit DESC {SqlLimit(10)}",
                "【第{i}名】 [@:{0}] 积分：{1}\n");
        }
    }
}
