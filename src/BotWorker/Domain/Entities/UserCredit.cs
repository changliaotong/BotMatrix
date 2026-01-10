using System.Data;
using System.Reflection;
using BotWorker.Domain.Models.Messages.BotMessages;
using BotWorker.Domain.Entities;
using BotWorker.Common.Extensions;
using BotWorker.Infrastructure.Persistence.ORM;

namespace BotWorker.Domain.Entities
{
    public partial class UserInfo : MetaDataGuid<UserInfo>
    {
        public static async Task<(int Result, long CreditValue, int LogId)> AddCreditAsync(long botUin, long groupId, string groupName, long qq, string name, long creditAdd, string creditInfo, IDbTransaction? trans = null)
        {
            var creditValue = await GetCreditAsync(botUin, groupId, qq, trans);
            
            using var wrapper = await BeginTransactionAsync(trans);
            try
            {
                // 1. 确保用户存在
                long ownerId = await GroupInfo.GetGroupOwnerAsync(groupId);
                await AppendAsync(botUin, groupId, qq, name, ownerId);

                // 2. 执行积分操作
                var (sql, paras) = await SqlAddCreditAsync(botUin, groupId, qq, creditAdd);
                await ExecAsync(sql, wrapper.Transaction, paras);

                // 3. 记录日志
                int logId = await CreditLog.AddLogAsync(botUin, groupId, groupName, qq, name, creditAdd, creditInfo, wrapper.Transaction);

                wrapper.Commit();

                SyncCacheField(qq, groupId, "Credit", creditValue + creditAdd);
                return (0, creditValue + creditAdd, logId);
            }
            catch (Exception ex)
            {
                wrapper.Rollback();
                Console.WriteLine($"[AddCredit Error] {ex.Message}");
                return (-1, creditValue, 0);
            }
        }

        //增加积分
        public static (int Result, long CreditValue) AddCredit(long botUin, long groupId, string groupName, long qq, string name, long creditAdd, string creditInfo)
        {
            var res = AddCreditAsync(botUin, groupId, groupName, qq, name, creditAdd, creditInfo).GetAwaiter().GetResult();
            return (res.Result, res.CreditValue);
        }

        public static (int, long) MinusCredit(long botUin, long groupId, string groupName, long qq, string name, long creditMinus, string creditInfo)
            => AddCredit(botUin, groupId, groupName, qq, name, -creditMinus, creditInfo);


        //增加积分sql
        public static (string, IDataParameter[]) SqlAddCredit(long botUin, long groupId, long userId, long creditPlus)
            => SqlAddCreditAsync(botUin, groupId, userId, creditPlus).GetAwaiter().GetResult();

        public static async Task<(string, IDataParameter[])> SqlAddCreditAsync(long botUin, long groupId, long userId, long creditPlus)
        {
            if (await GroupInfo.GetIsCreditAsync(groupId))
            {
                return GroupMember.SqlAddCredit(groupId, userId, creditPlus);
            }
            else if (await BotInfo.GetIsCreditAsync(botUin))
            {
                return Friend.SqlAddCredit(botUin, userId, creditPlus);
            }
            else
            {
                // 暂时保持同步 Exists，因为它调用频繁且是内存/主键检查
                if (await ExistsAsync(userId))
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
            string transferInfo)
        {
            // 1. 前置检查
            long senderCredit = await GetCreditAsync(botUin, groupId, senderId);
            if (senderCredit < creditMinus)
                return (-1, senderCredit, 0);

            // 2. 开启事务
            using var wrapper = await BeginTransactionAsync();
            try
            {
                // 3. 链式调用业务方法，全部复用同一个 trans (通过 AsyncLocal 自动传递)
                // 扣除发送者积分 (内部会自动记录日志)
                var res1 = await AddCreditAsync(botUin, groupId, groupName, senderId, senderName, -creditMinus, $"{transferInfo}扣分：{receiverId}", wrapper.Transaction);
                
                // 增加接收者积分 (内部会自动记录日志)
                var res2 = await AddCreditAsync(botUin, groupId, groupName, receiverId, receiverName, creditAdd, $"{transferInfo}加分：{senderId}", wrapper.Transaction);

                // 4. 提交事务
                wrapper.Commit();

                return (0, res1.CreditValue, res2.CreditValue);
            }
            catch (Exception ex)
            {
                wrapper.Rollback();
                Console.WriteLine($"[TransferCredit Error] {ex.Message}");
                return (-1, senderCredit, 0);
            }
        }

        //转账积分
        public static int TransferCredit(long botUin, long groupId, string groupName, long qq, string name, long qqTo, string nameTo, long creditMinus, long creditAdd, ref long creditValue, ref long creditValue2, string transferInfo)
        {
            var res = TransferCreditAsync(botUin, groupId, groupName, qq, name, qqTo, nameTo, creditMinus, creditAdd, transferInfo).GetAwaiter().GetResult();
            if (res.Result == 0)
            {
                creditValue = res.SenderCredit;
                creditValue2 = res.ReceiverCredit;
            }
            return res.Result;
        }


        public static async Task<long> GetCreditAsync(long botUin, long groupId, long qq, IDbTransaction? trans = null)
        {
            return groupId != 0 && await GroupInfo.GetIsCreditAsync(groupId)
                ? await GroupMember.GetGroupCreditAsync(groupId, qq, trans)
                : await GetCreditAsync(botUin, qq, trans);
        }

        public static long GetCredit(long botUin, long groupId, long qq)
        {
            return GetCreditAsync(botUin, groupId, qq).GetAwaiter().GetResult();
        }

        public static long GetCredit(long groupId, long qq)
        {
            return GetCreditAsync(BotInfo.BotUinDef, groupId, qq).GetAwaiter().GetResult();
        }

        public static async Task<long> GetCreditAsync(long userId, IDbTransaction? trans = null)
        {
            return await GetLongAsync("Credit", userId, null, trans);
        }

        public static long GetCredit(long userId)
        {
            return GetCreditAsync(userId).GetAwaiter().GetResult();
        }


        //读取积分
        public static async Task<long> GetCreditAsync(long botUin, long userId, IDbTransaction? trans = null)
        {
            return await BotInfo.GetIsCreditAsync(botUin) ? await Friend.GetCreditAsync(botUin, userId, trans) : await GetLongAsync("Credit", userId, null, trans);
        }

        //积分总额
        public static async Task<long> GetTotalCreditAsync(long botUin, long userId) => await GetCreditAsync(botUin, userId) + await GetSaveCreditAsync(botUin, userId);
        public static async Task<long> GetTotalCreditAsync(long botUin, long groupId, long userId) => await GetCreditAsync(botUin, groupId, userId) + await GetSaveCreditAsync(botUin, groupId, userId);
        public static async Task<long> GetSaveCreditAsync(long botUin, long userId)
        {
            return await BotInfo.GetIsCreditAsync(botUin)
                ? await Friend.GetSaveCreditAsync(botUin, userId)
                : await GetSaveCreditAsync(userId);
        }

        public static long GetSaveCredit(long botUin, long userId)
        {
            return GetSaveCreditAsync(botUin, userId).GetAwaiter().GetResult();
        }

        public static async Task<long> GetSaveCreditAsync(long botUin, long groupId, long qq)
        {
            return await GroupInfo.GetIsCreditAsync(groupId)
                ? await GroupMember.GetLongAsync("SaveCredit", groupId, qq)
                : await GetSaveCreditAsync(qq);
        }

        public static long GetSaveCredit(long botUin, long groupId, long qq)
        {
            return GetSaveCreditAsync(botUin, groupId, qq).GetAwaiter().GetResult();
        }

        public static async Task<long> GetSaveCreditAsync(long userId)
        {
            return await GetLongAsync("SaveCredit", userId);
        }

        public static long GetSaveCredit(long userId)
        {
            return GetSaveCreditAsync(userId).GetAwaiter().GetResult();
        }

        public static (string, IDataParameter[]) SqlSaveCredit(long botUin, long groupId, long userId, long creditSave)
            => SqlSaveCreditAsync(botUin, groupId, userId, creditSave).GetAwaiter().GetResult();

        public static async Task<(string, IDataParameter[])> SqlSaveCreditAsync(long botUin, long groupId, long userId, long creditSave)
        {
            return await GroupInfo.GetIsCreditAsync(groupId)
                ? GroupMember.SqlSaveCredit(groupId, userId, creditSave)
                : await BotInfo.GetIsCreditAsync(botUin) ? Friend.SqlSaveCredit(botUin, userId, creditSave)
                                 : SqlSetValues($"Credit = Credit - ({creditSave}), SaveCredit = {SqlIsNull("SaveCredit", "0")} + ({creditSave})", userId);
        }

        public static (string, IDataParameter[]) SqlFreezeCredit(long userId, long creditFreeze)
        {
            return SqlSetValues($"Credit = Credit - ({creditFreeze}), FreezeCredit = {SqlIsNull("FreezeCredit", "0")} + ({creditFreeze})", userId);
        }

        public static async Task<long> GetFreezeCreditAsync(long qq) => await GetLongAsync("FreezeCredit", qq);
        public static long GetFreezeCredit(long qq) => GetFreezeCreditAsync(qq).GetAwaiter().GetResult();

        //冻结积分 (重构异步版)
        public static async Task<int> FreezeCreditAsync(long botUin, long groupId, string groupName, long qq, string name, long creditFreeze)
        {
            long creditValue = await GetCreditAsync(botUin, groupId, qq);
            if (creditValue < creditFreeze) return -1;

            using var trans = await BeginTransactionAsync();
            try
            {
                var (sql, paras) = SqlFreezeCredit(qq, creditFreeze);
                await ExecAsync(sql, trans, paras);
                await CreditLog.AddLogAsync(botUin, groupId, groupName, qq, name, -creditFreeze, "冻结积分", trans);
                
                await trans.CommitAsync();
                SyncCacheField(qq, groupId, "Credit", creditValue - creditFreeze);
                return 0;
            }
            catch
            {
                await trans.RollbackAsync();
                return -1;
            }
        }

        //解冻积分 (重构异步版)
        public static async Task<int> UnfreezeCreditAsync(long botUin, long groupId, string groupName, long qq, string name, long creditUnfreeze)
        {
            long creditValue = await GetFreezeCreditAsync(qq);
            if (creditValue < creditUnfreeze) return -1;

            using var trans = await BeginTransactionAsync();
            try
            {
                var (sql, paras) = SqlFreezeCredit(qq, -creditUnfreeze);
                await ExecAsync(sql, trans, paras);
                await CreditLog.AddLogAsync(botUin, groupId, groupName, qq, name, creditUnfreeze, "解冻积分", trans);

                await trans.CommitAsync();
                SyncCacheField(qq, groupId, "FreezeCredit", creditValue - creditUnfreeze);
                return 0;
            }
            catch
            {
                await trans.RollbackAsync();
                return -1;
            }
        }

        public static int DoFreezeCredit(long botUin, long groupId, string groupName, long qq, string name, long creditFreeze)
        {
            return FreezeCreditAsync(botUin, groupId, groupName, qq, name, creditFreeze).GetAwaiter().GetResult();
        }

        public static int UnfreezeCredit(long botUin, long groupId, string groupName, long qq, string name, long creditUnfreeze)
        {
            return UnfreezeCreditAsync(botUin, groupId, groupName, qq, name, creditUnfreeze).GetAwaiter().GetResult();
        }

        public static long GetCreditRanking(long botUin, long groupId, long qq)
            => GetCreditRankingAsync(botUin, groupId, qq).GetAwaiter().GetResult();

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

        public static long GetCreditRankingAll(long botUin, long qq)
            => GetCreditRankingAllAsync(botUin, qq).GetAwaiter().GetResult();

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

        public static string GetCreditList(long groupId)
            => GetCreditListAsync(groupId).GetAwaiter().GetResult();

        public static async Task<string> GetCreditListAllAsync()
        {
            return await QueryResAsync(
                $"SELECT {SqlTop(10)} [Id], [Credit] + [SaveCredit] AS TotalCredit FROM {FullName} " +
                $"ORDER BY TotalCredit DESC {SqlLimit(10)}",
                "【第{i}名】 [@:{0}] 积分：{1}\n");
        }

        public static string GetCreditListAll()
            => GetCreditListAllAsync().GetAwaiter().GetResult();
    }
}
