using System;
using System.Data;
using System.Threading.Tasks;
using BotWorker.Domain.Interfaces;
using BotWorker.Domain.Repositories;
using BotWorker.Domain.Entities;
using Microsoft.Extensions.Logging;

namespace BotWorker.Application.Services
{
    public class UserService : IUserService
    {
        private readonly IUserRepository _userRepo;
        private readonly IGroupRepository _groupRepo;
        private readonly IBotRepository _botRepo;
        private readonly IGroupMemberRepository _groupMemberRepo;
        private readonly IBalanceLogRepository _balanceLogRepo;
        private readonly ITokensLogRepository _tokensLogRepo;
        private readonly ICreditLogRepository _creditLogRepo;
        private readonly IIncomeRepository _incomeRepo;
        private readonly IPartnerRepository _partnerRepo;
        private readonly ILogger<UserService> _logger;

        public UserService(
            IUserRepository userRepo,
            IGroupRepository groupRepo,
            IBotRepository botRepo,
            IGroupMemberRepository groupMemberRepo,
            IBalanceLogRepository balanceLogRepo,
            ITokensLogRepository tokensLogRepo,
            ICreditLogRepository creditLogRepo,
            IIncomeRepository incomeRepo,
            IPartnerRepository partnerRepo,
            ILogger<UserService> logger)
        {
            _userRepo = userRepo;
            _groupRepo = groupRepo;
            _botRepo = botRepo;
            _groupMemberRepo = groupMemberRepo;
            _balanceLogRepo = balanceLogRepo;
            _tokensLogRepo = tokensLogRepo;
            _creditLogRepo = creditLogRepo;
            _incomeRepo = incomeRepo;
            _partnerRepo = partnerRepo;
            _logger = logger;
        }

        #region Credit Methods

        public async Task<long> GetCreditAsync(long botUin, long groupId, long qq, IDbTransaction? trans = null)
        {
            return await _userRepo.GetCreditAsync(botUin, groupId, qq, trans);
        }

        public async Task<long> GetCreditForUpdateAsync(long botUin, long groupId, long qq, IDbTransaction? trans = null)
        {
            return await _userRepo.GetCreditForUpdateAsync(botUin, groupId, qq, trans);
        }

        public async Task<long> GetSaveCreditAsync(long botUin, long groupId, long qq, IDbTransaction? trans = null)
        {
            return await _userRepo.GetSaveCreditAsync(botUin, groupId, qq, trans);
        }

        public async Task<int> AppendUserAsync(long botUin, long groupId, long qq, string name, long ownerId, IDbTransaction? trans = null)
        {
            return await _userRepo.AppendAsync(botUin, groupId, qq, name, ownerId, trans);
        }

        public async Task<(int Result, long CreditValue, int LogId)> AddCreditAsync(long botUin, long groupId, string groupName, long qq, string name, long creditAdd, string creditInfo, IDbTransaction? trans = null)
        {
            try
            {
                long ownerId = await _groupRepo.GetGroupOwnerAsync(groupId, 0, trans);
                await _userRepo.AppendAsync(botUin, groupId, qq, name, ownerId, trans);

                var creditValue = await _userRepo.GetCreditForUpdateAsync(botUin, groupId, qq, trans);
                await _userRepo.IncrementValueAsync("credit", creditAdd, qq, trans);

                int logId = await CreditLog.AddLogAsync(botUin, groupId, groupName, qq, name, creditAdd, creditValue, creditInfo, trans);

                long newValue = creditValue + creditAdd;
                return (0, newValue, logId);
            }
            catch (Exception ex)
            {
                _logger.LogError(ex, "[AddCredit Error] {Message}", ex.Message);
                if (trans != null) throw; 
                return (-1, 0, 0);
            }
        }

        public async Task<(int Result, long CreditValue, int LogId)> AddCreditTransAsync(long botUin, long groupId, string groupName, long qq, string name, long creditAdd, string creditInfo, IDbTransaction? trans = null)
        {
            using var wrapper = await _userRepo.BeginTransactionAsync(trans);
            try
            {
                var res = await AddCreditAsync(botUin, groupId, groupName, qq, name, creditAdd, creditInfo, wrapper.Transaction);
                await wrapper.CommitAsync();

                if (trans == null)
                {
                    await SyncCreditCacheAsync(botUin, groupId, qq, res.CreditValue);
                }

                return res;
            }
            catch (Exception ex)
            {
                _logger.LogError(ex, "[AddCreditTrans Error] {Message}", ex.Message);
                await wrapper.RollbackAsync();
                if (trans != null) throw;
                return (-1, 0, 0);
            }
        }

        public async Task<(int Result, long SaveCreditValue, int LogId)> AddSaveCreditAsync(long botUin, long groupId, string groupName, long qq, string name, long creditAdd, string creditInfo, IDbTransaction? trans = null)
        {
            try
            {
                long ownerId = await _groupRepo.GetGroupOwnerAsync(groupId, 0, trans);
                await _userRepo.AppendAsync(botUin, groupId, qq, name, ownerId, trans);

                var creditValue = await _userRepo.GetSaveCreditForUpdateAsync(botUin, groupId, qq, trans);
                await _userRepo.AddSaveCreditAsync(botUin, groupId, qq, creditAdd, trans);

                int logId = await CreditLog.AddLogAsync(botUin, groupId, groupName, qq, name, creditAdd, creditValue, $"[存款]{creditInfo}", trans);

                long newValue = creditValue + creditAdd;
                return (0, newValue, logId);
            }
            catch (Exception ex)
            {
                _logger.LogError(ex, "[AddSaveCredit Error] {Message}", ex.Message);
                if (trans != null) throw;
                return (-1, 0, 0);
            }
        }

        public async Task<(int Result, long SaveCreditValue, int LogId)> AddSaveCreditTransAsync(long botUin, long groupId, string groupName, long qq, string name, long creditAdd, string creditInfo, IDbTransaction? trans = null)
        {
            using var wrapper = await _userRepo.BeginTransactionAsync(trans);
            try
            {
                var res = await AddSaveCreditAsync(botUin, groupId, groupName, qq, name, creditAdd, creditInfo, wrapper.Transaction);
                await wrapper.CommitAsync();

                if (trans == null)
                {
                    await SyncSaveCreditCacheAsync(botUin, groupId, qq, res.SaveCreditValue);
                }

                return res;
            }
            catch (Exception ex)
            {
                _logger.LogError(ex, "[AddSaveCreditTrans Error] {Message}", ex.Message);
                await wrapper.RollbackAsync();
                if (trans != null) throw;
                return (-1, 0, 0);
            }
        }

        public async Task<(int Result, long CreditValue, long SaveCreditValue, int LogId)> SaveCreditAsync(long botUin, long groupId, string groupName, long qq, string name, long creditOper, IDbTransaction? trans = null)
        {
            using var wrapper = await _userRepo.BeginTransactionAsync(trans);
            try
            {
                long ownerId = await _groupRepo.GetGroupOwnerAsync(groupId, 0, wrapper.Transaction);
                await _userRepo.AppendAsync(botUin, groupId, qq, name, ownerId, wrapper.Transaction);

                long creditValue = await _userRepo.GetCreditForUpdateAsync(botUin, groupId, qq, wrapper.Transaction);
                long creditSave = await _userRepo.GetSaveCreditForUpdateAsync(botUin, groupId, qq, wrapper.Transaction);

                string cmdName = creditOper > 0 ? "存分" : "取分";
                long absOper = Math.Abs(creditOper);

                if (creditOper > 0)
                {
                    if (creditValue < absOper)
                        return (-2, creditValue, creditSave, 0);
                }
                else
                {
                    if (creditSave < absOper)
                        return (-3, creditValue, creditSave, 0);
                }

                await _userRepo.IncrementValueAsync("credit", -creditOper, qq, wrapper.Transaction);
                await _userRepo.IncrementValueAsync("save_credit", creditOper, qq, wrapper.Transaction);

                int logId = await CreditLog.AddLogAsync(botUin, groupId, groupName, qq, name, -creditOper, creditValue, cmdName, wrapper.Transaction);

                await wrapper.CommitAsync();

                long newCredit = creditValue - creditOper;
                long newSave = creditSave + creditOper;

                if (trans == null)
                {
                    await SyncCreditCacheAsync(botUin, groupId, qq, newCredit);
                    await SyncSaveCreditCacheAsync(botUin, groupId, qq, newSave);
                }

                return (0, newCredit, newSave, logId);
            }
            catch (Exception ex)
            {
                _logger.LogError(ex, "[SaveCredit Error] {Message}", ex.Message);
                await wrapper.RollbackAsync();
                if (trans != null) throw;
                return (-1, 0, 0, 0);
            }
        }

        public async Task<(int Result, long SenderCredit, long ReceiverCredit)> TransferCreditAsync(
            long botUin, long groupId, string groupName, 
            long senderId, string senderName, 
            long receiverId, string receiverName, 
            long creditMinus, long creditAdd, 
            string transferInfo, IDbTransaction? trans = null)
        {
            using var wrapper = await _userRepo.BeginTransactionAsync(trans);
            try
            {
                long firstId = senderId < receiverId ? senderId : receiverId;
                long secondId = senderId < receiverId ? receiverId : senderId;
                string firstName = firstId == senderId ? senderName : receiverName;
                string secondName = secondId == senderId ? senderName : receiverName;

                long ownerId = await _groupRepo.GetGroupOwnerAsync(groupId, 0, wrapper.Transaction);
                await _userRepo.AppendAsync(botUin, groupId, firstId, firstName, ownerId, wrapper.Transaction);
                await _userRepo.AppendAsync(botUin, groupId, secondId, secondName, ownerId, wrapper.Transaction);

                if (firstId == senderId)
                {
                    await _userRepo.GetCreditForUpdateAsync(botUin, groupId, senderId, wrapper.Transaction);
                    await _userRepo.GetCreditForUpdateAsync(botUin, groupId, receiverId, wrapper.Transaction);
                }
                else
                {
                    await _userRepo.GetCreditForUpdateAsync(botUin, groupId, receiverId, wrapper.Transaction);
                    await _userRepo.GetCreditForUpdateAsync(botUin, groupId, senderId, wrapper.Transaction);
                }

                long senderCredit = await _userRepo.GetCreditAsync(botUin, groupId, senderId, wrapper.Transaction);
                if (senderCredit < creditMinus)
                    return (-1, senderCredit, 0);

                var res1 = await AddCreditAsync(botUin, groupId, groupName, senderId, senderName, -creditMinus, $"{transferInfo}扣分：{receiverId}", wrapper.Transaction);
                var res2 = await AddCreditAsync(botUin, groupId, groupName, receiverId, receiverName, creditAdd, $"{transferInfo}加分：{senderId}", wrapper.Transaction);

                await wrapper.CommitAsync();

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
                _logger.LogError(ex, "[TransferCredit Error] {Message}", ex.Message);
                if (trans != null) throw;
                return (-1, 0, 0);
            }
        }

        public async Task SyncCreditCacheAsync(long botUin, long groupId, long qq, long newValue)
        {
            if (await _groupRepo.GetIsCreditAsync(groupId))
                await _groupMemberRepo.IncrementValueAsync("group_credit", 0, groupId, qq); // Sync logic might be different, let's check original code
            else if (await _botRepo.GetIsCreditAsync(botUin))
                await _friendRepo.UpdateAsync(new Friend { BotUin = botUin, UserId = qq, Credit = newValue }); // Friend update logic
            else
                await _userRepo.SyncCacheFieldAsync(qq, "credit", newValue);
        }

        public async Task SyncSaveCreditCacheAsync(long botUin, long groupId, long qq, long newValue)
        {
            if (await _groupRepo.GetIsCreditAsync(groupId))
                await GroupMember.SyncCacheFieldAsync(groupId, qq, "save_credit", newValue);
            else if (await _botRepo.GetIsCreditAsync(botUin))
                await Friend.SyncCacheFieldAsync(botUin, qq, "save_credit", newValue);
            else
                await _userRepo.SyncCacheFieldAsync(qq, "save_credit", newValue);
        }

        #endregion

        #region Balance Methods

        public async Task<decimal> GetBalanceAsync(long qq, IDbTransaction? trans = null)
        {
            return await _userRepo.GetBalanceAsync(qq, trans);
        }

        public async Task<decimal> GetBalanceForUpdateAsync(long qq, IDbTransaction? trans = null)
        {
            return await _userRepo.GetBalanceForUpdateAsync(qq, trans);
        }

        public async Task<decimal> GetFreezeBalanceAsync(long qq, IDbTransaction? trans = null)
        {
            return await _userRepo.GetFreezeBalanceAsync(qq, trans);
        }

        public async Task<decimal> GetFreezeBalanceForUpdateAsync(long qq, IDbTransaction? trans = null)
        {
            return await _userRepo.GetFreezeBalanceForUpdateAsync(qq, trans);
        }

        public async Task<(int Result, decimal BalanceValue)> AddBalanceAsync(long botUin, long groupId, string groupName, long qq, string name, decimal balanceAdd, string balanceInfo, IDbTransaction? trans = null)
        {
            try
            {
                long ownerId = await _groupRepo.GetGroupOwnerAsync(groupId, 0, trans);
                await _userRepo.AppendAsync(botUin, groupId, qq, name, ownerId, trans);

                var balanceValue = await _userRepo.GetBalanceForUpdateAsync(qq, trans);
                var newValue = balanceValue + balanceAdd;

                await _userRepo.AddBalanceAsync(qq, balanceAdd, trans);
                await _balanceLogRepo.AddLogAsync(botUin, groupId, groupName, qq, name, balanceAdd, newValue, balanceInfo, trans);

                return (0, newValue);
            }
            catch (Exception ex)
            {
                _logger.LogError(ex, "[AddBalance Error] {Message}", ex.Message);
                if (trans != null) throw;
                return (-1, await _userRepo.GetBalanceAsync(qq));
            }
        }

        public async Task<(int Result, decimal BalanceValue)> AddBalanceTransAsync(long botUin, long groupId, string groupName, long qq, string name, decimal balanceAdd, string balanceInfo)
        {
            using var wrapper = await _userRepo.BeginTransactionAsync();
            try
            {
                var res = await AddBalanceAsync(botUin, groupId, groupName, qq, name, balanceAdd, balanceInfo, wrapper.Transaction);
                await wrapper.CommitAsync();

                await _userRepo.SyncCacheFieldAsync(qq, "Balance", res.BalanceValue);

                return res;
            }
            catch (Exception ex)
            {
                _logger.LogError(ex, "[AddBalanceTrans Error] {Message}", ex.Message);
                await wrapper.RollbackAsync();
                return (-1, 0);
            }
        }

        public async Task<(int Result, decimal BalanceValue, decimal FreezeValue)> FreezeBalanceAsync(long botUin, long groupId, string groupName, long qq, string name, decimal balanceFreeze)
        {
            using var wrapper = await _userRepo.BeginTransactionAsync();
            try
            {
                decimal balanceValue = await _userRepo.GetBalanceForUpdateAsync(qq, wrapper.Transaction);
                if (balanceValue < balanceFreeze)
                {
                    await wrapper.RollbackAsync();
                    return (-1, balanceValue, 0);
                }

                await _userRepo.FreezeBalanceAsync(qq, balanceFreeze, wrapper.Transaction);
                
                decimal freezeValue = await _userRepo.GetFreezeBalanceAsync(qq, wrapper.Transaction);
                decimal newBalance = balanceValue - balanceFreeze;

                await _balanceLogRepo.AddLogAsync(botUin, groupId, groupName, qq, name, -balanceFreeze, newBalance, "冻结余额", wrapper.Transaction);

                await wrapper.CommitAsync();

                await _userRepo.SyncCacheFieldAsync(qq, "Balance", newBalance);
                await _userRepo.SyncCacheFieldAsync(qq, "BalanceFreeze", freezeValue);
                return (0, newBalance, freezeValue);
            }
            catch (Exception ex)
            {
                await wrapper.RollbackAsync();
                _logger.LogError(ex, "[FreezeBalance Error] {Message}", ex.Message);
                return (-1, await _userRepo.GetBalanceAsync(qq), await _userRepo.GetFreezeBalanceAsync(qq));
            }
        }

        public async Task<(int Result, decimal BalanceValue, decimal FreezeValue)> UnfreezeBalanceAsync(long botUin, long groupId, string groupName, long qq, string name, decimal balanceUnfreeze)
        {
            using var wrapper = await _userRepo.BeginTransactionAsync();
            try
            {
                decimal freezeValue = await _userRepo.GetFreezeBalanceForUpdateAsync(qq, wrapper.Transaction);
                if (freezeValue < balanceUnfreeze)
                {
                    await wrapper.RollbackAsync();
                    return (-1, 0, freezeValue);
                }

                await _userRepo.FreezeBalanceAsync(qq, -balanceUnfreeze, wrapper.Transaction);

                decimal balanceValue = await _userRepo.GetBalanceAsync(qq, wrapper.Transaction);
                decimal newFreeze = freezeValue - balanceUnfreeze;

                await _balanceLogRepo.AddLogAsync(botUin, groupId, groupName, qq, name, balanceUnfreeze, balanceValue, "解冻余额", wrapper.Transaction);

                await wrapper.CommitAsync();

                await _userRepo.SyncCacheFieldAsync(qq, "Balance", balanceValue);
                await _userRepo.SyncCacheFieldAsync(qq, "BalanceFreeze", newFreeze);
                return (0, balanceValue, newFreeze);
            }
            catch (Exception ex)
            {
                await wrapper.RollbackAsync();
                _logger.LogError(ex, "[UnfreezeBalance Error] {Message}", ex.Message);
                return (-1, await _userRepo.GetBalanceAsync(qq), await _userRepo.GetFreezeBalanceAsync(qq));
            }
        }

        public async Task<(int Result, decimal SenderBalance, decimal ReceiverBalance)> TransferBalanceAsync(long botUin, long groupId, string groupName, long qq, string name, long qqTo, string nameTo, decimal balanceMinus, decimal balanceAdd)
        {
            using var wrapper = await _userRepo.BeginTransactionAsync();
            try
            {
                long firstId = qq < qqTo ? qq : qqTo;
                long secondId = qq < qqTo ? qqTo : qq;
                string firstName = firstId == qq ? name : nameTo;
                string secondName = secondId == qq ? name : nameTo;

                long ownerId = await _groupRepo.GetGroupOwnerAsync(groupId, 0, wrapper.Transaction);
                await _userRepo.AppendAsync(botUin, groupId, firstId, firstName, ownerId, wrapper.Transaction);
                await _userRepo.AppendAsync(botUin, groupId, secondId, secondName, ownerId, wrapper.Transaction);

                await _userRepo.GetBalanceForUpdateAsync(firstId, wrapper.Transaction);
                await _userRepo.GetBalanceForUpdateAsync(secondId, wrapper.Transaction);

                var currentBalance = await _userRepo.GetBalanceAsync(qq, wrapper.Transaction);
                if (currentBalance < balanceMinus)
                {
                    await wrapper.RollbackAsync();
                    return (-2, currentBalance, 0);
                }

                var res1 = await AddBalanceAsync(botUin, groupId, groupName, qq, name, -balanceMinus, $"转账给：{qqTo}", wrapper.Transaction);
                var res2 = await AddBalanceAsync(botUin, groupId, groupName, qqTo, nameTo, balanceAdd, $"转账来自：{qq}", wrapper.Transaction);

                if (res1.Result == -1 || res2.Result == -1)
                    throw new Exception("Transfer failed");

                await wrapper.CommitAsync();

                await _userRepo.SyncCacheFieldAsync(qq, "Balance", res1.BalanceValue);
                await _userRepo.SyncCacheFieldAsync(qqTo, "Balance", res2.BalanceValue);

                return (0, res1.BalanceValue, res2.BalanceValue);
            }
            catch (Exception ex)
            {
                await wrapper.RollbackAsync();
                _logger.LogError(ex, "[Transfer Error] {Message}", ex.Message);
                return (-1, 0, 0);
            }
        }

        public async Task<string> GetBalanceListAsync(long groupId, long qq)
        {
            return await _userRepo.GetBalanceListAsync(groupId, qq);
        }

        public async Task<string> GetMyBalanceRankAsync(long groupId, long qq)
        {
            return await _userRepo.GetRankAsync(groupId, qq);
        }

        public async Task SyncBalanceCacheAsync(long qq, decimal newValue)
        {
            await _userRepo.SyncCacheFieldAsync(qq, "Balance", newValue);
        }

        public async Task SyncBalanceFreezeCacheAsync(long qq, decimal newValue)
        {
            await _userRepo.SyncCacheFieldAsync(qq, "BalanceFreeze", newValue);
        }

        #endregion

        #region Tokens Methods

        public async Task<long> GetTokensAsync(long qq)
        {
            return await _userRepo.GetTokensAsync(qq);
        }

        public async Task<long> GetTokensForUpdateAsync(long qq, IDbTransaction? trans = null)
        {
            return await _userRepo.GetTokensForUpdateAsync(qq, trans);
        }

        public async Task<(int Result, long TokensValue, int LogId)> AddTokensAsync(long botUin, long groupId, string groupName, long qq, string name, long tokensAdd, string tokensInfo, IDbTransaction? trans = null)
        {
            try
            {
                long ownerId = await _groupRepo.GetGroupOwnerAsync(groupId, 0, trans);
                await _userRepo.AppendAsync(botUin, groupId, qq, name, ownerId, trans);

                var tokensValue = await _userRepo.GetTokensForUpdateAsync(qq, trans);

                if (tokensAdd < 0 && tokensValue < Math.Abs(tokensAdd))
                {
                    return (-2, tokensValue, 0);
                }

                int logId = await _tokensLogRepo.AddLogAsync(botUin, groupId, groupName, qq, name, tokensAdd, tokensValue, tokensInfo, trans);
                await _userRepo.AddTokensAsync(qq, tokensAdd, trans);

                return (0, tokensValue + tokensAdd, logId);
            }
            catch (Exception ex)
            {
                _logger.LogError(ex, "[AddTokens Error] {Message}", ex.Message);
                if (trans != null) throw;
                return (-1, 0, 0);
            }
        }

        public async Task<(int Result, long TokensValue, int LogId)> AddTokensTransAsync(long botUin, long groupId, string groupName, long qq, string name, long tokensAdd, string tokensInfo)
        {
            using var wrapper = await _userRepo.BeginTransactionAsync();
            try
            {
                var res = await AddTokensAsync(botUin, groupId, groupName, qq, name, tokensAdd, tokensInfo, wrapper.Transaction);
                await wrapper.CommitAsync();

                await _userRepo.SyncCacheFieldAsync(qq, "tokens", res.TokensValue);
                return res;
            }
            catch (Exception ex)
            {
                _logger.LogError(ex, "[AddTokensTrans Error] {Message}", ex.Message);
                await wrapper.RollbackAsync();
                return (-1, 0, 0);
            }
        }

        public async Task<string> GetTokensListAsync(long groupId, long top)
        {
            return await _userRepo.GetTokensListAsync(groupId, (int)top);
        }

        public async Task<long> GetTokensRankingAsync(long groupId, long qq)
        {
            return await _userRepo.GetTokensRankingAsync(groupId, qq);
        }

        public async Task<long> GetDayTokensGroupAsync(long groupId, long userId)
        {
            return await _tokensLogRepo.GetDayTokensGroupAsync(groupId, userId);
        }

        public async Task<long> GetDayTokensAsync(long userId)
        {
            return await _tokensLogRepo.GetDayTokensAsync(userId);
        }

        public async Task SyncTokensCacheAsync(long qq, long newValue)
        {
            await _userRepo.SyncCacheFieldAsync(qq, "tokens", newValue);
        }

        #endregion

        #region State and General Fields

        public async Task<int> SetStateAsync(int state, long qq)
        {
            return await _userRepo.SetValueAsync("state", state, qq);
        }

        public async Task<string> GetStateResAsync(int state)
        {
            return state switch
            {
                0 => "闲聊",
                1 => "猜数字",
                2 => "猜大小",
                3 => "三公",
                4 => "猜拳",
                _ => "闲聊"
            };
        }

        public async Task<int> SetUserFieldAsync(string field, object value, long qq)
        {
            return await _userRepo.SetValueAsync(field, value, qq);
        }

        public async Task<Guid> GetGuidAsync(long qq)
        {
            return await _userRepo.GetGuidAsync(qq);
        }

        public async Task<bool> GetIsSuperAsync(long qq)
        {
            return await _userRepo.GetIsSuperAsync(qq);
        }

        public async Task<string> GetHeadCQAsync(long user, int size = 100)
        {
            return $"[CQ:image,file={await GetHeadAsync(user, size)}]";
        }

        public async Task<string> GetHeadAsync(long user, int size = 100)
        {
            return $"https://q1.qlogo.cn/g?b=qq&nk={user}&s={size}";
        }

        #endregion

        #region UI/Response Methods

        public async Task<string> GetTransferBalanceResAsync(long botUin, long groupId, string groupName, long qq, string name, string cmdPara)
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

            var res = await TransferBalanceAsync(botUin, groupId, groupName, qq, name, qqTransfer, "", balanceTransfer, balanceTransfer);
            if (res.Result == -2)
                return $"余额{res.SenderBalance}不足{balanceTransfer}。";

            return res.Result == -1
                ? "系统繁忙，请稍后再试"
                : $"✅ 成功转出：{balanceTransfer}\n[@:{qqTransfer}] 的余额：{res.ReceiverBalance}\n你的余额：{res.SenderBalance}";
        }

        public async Task<string> GetBuyCreditResAsync(BotMessage context, long botUin, long groupId, string groupName, long qq, string name, string cmdPara)
        {
            if (!context.Group.IsCreditSystem)
                return CreditSystemClosed;

            if (cmdPara == "")
                return "格式：买分 + 金额\n例如：买分 16.88\n价格：1R=1200分\n您的余额：{余额}";

            if (!cmdPara.IsDouble())
                return "金额不正确";

            decimal balanceMinus = cmdPara.AsDecimal();
            decimal balanceValue = await GetBalanceAsync(qq);
            if (balanceMinus < 1)
                return "至少买1元";

            if (balanceMinus > balanceValue)
                return $"您的余额{balanceValue:N}不足{balanceMinus:N}";

            long creditAdd = Convert.ToInt32(balanceMinus * 1200);
            bool isPartner = await _partnerRepo.IsPartnerAsync(qq);
            if (isPartner) creditAdd *= 2;

            using var wrapper = await _userRepo.BeginTransactionAsync();
            try
            {
                // 1. 获取准确余额并锁定
                decimal balanceValueTrans = await GetBalanceForUpdateAsync(qq, wrapper.Transaction);
                if (balanceValueTrans < balanceMinus)
                {
                    await wrapper.RollbackAsync();
                    return $"您的余额{balanceValueTrans:N}不足{balanceMinus:N}";
                }
                decimal balanceNewTrans = balanceValueTrans - balanceMinus;

                // 2. 扣除余额 (含日志记录)
                var resBalance = await AddBalanceAsync(botUin, groupId, groupName, qq, name, -balanceMinus, "买分", wrapper.Transaction);
                if (resBalance.Result == -1) throw new Exception("更新余额失败");

                // 3. 增加积分 (含日志记录)
                var resCredit = await AddCreditAsync(botUin, groupId, groupName, qq, name, creditAdd, "买分", wrapper.Transaction);
                if (resCredit.Result == -1) throw new Exception("更新积分失败");

                await wrapper.CommitAsync();

                await _userRepo.SyncCacheFieldAsync(qq, "Balance", balanceNewTrans);
                await SyncCreditCacheAsync(botUin, groupId, qq, resCredit.CreditValue);

                return $"✅ 买分成功！\n积分：+{creditAdd}，累计：{resCredit.CreditValue}\n余额：-{balanceMinus:N}，累计：{balanceNewTrans:N}";
            }
            catch (Exception ex)
            {
                await wrapper.RollbackAsync();
                _logger.LogError(ex, "[GetBuyCredit Error] {Message}", ex.Message);
                return RetryMsg;
            }
        }

        public async Task<string> GetCreditTypeAsync(long botUin, long groupId, long qq)
        {
            if (await _userRepo.GetIsSuperAsync(qq)) return "超级积分";
            if (await _groupRepo.GetIsCreditAsync(groupId)) return "本群积分";
            if (await _botRepo.GetIsCreditAsync(botUin)) return "本机积分";
            return "通用积分";
        }

        public async Task<long> GetCreditRankingAsync(long botUin, long groupId, long qq)
        {
            if (await _groupRepo.GetIsCreditAsync(groupId))
                return await GroupMember.GetCreditRankingAsync(groupId, qq);
            
            if (await _botRepo.GetIsCreditAsync(botUin))
                return await Friend.GetCreditRankingAsync(botUin, qq);

            return await _userRepo.GetCreditRankingAsync(qq);
        }

        public async Task<long> GetCreditRankingAllAsync(long botUin, long qq)
        {
            return await _userRepo.GetCreditRankingAsync(qq);
        }

        public async Task<long> GetTotalCreditAsync(long botUin, long groupId, long qq)
        {
            return await GetCreditAsync(botUin, groupId, qq) + await GetSaveCreditAsync(botUin, groupId, qq);
        }

        public async Task<long> GetFreezeCreditAsync(long qq)
        {
            return await _userRepo.GetFreezeCreditAsync(qq);
        }

        #endregion

        #region Purchase Methods

        public async Task<int> BuyCreditAsync(long botUin, long groupId, string groupName, long buyerQQ, string buyerName, decimal payMoney, long creditAdd, string payMethod, string trade, string memo, int insertBy)
        {
            using var wrapper = await _userRepo.BeginTransactionAsync();
            try
            {
                // 1. 记录收入
                await _incomeRepo.AddAsync(new Income
                {
                    GroupId = groupId,
                    GoodsCount = creditAdd,
                    GoodsName = "积分",
                    UserId = buyerQQ,
                    IncomeMoney = payMoney,
                    PayMethod = payMethod,
                    IncomeTrade = trade,
                    IncomeInfo = memo,
                    InsertBy = insertBy,
                    IncomeDate = DateTime.Now
                }, wrapper.Transaction);

                // 2. 通用加积分函数 (含日志记录)
                var res = await AddCreditAsync(botUin, groupId, groupName, buyerQQ, buyerName, creditAdd, "买分", wrapper.Transaction);
                if (res.Result == -1) throw new Exception("更新积分失败");

                await wrapper.CommitAsync();

                await _userRepo.SyncCacheFieldAsync(buyerQQ, "Credit", res.CreditValue);
                return 0;
            }
            catch (Exception ex)
            {
                await wrapper.RollbackAsync();
                _logger.LogError(ex, "[BuyCredit Error] {Message}", ex.Message);
                return -1;
            }
        }

        public async Task<int> BuyBalanceAsync(long botUin, long groupId, string groupName, long buyerQQ, string buyerName, decimal payMoney, decimal balanceAdd, string payMethod, string trade, string memo, int insertBy)
        {
            using var wrapper = await _userRepo.BeginTransactionAsync();
            try
            {
                // 1. 记录收入
                await _incomeRepo.AddAsync(new Income
                {
                    GroupId = groupId,
                    GoodsCount = 1,
                    GoodsName = "余额",
                    UserId = buyerQQ,
                    IncomeMoney = payMoney,
                    PayMethod = payMethod,
                    IncomeTrade = trade,
                    IncomeInfo = memo,
                    InsertBy = insertBy,
                    IncomeDate = DateTime.Now
                }, wrapper.Transaction);

                // 2. 增加余额 (含日志记录)
                var res = await AddBalanceAsync(botUin, groupId, groupName, buyerQQ, buyerName, balanceAdd, "充值余额", wrapper.Transaction);
                if (res.Result == -1) throw new Exception("更新余额失败");

                await wrapper.CommitAsync();

                await _userRepo.SyncCacheFieldAsync(buyerQQ, "Balance", res.BalanceValue);
                return 0;
            }
            catch (Exception ex)
            {
                await wrapper.RollbackAsync();
                _logger.LogError(ex, "[BuyBalance Error] {Message}", ex.Message);
                return -1;
            }
        }

        public async Task<int> BuyTokensAsync(long botUin, long groupId, string groupName, long qqBuyer, string buyerName, decimal payMoney, long tokensAdd, string payMethod, string trade, string memo, int insertBy)
        {
            using var wrapper = await _userRepo.BeginTransactionAsync();
            try
            {
                // 1. 记录收入
                await _incomeRepo.AddAsync(new Income
                {
                    GroupId = groupId,
                    GoodsCount = tokensAdd,
                    GoodsName = "TOKENS",
                    UserId = qqBuyer,
                    IncomeMoney = payMoney,
                    PayMethod = payMethod,
                    IncomeTrade = trade,
                    IncomeInfo = memo,
                    InsertBy = insertBy,
                    IncomeDate = DateTime.Now
                }, wrapper.Transaction);

                // 2. 增加算力 (含日志记录)
                var res = await AddTokensAsync(botUin, groupId, groupName, qqBuyer, buyerName, tokensAdd, "购买算力", wrapper.Transaction);
                if (res.Result == -1) throw new Exception("更新算力失败");

                await wrapper.CommitAsync();

                await _userRepo.SyncCacheFieldAsync(qqBuyer, "Tokens", res.TokensValue);
                return 0;
            }
            catch (Exception ex)
            {
                await wrapper.RollbackAsync();
                _logger.LogError(ex, "[BuyTokens Error] {Message}", ex.Message);
                return -1;
            }
        }

        public async Task<string> GetBuyCreditAdminResAsync(long botUin, long groupId, string groupName, long qq, string msgId, long buyerQQ, decimal incomeMoney, string payMethod, bool isPublic = false)
        {
            if (!BotInfo.IsSuperAdmin(qq))
                return "您不是管理员，无权充值积分";
            payMethod = payMethod switch
            {
                "qq" => "QQ红包",
                "wx" => "微信支付",
                "zfb" => "支付宝",
                "微信" => "微信支付",
                _ => "QQ红包"
            };

            if (isPublic && await _userRepo.GetValueAsync("MsgId", qq) == msgId)
                return $"重复消息{RetryMsg}";

            long creditValue = await GetCreditAsync(botUin, groupId, buyerQQ);
            
            long creditAdd = (long)Math.Round(incomeMoney * 1200, 0);
            if (await _partnerRepo.IsPartnerAsync(buyerQQ))
            {
                if (await GetIsSuperAsync(buyerQQ))
                    creditAdd *= 2;
                else
                    creditAdd = (long)Math.Round(incomeMoney * 10000, 0);
            }

            return await BuyCreditAsync(botUin, groupId, groupName, buyerQQ, "", incomeMoney, creditAdd, payMethod, "", "", BotInfo.SystemUid) == -1
                ? RetryMsg
                : $"✅ 购买成功！\n{buyerQQ}积分：\n{creditValue}{(creditAdd > 0 ? $"+" : $"")}{creditAdd} = {await GetCreditAsync(botUin, groupId, buyerQQ)}";
        }

        #endregion

        #region Guild/OpenID Methods

        private const long MIN_USER_ID = 980000000000;
        private const long MAX_USER_ID = 990000000000;

        public async Task<long> GetUserIdAsync(long botUin, string userOpenid, string groupOpenid)
        {
            if (string.IsNullOrEmpty(userOpenid))
                return 0;

            var userId = await _userRepo.GetTargetUserIdAsync(userOpenid);
            if (userId != 0)
            {
                var bot = await _userRepo.GetBotUinByOpenidAsync(userOpenid);
                if (bot != botUin)
                    await _userRepo.SetValueAsync("bot_uin", botUin, userId);
                return userId;
            }

            userId = await GetMaxUserIdAsync();
            int i = await _userRepo.AppendAsync(botUin, 0, userId, "", 0, userOpenid, groupOpenid);
            return i == -1 ? 0 : userId;
        }

        public async Task<string> GetUserOpenidAsync(long selfId, long user)
        {
            return await _userRepo.GetUserOpenidAsync(selfId, user);
        }

        private async Task<long> GetMaxUserIdAsync()
        {
            var userId = await _userRepo.GetMaxIdInRangeAsync(MIN_USER_ID, MAX_USER_ID);
            return userId <= MIN_USER_ID ? MIN_USER_ID + 1 : userId + 1;
        }

        #endregion

        #region Coins Methods

        public async Task<string> GetCoinsListAllAsync(long qq, long top = 10)
        {
            return await _userRepo.GetCoinsListAllAsync(qq, (int)top);
        }

        public async Task<string> GetCoinsListAsync(long groupId, long userId, long top = 10)
        {
            return await _userRepo.GetCoinsListAsync(groupId, userId, (int)top);
        }

        public async Task<long> GetCoinsRankingAsync(long groupId, long qq)
        {
            return await _userRepo.GetCoinsRankingAsync(groupId, qq);
        }

        public async Task<long> GetCoinsRankingAllAsync(long qq)
        {
            return await _userRepo.GetCoinsRankingAllAsync(qq);
        }

        public Task<bool> StartWith285or300Async(long userId)
        {
            var userIdStr = userId.ToString();
            return Task.FromResult(userIdStr.StartsWith("285") || userIdStr.StartsWith("300"));
        }
        #endregion
    }
}
