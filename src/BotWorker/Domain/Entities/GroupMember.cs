using System;
using System.Text.RegularExpressions;
using System.Data;
using System.Threading.Tasks;
using BotWorker.Domain.Repositories;
using Dapper.Contrib.Extensions;
using Microsoft.Extensions.DependencyInjection;
using Newtonsoft.Json;

namespace BotWorker.Domain.Entities
{
    [Table("group_member")]
    public partial class GroupMember
    {
        private static IGroupMemberRepository Repository => 
            BotMessage.ServiceProvider?.GetRequiredService<IGroupMemberRepository>() 
            ?? throw new InvalidOperationException("IGroupMemberRepository not registered");

        private static IUserRepository UserRepository => 
            BotMessage.ServiceProvider?.GetRequiredService<IUserRepository>() 
            ?? throw new InvalidOperationException("IUserRepository not registered");

        private static IGroupRepository GroupRepository => 
            BotMessage.ServiceProvider?.GetRequiredService<IGroupRepository>() 
            ?? throw new InvalidOperationException("IGroupRepository not registered");

        // Helper to replace MetaData.BeginTransactionAsync
        private static async Task<BotWorker.Infrastructure.Persistence.TransactionWrapper> BeginTransactionAsync(IDbTransaction? trans = null)
        {
            return await BotWorker.Infrastructure.Persistence.TransactionWrapper.BeginTransactionAsync(trans);
        }

        [Key]
        public long Id { get; set; }

        public long GroupId { get; set; }
        public long UserId { get; set; }

        [JsonIgnore]
        [HighFrequency]
        public long GroupCredit { get; set; }
        [JsonIgnore]
        [HighFrequency]
        public long GoldCoins { get; set; }
        [JsonIgnore]
        [HighFrequency]
        public long BlackCoins { get; set; }
        [JsonIgnore]
        [HighFrequency]
        public long PurpleCoins { get; set; }
        [JsonIgnore]
        [HighFrequency]
        public long GameCoins { get; set; }
        [JsonIgnore]
        public int SignTimes { get; set; }
        [JsonIgnore]
        public int SignLevel { get; set; }
        [JsonIgnore]
        public DateTime SignDate { get; set; }
        [JsonIgnore]
        public int SignTimesAll { get; set; }

        public DateTime UpdatedAt { get; set; }

        public static async Task<string> AddCoinsResAsync(BotMessage botMsg)
        {
            var regexCmd = Regexs.AddMinus;
            return await AddCoinsResAsync(botMsg.SelfId, botMsg.GroupId, botMsg.GroupName, botMsg.UserId, botMsg.Name,
                botMsg.CurrentMessage.RegexGetValue(regexCmd, "CmdName"),
                botMsg.CurrentMessage.RegexGetValue(regexCmd, "cmdPara"),
                botMsg.CurrentMessage.RegexGetValue(regexCmd, "cmdPara2"),
                botMsg.CurrentMessage.RegexGetValue(regexCmd, "cmdPara3"));
        }

        //充值/扣除 积分 金币 黑金币 紫币 游戏币等 (异步重构版)
        public static async Task<string> AddCoinsResAsync(long botUin, long groupId, string groupName, long qq, string name, string cmdName, string cmdPara, string cmdPara2, string cmdPara3)
        {
            if (!await GroupRepository.IsOwnerAsync(groupId, qq))
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

            if (coins_type == (int)CoinsLog.CoinsType.groupCredit && !await GroupRepository.GetIsCreditAsync(groupId))
                return $"没有开启本群积分";

            // 使用事务进行异步链式调用
            using var wrapper = await BeginTransactionAsync();
            try
            {
                // 1. 确保用户存在 (按 ID 从小到大执行，防止 User 表死锁)
                long firstId = qq < coins_qq ? qq : coins_qq;
                long secondId = qq < coins_qq ? coins_qq : qq;
                string firstName = firstId == qq ? name : await UserRepository.GetValueAsync<string>("Name", firstId, trans: wrapper.Transaction); 
                string secondName = secondId == qq ? name : await UserRepository.GetValueAsync<string>("Name", secondId, trans: wrapper.Transaction);

                long ownerId = await GroupRepository.GetGroupOwnerAsync(groupId, 0, wrapper.Transaction);
                await UserRepository.AppendAsync(botUin, groupId, firstId, firstName, ownerId, trans: wrapper.Transaction);
                await UserRepository.AppendAsync(botUin, groupId, secondId, secondName, ownerId, trans: wrapper.Transaction);

                // 2. 统一加锁顺序，防止 GroupMember 表死锁 (按 ID 从小到大锁定)
                if (firstId == qq)
                {
                    await UserRepository.GetCreditForUpdateAsync(botUin, credit_group, qq, wrapper.Transaction);
                    if (qq != coins_qq)
                        await GetCoinsForUpdateAsync(coins_type, groupId, coins_qq, wrapper.Transaction);
                }
                else
                {
                    await GetCoinsForUpdateAsync(coins_type, groupId, coins_qq, wrapper.Transaction);
                    await UserRepository.GetCreditForUpdateAsync(botUin, credit_group, qq, wrapper.Transaction);
                }

                // 在事务内部获取分值
                long credit_value_locked = await UserRepository.GetCreditAsync(botUin, credit_group, qq, wrapper.Transaction);
                if (cmdName == "充值")
                {
                    if (credit_value_locked < minus_credit)
                    {
                        await wrapper.RollbackAsync();
                        return $"您有{credit_value_locked}分不足{minus_credit}，请先兑换";
                    }
                }
                else //扣除
                {
                    long coins_value_locked = await GetCoinsAsync(coins_type, groupId, coins_qq, wrapper.Transaction);
                    if (coins_value_locked < coins_oper)
                    {
                        await wrapper.RollbackAsync();
                        return $"[@:{coins_qq}]{cmdPara}{coins_value_locked}不足{coins_oper}，无法扣除";
                    }

                    minus_credit = -minus_credit;
                    coins_oper = -coins_oper;
                }

                // 3. 执行业务操作
                var res1 = await UserInfo.AddCreditAsync(botUin, credit_group, groupName, qq, name, -minus_credit, $"{cmdName}{cmdPara}*{coins_oper}", wrapper.Transaction);
                var res2 = await AddCoinsAsync(botUin, groupId, coins_qq, "", coins_type, coins_oper, $"{cmdName}{cmdPara}*{coins_oper}", wrapper.Transaction);

                await wrapper.CommitAsync();

                // 同步缓存
                await UserInfo.SyncCreditCacheAsync(botUin, credit_group, qq, res1.CreditValue);
                await SyncCacheFieldAsync(groupId, coins_qq, CoinsLog.conisFields[coins_type], res2.CoinsValue);

                return $"{cmdName}{cmdPara}：{Math.Abs(coins_oper)}成功！\n[@:{coins_qq}]{cmdPara}:{res2.CoinsValue}\n您：{-minus_credit}分，累计：{res1.CreditValue}";
            }
            catch (Exception ex)
            {
                await wrapper.RollbackAsync();
                Logger.Error($"[AddCoinsRes Error] {ex.Message}");
                return RetryMsg;
            }
        }

        public static async Task<string> ExchangeCoinsAsync(long botUin, long groupId, string groupName, long qq, string name, int coins_type, string cmdName, string cmdPara, long minus_credit, long coins_oper, long coins_qq)
        {
            long credit_group = groupId; // 简化处理，通常积分群就是当前群
            if (coins_type == (int)CoinsLog.CoinsType.groupCredit)
            {
                if (!await GroupRepository.GetIsCreditAsync(groupId))
                    return $"没有开启本群积分";
            }

            using var wrapper = await BeginTransactionAsync();
            try
            {
                // 1. 确保用户存在 (按 ID 从小到大执行，防止 User 表死锁)
                long firstId = qq < coins_qq ? qq : coins_qq;
                long secondId = qq < coins_qq ? coins_qq : qq;
                string firstName = firstId == qq ? name : await UserRepository.GetValueAsync<string>("Name", firstId, trans: wrapper.Transaction);
                string secondName = secondId == qq ? name : await UserRepository.GetValueAsync<string>("Name", secondId, trans: wrapper.Transaction);

                long ownerId = await GroupRepository.GetGroupOwnerAsync(groupId, 0, wrapper.Transaction);
                await UserRepository.AppendAsync(botUin, groupId, firstId, firstName, ownerId, trans: wrapper.Transaction);
                await UserRepository.AppendAsync(botUin, groupId, secondId, secondName, ownerId, trans: wrapper.Transaction);

                // 2. 统一加锁顺序，防止 GroupMember 表死锁 (按 ID 从小到大锁定)
                if (firstId == qq)
                {
                    await UserRepository.GetCreditForUpdateAsync(botUin, credit_group, qq, wrapper.Transaction);
                    if (qq != coins_qq)
                        await GetCoinsForUpdateAsync(coins_type, groupId, coins_qq, wrapper.Transaction);
                }
                else
                {
                    await GetCoinsForUpdateAsync(coins_type, groupId, coins_qq, wrapper.Transaction);
                    await UserRepository.GetCreditForUpdateAsync(botUin, credit_group, qq, wrapper.Transaction);
                }

                // 在事务内部获取分值
                long credit_value_locked = await UserRepository.GetCreditAsync(botUin, credit_group, qq, wrapper.Transaction);
                if (cmdName == "充值")
                {
                    if (credit_value_locked < minus_credit)
                    {
                        await wrapper.RollbackAsync();
                        return $"您有{credit_value_locked}分不足{minus_credit}，请先兑换";
                    }
                }
                else //扣除
                {
                    long coins_value_locked = await GetCoinsAsync(coins_type, groupId, coins_qq, wrapper.Transaction);
                    if (coins_value_locked < minus_credit)
                    {
                        await wrapper.RollbackAsync();
                        return $"[@:{coins_qq}]{cmdPara}{coins_value_locked}不足{coins_oper}，无法扣除";
                    }

                    minus_credit = -minus_credit;
                    coins_oper = -coins_oper;
                }

                // 3. 执行业务操作
                var res1 = await UserInfo.AddCreditAsync(botUin, credit_group, groupName, qq, name, -minus_credit, $"{cmdName}{cmdPara}*{coins_oper}", wrapper.Transaction);
                var res2 = await AddCoinsAsync(botUin, groupId, coins_qq, "", coins_type, coins_oper, $"{cmdName}{cmdPara}*{coins_oper}", wrapper.Transaction);

                await wrapper.CommitAsync();

                // 同步缓存
                await UserInfo.SyncCreditCacheAsync(botUin, credit_group, qq, res1.CreditValue);
                await SyncCacheFieldAsync(groupId, coins_qq, CoinsLog.conisFields[coins_type], res2.CoinsValue);

                return $"{cmdName}{cmdPara}：{Math.Abs(coins_oper)}成功！\n[@:{coins_qq}]{cmdPara}:{res2.CoinsValue}\n您：{-minus_credit}分，累计：{res1.CreditValue}";
            }
            catch (Exception ex)
            {
                await wrapper.RollbackAsync();
                Logger.Error($"[ExchangeCoins Error] {ex.Message}");
                return RetryMsg;
            }
        }

        // 本群积分
        public static async Task<long> GetGroupCreditAsync(long groupId, long qq, IDbTransaction? trans = null)
        {
            return await Repository.GetCoinsAsync((int)CoinsLog.CoinsType.groupCredit, groupId, qq, trans);
        }

        public static async Task<long> GetGroupCreditForUpdateAsync(long groupId, long qq, IDbTransaction? trans = null)
        {
            if (trans == null) throw new ArgumentNullException(nameof(trans), "GetGroupCreditForUpdateAsync requires an active transaction");
            return await Repository.GetForUpdateAsync(CoinsLog.conisFields[(int)CoinsLog.CoinsType.groupCredit], groupId, qq, trans);
        }

        public static async Task<long> GetSaveCreditForUpdateAsync(long groupId, long qq, IDbTransaction? trans = null)
        {
            if (trans == null) throw new ArgumentNullException(nameof(trans), "GetSaveCreditForUpdateAsync requires an active transaction");
            return await Repository.GetForUpdateAsync("SaveCredit", groupId, qq, trans);
        }

        // 金币余额
        public static async Task<long> GetCoinsAsync(int coinsType, long groupId, long qq, IDbTransaction? trans = null)
        {
            return await Repository.GetCoinsAsync(coinsType, groupId, qq, trans);
        }

        public static async Task<long> GetCoinsForUpdateAsync(int coinsType, long groupId, long qq, IDbTransaction? trans = null)
        {
            if (trans == null) throw new ArgumentNullException(nameof(trans), "GetCoinsForUpdateAsync requires an active transaction");
            return await Repository.GetCoinsForUpdateAsync(coinsType, groupId, qq, trans);
        }

        // 金币余额
        public static async Task<long> GetGoldCoinsAsync(long groupId, long qq)
        {
            return await Repository.GetLongAsync("GoldCoins", groupId, qq);
        }

        // 紫币
        public static async Task<long> GetPurpleCoinsAsync(long groupId, long qq)
        {
            return await Repository.GetLongAsync("PurpleCoins", groupId, qq);
        }

        // 黑金币
        public static async Task<long> GetBlackCoinsAsync(long groupId, long qq)
        {
            return await Repository.GetLongAsync("BlackCoins", groupId, qq);
        }

        // 游戏币
        public static async Task<long> GetGameCoinsAsync(long groupId, long qq)
        {
            return await Repository.GetLongAsync("GameCoins", groupId, qq);
        }

        // 加金币/黑金币/紫币/游戏币 (异步事务版)
        public static async Task<(int Result, long CoinsValue, int LogId)> AddCoinsAsync(long botUin, long groupId, long qq, string name, int coinsType, long coinsAdd, string coinsInfo, IDbTransaction? trans = null)
        {
            try
            {
                // 1. 确保用户存在
                if (!await Repository.ExistsAsync(groupId, qq, trans) && await AppendAsync(groupId, qq, name, trans: trans) == -1)
                    return (-1, 0, 0);

                // 2. 获取当前准确值并加锁
                using var wrapper = await BeginTransactionAsync(trans);
                var coinsValue = await GetCoinsForUpdateAsync(coinsType, groupId, qq, wrapper.Transaction);

                // 3. 如果是扣除金币，检查是否足够
                if (coinsAdd < 0 && coinsValue < Math.Abs(coinsAdd))
                {
                    return (-2, coinsValue, 0); // -2 表示余额不足
                }

                // 4. 更新金币
                await Repository.IncrementValueAsync(CoinsLog.conisFields[coinsType], coinsAdd, groupId, qq, wrapper.Transaction);
                
                // 5. 记录日志
                int logId = await CoinsLog.AddLogAsync(botUin, groupId, "", qq, name, coinsType, coinsAdd, coinsValue, coinsInfo, wrapper.Transaction);
                
                if (trans == null) await wrapper.CommitAsync();

                long finalCoinsValue = coinsValue + coinsAdd;
                await SyncCacheFieldAsync(groupId, qq, CoinsLog.conisFields[coinsType], finalCoinsValue);

                return (0, finalCoinsValue, logId);
            }
            catch (Exception ex)
            {
                Logger.Error($"[AddCoins Error] {ex.Message}");
                if (trans != null) throw;
                return (-1, 0, 0);
            }
        }

        public static async Task<(int Result, long CoinsValue, int LogId)> AddCoinsTransAsync(long botUin, long groupId, long qq, string name, int coinsType, long coinsAdd, string coinsInfo, IDbTransaction? trans = null)
        {
            using var wrapper = await BeginTransactionAsync(trans);
            try
            {
                var res = await AddCoinsAsync(botUin, groupId, qq, name, coinsType, coinsAdd, coinsInfo, wrapper.Transaction);
                await wrapper.CommitAsync();

                await SyncCacheFieldAsync(groupId, qq, CoinsLog.conisFields[coinsType], res.CoinsValue);
                return res;
            }
            catch (Exception ex)
            {
                Logger.Error($"[AddCoinsTrans Error] {ex.Message}");
                await wrapper.RollbackAsync();
                if (trans != null) throw;
                return (-1, 0, 0);
            }
        }

        //转账金币
        public static async Task<(int Result, long SenderCoins, long ReceiverCoins)> TransferCoinsAsync(long botUin, long groupId, long qq, string name, long qqTo, string nameTo, int coinsType, long coinsMinus, long coinsAdd, string transferInfo)
        {
            using var wrapper = await BeginTransactionAsync();
            try
            {
                // 1. 确保用户存在 (按 ID 从小到大执行，防止 User 表死锁)
                long firstId = qq < qqTo ? qq : qqTo;
                long secondId = qq < qqTo ? qqTo : qq;
                string firstName = firstId == qq ? name : await UserRepository.GetValueAsync<string>("Name", firstId, trans: wrapper.Transaction);
                string secondName = secondId == qq ? name : await UserRepository.GetValueAsync<string>("Name", secondId, trans: wrapper.Transaction);

                long ownerId = await GroupRepository.GetGroupOwnerAsync(groupId, 0, wrapper.Transaction);
                await UserInfo.AppendAsync(botUin, groupId, firstId, firstName, ownerId, trans: wrapper.Transaction);
                await UserInfo.AppendAsync(botUin, groupId, secondId, secondName, ownerId, trans: wrapper.Transaction);

                // 2. 统一加锁顺序，防止 GroupMember 表死锁 (按 ID 从小到大锁定)
                if (firstId == qq)
                {
                    await GetCoinsForUpdateAsync(coinsType, groupId, qq, wrapper.Transaction);
                    await GetCoinsForUpdateAsync(coinsType, groupId, qqTo, wrapper.Transaction);
                }
                else
                {
                    await GetCoinsForUpdateAsync(coinsType, groupId, qqTo, wrapper.Transaction);
                    await GetCoinsForUpdateAsync(coinsType, groupId, qq, wrapper.Transaction);
                }

                // 获取发送者金币，检查是否足够
                var currentCoins = await GetCoinsAsync(coinsType, groupId, qq, wrapper.Transaction);
                if (currentCoins < coinsMinus)
                {
                    await wrapper.RollbackAsync();
                    return (-2, currentCoins, 0); // -2 表示余额不足
                }

                // 3. 执行业务操作
                var res1 = await AddCoinsAsync(botUin, groupId, qq, name, coinsType, -coinsMinus, $"{transferInfo}转出:{qqTo}", wrapper.Transaction);
                var res2 = await AddCoinsAsync(botUin, groupId, qqTo, nameTo, coinsType, coinsAdd, $"{transferInfo}转入:{qq}", wrapper.Transaction);

                await wrapper.CommitAsync();

                await SyncCacheFieldAsync(groupId, qq, CoinsLog.conisFields[coinsType], res1.CoinsValue);
                await SyncCacheFieldAsync(groupId, qqTo, CoinsLog.conisFields[coinsType], res2.CoinsValue);

                return (0, res1.CoinsValue, res2.CoinsValue);
            }
            catch (Exception ex)
            {
                Logger.Error($"[TransferCoins Error] {ex.Message}");
                await wrapper.RollbackAsync();
                return (-1, 0, 0);
            }
        }

        public static async Task<int> AddCreditAsync(long groupId, long userId, long creditAdd, IDbTransaction? trans = null)
        {
            if (await Repository.ExistsAsync(groupId, userId, trans))
            {
                return await Repository.IncrementValueAsync("GroupCredit", creditAdd, groupId, userId, trans);
            }
            else
            {
                return await Repository.AppendAsync(groupId, userId, "", groupCredit: creditAdd, trans: trans);
            }
        }

        // 添加群成员 (异步版本)
        public static int Append(long groupId, long userId, string name) 
            => AppendAsync(groupId, userId, name).GetAwaiter().GetResult();

        public static async Task<int> AppendAsync(long groupId, long userId, string name, string displayName = "", long groupCredit = 0, string confirmCode = "", IDbTransaction? trans = null)
        {
            if (userId.In(2107992324, 3677524472, 3662527857, 2174158062, 2188157235, 3375620034, 1611512438, 3227607419, 3586811032,
                3835195413, 3527470977, 3394199803, 2437953621, 3082166471, 2375832958, 1807139582, 2704647312, 1420694846, 3788007880)) return 0;

            return await Repository.AppendAsync(groupId, userId, name, displayName, groupCredit, confirmCode, trans);
        }

        //上下分 (异步重构版)
        public static async Task<string> GetShangFenAsync(long botUin, long groupId, string groupName, long userId, string cmdName, string cmdPara)
        {
            if (!await GroupRepository.IsOwnerAsync(groupId, userId) || !await BotInfo.IsAdminAsync(botUin, userId))
                return OwnerOnlyMsg;

            if (await Income.TotalAsync(userId) < 400)            
                return "您无权使用此命令，请联系客服";
            
            if (await BotInfo.GetIsCreditAsync(botUin) || await GroupRepository.GetIsCreditAsync(groupId))
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

                using var wrapper = await BeginTransactionAsync();
                try
                {
                    // 1. 确保用户存在
                    long ownerId = await GroupRepository.GetGroupOwnerAsync(groupId, 0, wrapper.Transaction);
                    await UserRepository.AppendAsync(botUin, groupId, creditQQ, "", ownerId, trans: wrapper.Transaction);

                    var creditValue = await UserRepository.GetCreditForUpdateAsync(botUin, groupId, creditQQ, wrapper.Transaction);

                    if (cmdName == "下分")
                    {
                        if (creditValue < creditAdd)
                        {
                            await wrapper.RollbackAsync();
                            return $"对方只有{creditValue}分";
                        }
                        creditAdd = -creditAdd;
                    }

                    var res = await UserInfo.AddCreditAsync(botUin, groupId, groupName, creditQQ, "", creditAdd, cmdName, wrapper.Transaction);
                    if (res.Result == -1)
                    {
                        await wrapper.RollbackAsync();
                        return RetryMsg;
                    }

                    await wrapper.CommitAsync();

                    await UserInfo.SyncCreditCacheAsync(botUin, groupId, creditQQ, res.CreditValue);

                    return $"[@:{creditQQ}] {cmdName}成功！\n积分：{Math.Abs(creditAdd)}，累计：{res.CreditValue}";
                }
                catch (Exception ex)
                {
                    await wrapper.RollbackAsync();
                    Logger.Error($"[GetShangFen Error] {ex.Message}");
                    return RetryMsg;
                }
            }
            else
                return $"此群未开通本群积分，不能上下分";
        }

        // 获取签到次数
        public static async Task<int> GetSignTimesAsync(long groupId, long userId)
        {
            return await GetIntAsync("SignTimes", groupId, userId);
        }

        // 获取签到列表
        public static async Task<string> GetSignListAsync(long groupId, int topN = 10)
        {
            return await Repository.GetSignListAsync(groupId, topN);
        }

        // 积分提现 (异步版)
        public static async Task<string> WithdrawCreditAsync(long botUin, long groupId, string groupName, long userId, string name, long withdrawAmount)
        {
            if (withdrawAmount < 100) return "提现金额不能低于100";

            using var wrapper = await BeginTransactionAsync();
            try
            {
                // 1. 获取积分并加锁
                long creditValue = await UserInfo.GetCreditForUpdateAsync(botUin, groupId, userId, wrapper.Transaction);
                if (creditValue < withdrawAmount)
                {
                    await wrapper.RollbackAsync();
                    return $"您的积分{creditValue}不足{withdrawAmount}";
                }

                // 2. 扣除积分
                var res = await UserInfo.AddCreditAsync(botUin, groupId, groupName, userId, name, -withdrawAmount, "积分提现", wrapper.Transaction);
                if (res.Result == -1) throw new Exception("扣除积分失败");

                // 3. 增加余额
                await UserInfo.IncrementValueAsync("Balance", (decimal)withdrawAmount / 100, userId, wrapper.Transaction);

                await wrapper.CommitAsync();
                
                // 4. 同步缓存
                await UserInfo.SyncCreditCacheAsync(botUin, groupId, userId, res.CreditValue);
                UserInfo.SyncCacheField(userId, "Balance", await UserInfo.GetBalanceAsync(userId));

                return $"✅ 提现成功！\n提现积分：{withdrawAmount}\n到账余额：{(decimal)withdrawAmount / 100:N2}\n剩余积分：{res.CreditValue}";
            }
            catch (Exception ex)
            {
                await wrapper.RollbackAsync();
                Logger.Error($"[WithdrawCredit Error] {ex.Message}");
                return RetryMsg;
            }
        }
        // 金币排行榜
        public static async Task<string> GetCoinsRankingAsync(long groupId, long userId)
        {
            return (await Repository.GetCoinsRankingAsync(groupId, userId)).ToString();
        }

        public static async Task<string> GetCoinsRankingAllAsync(long userId)
        {
            return (await Repository.GetCoinsRankingAllAsync(userId)).ToString();
        }

        // 获取签到日期差 (异步版)
        public static async Task<int> GetSignDateDiffAsync(long groupId, long userId)
        {
            return await Repository.GetSignDateDiffAsync(groupId, userId);
        }

        // 更新签到信息 SQL
        public static async Task<bool> UpdateSignInfoAsync(long groupId, long userId, int signTimes, int signLevel, IDbTransaction? trans = null)
        {
            return await Repository.UpdateSignInfoAsync(groupId, userId, signTimes, signLevel, trans);
        }

        public static async Task SyncCacheFieldAsync(long groupId, long userId, string field, object value)
        {
            await Repository.SyncCacheFieldAsync(groupId, userId, field, value);
        }

        public static void SyncCacheField(long groupId, long userId, string field, object value)
        {
            SyncCacheFieldAsync(groupId, userId, field, value).GetAwaiter().GetResult();
        }
    }
}
