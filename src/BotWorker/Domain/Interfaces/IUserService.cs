using System;
using System.Data;
using System.Threading.Tasks;

namespace BotWorker.Domain.Interfaces
{
    public interface IUserService
    {
        // Credit related
        Task<long> GetCreditAsync(long botUin, long groupId, long qq, IDbTransaction? trans = null);
        Task<long> GetCreditForUpdateAsync(long botUin, long groupId, long qq, IDbTransaction? trans = null);
        Task<long> GetSaveCreditAsync(long botUin, long groupId, long qq, IDbTransaction? trans = null);
        Task<int> AppendUserAsync(long botUin, long groupId, long qq, string name, long ownerId, IDbTransaction? trans = null);
        Task<(int Result, long CreditValue, int LogId)> AddCreditAsync(long botUin, long groupId, string groupName, long qq, string name, long creditAdd, string creditInfo, IDbTransaction? trans = null);
        Task<(int Result, long CreditValue, int LogId)> AddCreditTransAsync(long botUin, long groupId, string groupName, long qq, string name, long creditAdd, string creditInfo, IDbTransaction? trans = null);
        Task<(int Result, long SaveCreditValue, int LogId)> AddSaveCreditAsync(long botUin, long groupId, string groupName, long qq, string name, long creditAdd, string creditInfo, IDbTransaction? trans = null);
        Task<(int Result, long SaveCreditValue, int LogId)> AddSaveCreditTransAsync(long botUin, long groupId, string groupName, long qq, string name, long creditAdd, string creditInfo, IDbTransaction? trans = null);
        Task<(int Result, long CreditValue, long SaveCreditValue, int LogId)> SaveCreditAsync(long botUin, long groupId, string groupName, long qq, string name, long creditOper, IDbTransaction? trans = null);
        Task<(int Result, long SenderCredit, long ReceiverCredit)> TransferCreditAsync(
            long botUin, long groupId, string groupName, 
            long senderId, string senderName, 
            long receiverId, string receiverName, 
            long creditMinus, long creditAdd, 
            string transferInfo, IDbTransaction? trans = null);
        Task SyncCreditCacheAsync(long botUin, long groupId, long qq, long newValue);
        Task SyncSaveCreditCacheAsync(long botUin, long groupId, long qq, long newValue);

        // Balance related
        Task<decimal> GetBalanceAsync(long qq, IDbTransaction? trans = null);
        Task<decimal> GetBalanceForUpdateAsync(long qq, IDbTransaction? trans = null);
        Task<decimal> GetFreezeBalanceAsync(long qq, IDbTransaction? trans = null);
        Task<decimal> GetFreezeBalanceForUpdateAsync(long qq, IDbTransaction? trans = null);
        Task<(int Result, decimal BalanceValue)> AddBalanceAsync(long botUin, long groupId, string groupName, long qq, string name, decimal balanceAdd, string balanceInfo, IDbTransaction? trans = null);
        Task<(int Result, decimal BalanceValue)> AddBalanceTransAsync(long botUin, long groupId, string groupName, long qq, string name, decimal balanceAdd, string balanceInfo);
        Task<(int Result, decimal BalanceValue, decimal FreezeValue)> FreezeBalanceAsync(long botUin, long groupId, string groupName, long qq, string name, decimal balanceFreeze);
        Task<(int Result, decimal BalanceValue, decimal FreezeValue)> UnfreezeBalanceAsync(long botUin, long groupId, string groupName, long qq, string name, decimal balanceUnfreeze);
        Task<(int Result, decimal SenderBalance, decimal ReceiverBalance)> TransferBalanceAsync(long botUin, long groupId, string groupName, long qq, string name, long qqTo, string nameTo, decimal balanceMinus, decimal balanceAdd);
        Task<string> GetBalanceListAsync(long groupId, long qq);
        Task<string> GetMyBalanceRankAsync(long groupId, long qq);
        Task SyncBalanceCacheAsync(long qq, decimal newValue);
        Task SyncBalanceFreezeCacheAsync(long qq, decimal newValue);

        // Tokens related
        Task<long> GetTokensAsync(long qq);
        Task<long> GetTokensForUpdateAsync(long qq, IDbTransaction? trans = null);
        Task<(int Result, long TokensValue, int LogId)> AddTokensAsync(long botUin, long groupId, string groupName, long qq, string name, long tokensAdd, string tokensInfo, IDbTransaction? trans = null);
        Task<(int Result, long TokensValue, int LogId)> AddTokensTransAsync(long botUin, long groupId, string groupName, long qq, string name, long tokensAdd, string tokensInfo);
        Task<string> GetTokensListAsync(long groupId, long top);
        Task<long> GetTokensRankingAsync(long groupId, long qq);
        Task<long> GetDayTokensGroupAsync(long groupId, long userId);
        Task<long> GetDayTokensAsync(long userId);
        Task SyncTokensCacheAsync(long qq, long newValue);

        // State and General fields
        Task<int> SetStateAsync(int state, long qq);
        Task<string> GetStateResAsync(int state);
        Task<int> SetUserFieldAsync(string field, object value, long qq);
        Task<Guid> GetGuidAsync(long qq);
        Task<bool> GetIsSuperAsync(long qq);
        Task<string> GetHeadCQAsync(long user, int size = 100);
        Task<string> GetHeadAsync(long user, int size = 100);

        // UI/Response related
        Task<string> GetTransferBalanceResAsync(long botUin, long groupId, string groupName, long qq, string name, string cmdPara);
        Task<string> GetBuyCreditResAsync(BotMessage botMsg, long botUin, long groupId, string groupName, long qq, string name, string cmdPara);
        Task<string> GetCreditTypeAsync(long botUin, long groupId, long qq);
        Task<long> GetCreditRankingAsync(long botUin, long groupId, long qq);
        Task<long> GetCreditRankingAllAsync(long botUin, long qq);
        Task<long> GetTotalCreditAsync(long botUin, long groupId, long qq);
        Task<long> GetFreezeCreditAsync(long qq);

        // Guild/OpenID related
        Task<long> GetUserIdAsync(long botUin, string userOpenid, string groupOpenid);
        Task<string> GetUserOpenidAsync(long selfId, long user);

        // Purchase related
        Task<int> BuyCreditAsync(long botUin, long groupId, string groupName, long buyerQQ, string buyerName, decimal payMoney, long creditAdd, string payMethod, string trade, string memo, int insertBy);
        Task<int> BuyBalanceAsync(long botUin, long groupId, string groupName, long buyerQQ, string buyerName, decimal payMoney, decimal balanceAdd, string payMethod, string trade, string memo, int insertBy);
        Task<int> BuyTokensAsync(long botUin, long groupId, string groupName, long qqBuyer, string buyerName, decimal payMoney, long tokensAdd, string payMethod, string trade, string memo, int insertBy);
        Task<string> GetBuyCreditAdminResAsync(long botUin, long groupId, string groupName, long qq, string msgId, long buyerQQ, decimal incomeMoney, string payMethod, bool isPublic = false);

        // Coins related (moved from CoinsMessage/UserInfo)
        Task<string> GetCoinsListAllAsync(long qq, long top = 10);
        Task<string> GetCoinsListAsync(long groupId, long userId, long top = 10);
        Task<long> GetCoinsRankingAsync(long groupId, long qq);
        Task<long> GetCoinsRankingAllAsync(long qq);
        Task<bool> StartWith285or300Async(long userId);
    }
}
