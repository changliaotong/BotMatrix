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
        //增加积分 (支持事务)
        public static async Task<(int Result, long CreditValue, int LogId)> AddCreditAsync(long botUin, long groupId, string groupName, long qq, string name, long creditAdd, string creditInfo, IDbTransaction? trans = null)
        {
            var creditValue = await GetCreditAsync(groupId, qq);
            
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
                if (await AppendAsync(botUin, groupId, qq, name, await GroupInfo.GetGroupOwnerAsync(groupId)) == -1)
                    return (-1, creditValue, 0);

                // 2. 执行积分操作
                var (sql, paras) = await SqlAddCreditAsync(botUin, groupId, qq, creditAdd);
                await ExecAsync(sql, trans, paras);

                // 3. 记录日志
                int logId = await CreditLog.AddLogAsync(botUin, groupId, groupName, qq, name, creditAdd, creditInfo, trans);

                if (isNewTrans)
                    await trans.CommitAsync();

                SyncCacheField(qq, groupId, "Credit", creditValue + creditAdd);
                return (0, creditValue + creditAdd, logId);
            }
            catch
            {
                if (isNewTrans)
                    await trans.RollbackAsync();
                return (-1, creditValue, 0);
            }
            finally
            {
                if (isNewTrans)
                {
                    trans.Connection?.Close();
                    trans.Dispose();
                }
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
                    return SqlInsert(new List<Cov> {
                        new Cov("BotUin", botUin),
                        new Cov("GroupId", groupId),
                        new Cov("Id", userId),
                        new Cov("Credit", creditPlus),
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
            long senderCredit = await GetCreditAsync(groupId, senderId);
            if (senderCredit < creditMinus)
                return (-1, senderCredit, 0);

            // 2. 开启事务
            using var trans = await BeginTransactionAsync();
            try
            {
                // 3. 链式调用业务方法，全部复用同一个 trans
                // 扣除发送者积分 (内部会自动记录日志)
                var res1 = await AddCreditAsync(botUin, groupId, groupName, senderId, senderName, -creditMinus, $"{transferInfo}扣分：{receiverId}", trans);
                
                // 增加接收者积分 (内部会自动记录日志)
                var res2 = await AddCreditAsync(botUin, groupId, groupName, receiverId, receiverName, creditAdd, $"{transferInfo}加分：{senderId}", trans);

                // 4. 提交事务
                await trans.CommitAsync();

                return (0, res1.CreditValue, res2.CreditValue);
            }
            catch (Exception ex)
            {
                await trans.RollbackAsync();
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


        public static long GetCredit(long groupId, long qq)
        {
            return GetCreditAsync(BotInfo.BotUinDef, groupId, qq).GetAwaiter().GetResult();
        }

        public static async Task<long> GetCreditAsync(long botUin, long groupId, long qq)
        {
            return groupId != 0 && await GroupInfo.GetIsCreditAsync(groupId)
                ? await GroupMember.GetGroupCreditAsync(groupId, qq)
                : await GetCreditAsync(botUin, qq);
        }

        public static long GetCredit(long botUin, long groupId, long qq)
        {
            return GetCreditAsync(botUin, groupId, qq).GetAwaiter().GetResult();
        }

        public static async Task<long> GetCreditAsync(long userId)
        {
            return await GetLongAsync("Credit", userId);
        }

        public static long GetCredit(long userId)
        {
            return GetCreditAsync(userId).GetAwaiter().GetResult();
        }


        //读取积分
        public static async Task<long> GetCreditAsync(long botUin, long userId)
        {
            return await BotInfo.GetIsCreditAsync(botUin) ? await Friend.GetCreditAsync(botUin, userId) : await GetLongAsync("credit", userId);
        }

        //积分总额
        public static async Task<long> GetTotalCreditAsync(long userId) => await GetCreditAsync(userId) + await GetSaveCreditAsync(userId);
        public static long GetTotalCredit(long userId) => GetTotalCreditAsync(userId).GetAwaiter().GetResult();

        public static async Task<long> GetTotalCreditAsync(long groupId, long qq) => await GetCreditAsync(groupId, qq) + await GetSaveCreditAsync(groupId, qq);
        public static long GetTotalCredit(long groupId, long qq) => GetTotalCreditAsync(groupId, qq).GetAwaiter().GetResult();

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
            long creditValue = await GetCreditAsync(groupId, qq);
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
            long credit_value = await GetCreditAsync(groupId, qq);
            return await GroupInfo.GetIsCreditAsync(groupId)
                ? await GroupMember.CountWhereAsync($"GroupId = {groupId} and Credit > {credit_value}") + 1
                : await BotInfo.GetBoolAsync("IsCredit", botUin)
                    ? await Friend.CountWhereAsync($"BotUin = {botUin} and Credit > {credit_value} and Id in (select UserId from {GroupMember.FullName} where GroupId = {groupId})") + 1
                    : await CountWhereAsync($"Credit > {credit_value} and Id in (select UserId from {GroupMember.FullName} where GroupId = {groupId})") + 1;
        }

        public static long GetCreditRankingAll(long qq)
            => GetCreditRankingAllAsync(qq).GetAwaiter().GetResult();

        public static async Task<long> GetCreditRankingAllAsync(long qq)
        {
            return await CountWhereAsync($"Credit + SaveCredit > {await GetTotalCreditAsync(qq)}") + 1;
        }

        public static string GetCreditList(long groupId)
        {
            return QueryRes($"select top 10 UserId, Credit from {GroupMember.FullName} where GroupId = {groupId} order by Credit desc",
                "【第{i}名】 [@:{0}] 积分：{1}\n");
        }

        public static string GetCreditListAll()
        {
            return QueryRes($"select top 10 Id, Credit + SaveCredit as TotalCredit from {FullName} order by TotalCredit desc",
                "【第{i}名】 [@:{0}] 积分：{1}\n");
        }
    }
}
