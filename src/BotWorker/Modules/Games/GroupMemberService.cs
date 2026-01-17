using System;
using System.Data;
using System.Threading.Tasks;
using System.Text.RegularExpressions;
using BotWorker.Domain.Entities;
using BotWorker.Domain.Interfaces;
using BotWorker.Domain.Repositories;
using BotWorker.Common;
using Dapper;
using Microsoft.Extensions.Logging;

namespace BotWorker.Modules.Games
{
    public class GroupMemberService : IGroupMemberService
    {
        private readonly IGroupMemberRepository _groupMemberRepo;
        private readonly IUserRepository _userRepo;
        private readonly IGroupRepository _groupRepo;
        private readonly IBotRepository _botRepo;
        private readonly ICoinsLogRepository _coinsLogRepo;
        private readonly IIncomeRepository _incomeRepo;
        private readonly ILogger<GroupMemberService> _logger;

        public GroupMemberService(
            IGroupMemberRepository groupMemberRepo,
            IUserRepository userRepo,
            IGroupRepository groupRepo,
            IBotRepository botRepo,
            ICoinsLogRepository coinsLogRepo,
            IIncomeRepository incomeRepo,
            ILogger<GroupMemberService> logger)
        {
            _groupMemberRepo = groupMemberRepo;
            _userRepo = userRepo;
            _groupRepo = groupRepo;
            _botRepo = botRepo;
            _coinsLogRepo = coinsLogRepo;
            _incomeRepo = incomeRepo;
            _logger = logger;
        }

        private const string OwnerOnlyMsg = "只有主人或群主才能使用此功能";
        private const string RetryMsg = "系统繁忙，请稍后再试";

        public async Task<string> AddCoinsResAsync(long botUin, long groupId, string groupName, long qq, string name, string cmdName, string cmdPara, string cmdPara2, string cmdPara3)
        {
            if (!await _groupRepo.IsOwnerAsync(groupId, qq) && !await _userRepo.GetIsSuperAsync(qq))
                return OwnerOnlyMsg;

            long coinsQQ = 0;
            long coinsAdd = 0;
            string regexCoins = "";

            if (cmdPara.IsMatch(Regexs.CoinsParaAt))
                regexCoins = Regexs.CoinsParaAt;
            else if (cmdPara.IsMatch(Regexs.CoinsParaAt2))
                regexCoins = Regexs.CoinsParaAt2;
            else if (cmdPara.IsMatch(Regexs.CoinsPara))
                regexCoins = Regexs.CoinsPara;
            else
                return $"格式：{cmdName} + QQ + 数量\n例如：{cmdName} {qq} 5000";

            foreach (Match match in cmdPara.Matches(regexCoins))
            {
                coinsQQ = match.Groups["UserId"].Value.AsLong();
                coinsAdd = match.Groups["coins"].Value.AsLong();
            }

            if (coinsAdd < 10)
                return $"至少{(cmdName.Contains("加") ? "加" : "扣")}10金币";

            if (cmdName.Contains("扣"))
                coinsAdd = -coinsAdd;

            var res = await AddCoinsAsync(botUin, groupId, coinsQQ, "", 1, coinsAdd, cmdName);
            if (res.Result == -1) return RetryMsg;

            return $"[@:{coinsQQ}] {cmdName}成功！\n金币：{Math.Abs(coinsAdd)}，累计：{res.CoinsValue}";
        }

        public async Task<string> ExchangeCoinsAsync(long botUin, long groupId, string groupName, long qq, string name, int coins_type, string cmdName, string cmdPara, long minus_credit, long coins_oper, long coins_qq)
        {
            if (minus_credit <= 0 || coins_oper <= 0) return RetryMsg;

            long credit_value = await _userRepo.GetCreditAsync(botUin, groupId, qq);
            if (credit_value < minus_credit)
                return $"您的积分{credit_value}不足{minus_credit}";

            using var wrapper = await _groupMemberRepo.BeginTransactionAsync();
            try
            {
                var addRes = await _userRepo.AddCreditAsync(botUin, groupId, groupName, qq, name, -minus_credit, cmdName, wrapper.Transaction);
                if (addRes.Result == -1)
                {
                    await wrapper.RollbackAsync();
                    return RetryMsg;
                }

                var coinsRes = await AddCoinsAsync(botUin, groupId, coins_qq, "", coins_type, coins_oper, cmdName, wrapper.Transaction);
                if (coinsRes.Result == -1)
                {
                    await wrapper.RollbackAsync();
                    return RetryMsg;
                }

                await wrapper.CommitAsync();

                await _userRepo.SyncCreditCacheAsync(botUin, groupId, qq, addRes.CreditValue);
                await SyncCacheFieldAsync(groupId, coins_qq, GetCoinsField(coins_type), coinsRes.CoinsValue);

                return $"✅ {cmdName}成功！\n获得：{coins_oper}{GetCoinsName(coins_type)}\n消耗：{minus_credit}积分\n剩余：{addRes.CreditValue}积分";
            }
            catch (Exception ex)
            {
                await wrapper.RollbackAsync();
                _logger.LogError(ex, "ExchangeCoins Error");
                return RetryMsg;
            }
        }

        public async Task<long> GetGroupCreditAsync(long groupId, long qq, IDbTransaction? trans = null)
        {
            return await _groupMemberRepo.GetLongAsync("GroupCredit", groupId, qq, trans);
        }

        public async Task<long> GetGroupCreditForUpdateAsync(long groupId, long qq, IDbTransaction trans)
        {
            return await _groupMemberRepo.GetForUpdateAsync("GroupCredit", groupId, qq, trans);
        }

        public async Task<long> GetSaveCreditForUpdateAsync(long groupId, long qq, IDbTransaction trans)
        {
            return await _groupMemberRepo.GetForUpdateAsync("SaveCredit", groupId, qq, trans);
        }

        public async Task<long> GetCoinsAsync(int coinsType, long groupId, long qq, IDbTransaction? trans = null)
        {
            return await _groupMemberRepo.GetCoinsAsync(coinsType, groupId, qq, trans);
        }

        public async Task<long> GetCoinsForUpdateAsync(int coinsType, long groupId, long qq, IDbTransaction trans)
        {
            return await _groupMemberRepo.GetCoinsForUpdateAsync(coinsType, groupId, qq, trans);
        }

        public async Task<long> GetGoldCoinsAsync(long groupId, long qq) => await GetCoinsAsync(1, groupId, qq);
        public async Task<long> GetPurpleCoinsAsync(long groupId, long qq) => await GetCoinsAsync(3, groupId, qq);
        public async Task<long> GetBlackCoinsAsync(long groupId, long qq) => await GetCoinsAsync(2, groupId, qq);
        public async Task<long> GetGameCoinsAsync(long groupId, long qq) => await GetCoinsAsync(4, groupId, qq);

        public async Task<(int Result, long CoinsValue, int LogId)> AddCoinsAsync(long botUin, long groupId, long qq, string name, int coinsType, long coinsAdd, string coinsInfo, IDbTransaction? trans = null)
        {
            try
            {
                if (!await _groupMemberRepo.ExistsAsync(groupId, qq, trans))
                {
                    await _groupMemberRepo.AppendAsync(groupId, qq, name, trans: trans);
                }

                long coinsValue = await _groupMemberRepo.GetCoinsAsync(coinsType, groupId, qq, trans);
                await _groupMemberRepo.IncrementValueAsync(GetCoinsField(coinsType), coinsAdd, groupId, qq, trans);

                int logId = await _coinsLogRepo.AddLogAsync(botUin, groupId, "", qq, name, coinsType, coinsAdd, coinsValue, coinsInfo, trans);

                return (1, coinsValue + coinsAdd, logId);
            }
            catch (Exception ex)
            {
                _logger.LogError(ex, "AddCoins Error");
                return (-1, 0, 0);
            }
        }

        public async Task<(int Result, long CoinsValue, int LogId)> AddCoinsTransAsync(long botUin, long groupId, long qq, string name, int coinsType, long coinsAdd, string coinsInfo, IDbTransaction? trans = null)
        {
            using var wrapper = await _groupMemberRepo.BeginTransactionAsync(trans);
            try
            {
                var res = await AddCoinsAsync(botUin, groupId, qq, name, coinsType, coinsAdd, coinsInfo, wrapper.Transaction);
                if (res.Result == 1) await wrapper.CommitAsync();
                else await wrapper.RollbackAsync();
                return res;
            }
            catch (Exception ex)
            {
                await wrapper.RollbackAsync();
                _logger.LogError(ex, "AddCoinsTrans Error");
                return (-1, 0, 0);
            }
        }

        public async Task<(int Result, long SenderCoins, long ReceiverCoins)> TransferCoinsAsync(long botUin, long groupId, long qq, string name, long qqTo, string nameTo, int coinsType, long coinsMinus, long coinsAdd, string transferInfo)
        {
            using var wrapper = await _groupMemberRepo.BeginTransactionAsync();
            try
            {
                var senderRes = await AddCoinsAsync(botUin, groupId, qq, name, coinsType, -coinsMinus, transferInfo, wrapper.Transaction);
                if (senderRes.Result == -1)
                {
                    await wrapper.RollbackAsync();
                    return (-1, 0, 0);
                }

                var receiverRes = await AddCoinsAsync(botUin, groupId, qqTo, nameTo, coinsType, coinsAdd, transferInfo, wrapper.Transaction);
                if (receiverRes.Result == -1)
                {
                    await wrapper.RollbackAsync();
                    return (-1, 0, 0);
                }

                await wrapper.CommitAsync();
                return (1, senderRes.CoinsValue, receiverRes.CoinsValue);
            }
            catch (Exception ex)
            {
                _logger.LogError(ex, "TransferCoins Error");
                await wrapper.RollbackAsync();
                return (-1, 0, 0);
            }
        }

        public async Task<int> AddCreditAsync(long groupId, long userId, long creditAdd, IDbTransaction? trans = null)
        {
            if (await _groupMemberRepo.ExistsAsync(groupId, userId, trans))
            {
                return await _groupMemberRepo.IncrementValueAsync("GroupCredit", creditAdd, groupId, userId, trans);
            }
            else
            {
                return (int)await _groupMemberRepo.AppendAsync(groupId, userId, "", groupCredit: creditAdd, trans: trans);
            }
        }

        public async Task<int> AppendAsync(long groupId, long userId, string name, string displayName = "", long groupCredit = 0, string confirmCode = "", IDbTransaction? trans = null)
        {
            return await _groupMemberRepo.AppendAsync(groupId, userId, name, displayName, groupCredit, confirmCode, trans);
        }

        public async Task<string> GetShangFenAsync(long botUin, long groupId, string groupName, long userId, string cmdName, string cmdPara)
        {
            if (!await _groupRepo.IsOwnerAsync(groupId, userId) && await _botRepo.GetRobotAdminAsync(botUin) != userId)
            {
                if (!await _userRepo.GetIsSuperAsync(userId))
                    return OwnerOnlyMsg;
            }

            if (await _incomeRepo.GetTotalAsync(userId) < 400)
                return "您无权使用此功能，请联系客服";

            if (await _botRepo.GetIsCreditAsync(botUin) || await _groupRepo.GetIsCreditAsync(groupId))
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
                    return $"格式：{cmdName} + QQ + 数量\n例如：{cmdName} 123456 5000";

                long creditAdd = 0;

                foreach (Match match in cmdPara.Matches(regexShangFen))
                {
                    creditQQ = match.Groups["UserId"].Value.AsLong();
                    creditAdd = match.Groups["credit"].Value.AsLong();
                }

                if (creditAdd < 10)
                    return $"至少{(cmdName == "上分" ? "上" : "下")}10分";

                var creditValue = await _userRepo.GetCreditAsync(botUin, groupId, creditQQ);

                if (cmdName == "下分")
                {
                    if (creditValue < creditAdd)
                        return $"对方只有{creditValue}分";
                    creditAdd = -creditAdd;
                }

                using var wrapper = await _groupMemberRepo.BeginTransactionAsync();
                try
                {
                    var res = await _userRepo.AddCreditAsync(botUin, groupId, groupName, creditQQ, "", creditAdd, cmdName, wrapper.Transaction);
                    if (res.Result == -1)
                    {
                        await wrapper.RollbackAsync();
                        return RetryMsg;
                    }

                    await wrapper.CommitAsync();
                    await _userRepo.SyncCreditCacheAsync(botUin, groupId, creditQQ, res.CreditValue);

                    return $"[@:{creditQQ}] {cmdName}成功！\n积分：{creditAdd}，累计：{res.CreditValue}";
                }
                catch (Exception ex)
                {
                    await wrapper.RollbackAsync();
                    _logger.LogError(ex, "GetShangFen Error");
                    return RetryMsg;
                }
            }
            else
                return $"此群未开通本群积分，不能上下分";
        }

        public async Task SyncCacheFieldAsync(long groupId, long qq, string field, object value)
        {
            await _groupMemberRepo.SyncCacheFieldAsync(groupId, qq, field, value);
        }

        public async Task<int> GetSignTimesAsync(long groupId, long userId)
        {
            return await _groupMemberRepo.GetIntAsync("SignTimes", groupId, userId);
        }

        public async Task<string> GetSignListAsync(long groupId, int topN = 10)
        {
            return await _groupMemberRepo.GetSignListAsync(groupId, topN);
        }

        public async Task<string> WithdrawCreditAsync(long botUin, long groupId, string groupName, long userId, string name, long withdrawAmount)
        {
            if (withdrawAmount < 100) return "提现金额不能低于100";

            using var wrapper = await _groupMemberRepo.BeginTransactionAsync();
            try
            {
                // 1. 获取积分并加锁
                long creditValue = await _userRepo.GetCreditForUpdateAsync(botUin, groupId, userId, wrapper.Transaction);
                if (creditValue < withdrawAmount)
                {
                    await wrapper.RollbackAsync();
                    return $"您的积分{creditValue}不足{withdrawAmount}";
                }

                // 2. 扣除积分
                var res = await _userRepo.AddCreditAsync(botUin, groupId, groupName, userId, name, -withdrawAmount, "积分提现", wrapper.Transaction);
                if (!res.Success) throw new Exception("扣除积分失败");

                // 3. 增加余额
                await _userRepo.IncrementValueAsync("Balance", (decimal)withdrawAmount / 100, userId, wrapper.Transaction);

                await wrapper.CommitAsync();
                
                // 4. 同步缓存
                await _userRepo.SyncCreditCacheAsync(botUin, groupId, userId, res.CreditValue);
                await _userRepo.SyncCacheFieldAsync(userId, "Balance", await _userRepo.GetBalanceAsync(userId));

                return $"✅ 提现成功！\n提现积分：{withdrawAmount}\n到账余额：{(decimal)withdrawAmount / 100:N2}\n剩余积分：{res.CreditValue}";
            }
            catch (Exception ex)
            {
                await wrapper.RollbackAsync();
                _logger.LogError(ex, "WithdrawCredit Error");
                return "系统繁忙，请稍后再试";
            }
        }

        public async Task<string> GetCoinsRankingAsync(long groupId, long userId)
        {
            return (await _groupMemberRepo.GetCoinsRankingAsync(groupId, userId)).ToString();
        }

        public async Task<string> GetCoinsRankingAllAsync(long userId)
        {
            return (await _groupMemberRepo.GetCoinsRankingAllAsync(userId)).ToString();
        }

        public async Task<int> GetSignDateDiffAsync(long groupId, long userId)
        {
            return await _groupMemberRepo.GetSignDateDiffAsync(groupId, userId);
        }

        public async Task<bool> UpdateSignInfoAsync(long groupId, long userId, int signTimes, int signLevel, IDbTransaction? trans = null)
        {
            return await _groupMemberRepo.UpdateSignInfoAsync(groupId, userId, signTimes, signLevel, trans);
        }

        private string GetCoinsField(int coinsType)
        {
            return coinsType switch
            {
                0 => "GroupCredit",
                1 => "GoldCoins",
                2 => "BlackCoins",
                3 => "PurpleCoins",
                4 => "GameCoins",
                _ => throw new ArgumentException("Invalid coins type")
            };
        }

        private string GetCoinsName(int coinsType)
        {
            return coinsType switch
            {
                0 => "群积分",
                1 => "金币",
                2 => "黑币",
                3 => "紫币",
                4 => "游戏币",
                _ => "未知货币"
            };
        }
    }
}
