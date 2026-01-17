using System.Threading.Tasks;
using System.Data;

namespace BotWorker.Domain.Interfaces
{
    public interface IGroupMemberService
    {
        Task<string> AddCoinsResAsync(long botUin, long groupId, string groupName, long qq, string name, string cmdName, string cmdPara, string cmdPara2, string cmdPara3);
        Task<string> ExchangeCoinsAsync(long botUin, long groupId, string groupName, long qq, string name, int coins_type, string cmdName, string cmdPara, long minus_credit, long coins_oper, long coins_qq);
        Task<long> GetGroupCreditAsync(long groupId, long qq, IDbTransaction? trans = null);
        Task<long> GetGroupCreditForUpdateAsync(long groupId, long qq, IDbTransaction trans);
        Task<long> GetSaveCreditForUpdateAsync(long groupId, long qq, IDbTransaction trans);
        Task<long> GetCoinsAsync(int coinsType, long groupId, long qq, IDbTransaction? trans = null);
        Task<long> GetCoinsForUpdateAsync(int coinsType, long groupId, long qq, IDbTransaction trans);
        Task<long> GetGoldCoinsAsync(long groupId, long qq);
        Task<long> GetPurpleCoinsAsync(long groupId, long qq);
        Task<long> GetBlackCoinsAsync(long groupId, long qq);
        Task<long> GetGameCoinsAsync(long groupId, long qq);
        Task<(int Result, long CoinsValue, int LogId)> AddCoinsAsync(long botUin, long groupId, long qq, string name, int coinsType, long coinsAdd, string coinsInfo, IDbTransaction? trans = null);
        Task<(int Result, long CoinsValue, int LogId)> AddCoinsTransAsync(long botUin, long groupId, long qq, string name, int coinsType, long coinsAdd, string coinsInfo, IDbTransaction? trans = null);
        Task<(int Result, long SenderCoins, long ReceiverCoins)> TransferCoinsAsync(long botUin, long groupId, long qq, string name, long qqTo, string nameTo, int coinsType, long coinsMinus, long coinsAdd, string transferInfo);
        Task<int> AddCreditAsync(long groupId, long userId, long creditAdd, IDbTransaction? trans = null);
        Task<int> AppendAsync(long groupId, long userId, string name, string displayName = "", long groupCredit = 0, string confirmCode = "", IDbTransaction? trans = null);
        Task<string> GetShangFenAsync(long botUin, long groupId, string groupName, long userId, string cmdName, string cmdPara);
        Task SyncCacheFieldAsync(long groupId, long qq, string field, object value);
        Task<int> GetSignTimesAsync(long groupId, long userId);
        Task<string> GetSignListAsync(long groupId, int topN = 10);
        Task<string> WithdrawCreditAsync(long botUin, long groupId, string groupName, long userId, string name, long withdrawAmount);
        Task<string> GetCoinsRankingAsync(long groupId, long userId);
        Task<string> GetCoinsRankingAllAsync(long userId);
        Task<int> GetSignDateDiffAsync(long groupId, long userId);
        Task<bool> UpdateSignInfoAsync(long groupId, long userId, int signTimes, int signLevel, IDbTransaction? trans = null);
    }
}
