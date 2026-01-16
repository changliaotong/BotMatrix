using System;
using System.Threading.Tasks;
using System.Data;
using BotWorker.Domain.Entities;

namespace BotWorker.Domain.Repositories
{
    public interface IUserRepository : IBaseRepository<UserInfo>
    {
        Task<UserInfo?> GetByOpenIdAsync(string openId, long botUin);
        Task<UserInfo?> GetBySz84UidAsync(int sz84Uid);
        Task<long> AddAsync(UserInfo user);
        Task<bool> UpdateAsync(UserInfo user);
        
        Task<long> GetCreditAsync(long botUin, long groupId, long qq, IDbTransaction? trans = null);
        Task<(bool Success, long CreditValue)> AddCreditAsync(long botUin, long groupId, string groupName, long qq, string name, long amount, string reason, IDbTransaction? trans = null);
        Task<long> GetCreditForUpdateAsync(long botUin, long groupId, long qq, IDbTransaction? trans = null);
        
        Task<long> GetTokensAsync(long qq);
        Task<long> GetTokensForUpdateAsync(long qq, IDbTransaction? trans = null);
        Task<bool> AddTokensAsync(long qq, long amount, IDbTransaction? trans = null);
        Task<long> GetDayTokensGroupAsync(long groupId, long userId);
        Task<long> GetDayTokensAsync(long userId);
        Task<string> GetTokensListAsync(long groupId, int top);
        Task<long> GetTokensRankingAsync(long groupId, long qq);
        
        Task<bool> GetIsBlackAsync(long qq);
        Task<bool> GetIsFreezeAsync(long qq);
        Task<bool> GetIsShutupAsync(long qq);
        Task<bool> GetIsSuperAsync(long qq);
        
        Task<bool> UpdateCszGameAsync(long qq, int cszRes, long cszCredit, int cszTimes);
        
        Task<long> GetCoinsAsync(long userId);
        
        Task<long> GetBotUinByOpenidAsync(string userOpenid);
        Task<long> GetTargetUserIdAsync(string userOpenid);
        Task<long> GetMaxIdInRangeAsync(long min, long max);
        Task<string> GetUserOpenidAsync(long botUin, long userId);

        Task<decimal> GetBalanceAsync(long qq, IDbTransaction? trans = null);
        Task<decimal> GetBalanceForUpdateAsync(long qq, IDbTransaction? trans = null);
        Task<bool> AddBalanceAsync(long qq, decimal amount, IDbTransaction? trans = null);
        Task<decimal> GetFreezeBalanceAsync(long qq, IDbTransaction? trans = null);
        Task<decimal> GetFreezeBalanceForUpdateAsync(long qq, IDbTransaction? trans = null);
        Task<bool> FreezeBalanceAsync(long qq, decimal amount, IDbTransaction? trans = null);
        
        Task<string> GetBalanceListAsync(long groupId, long qq);
        Task<string> GetRankAsync(long groupId, long qq);
        
        Task SyncCacheFieldAsync(long userId, string field, object value);
        Task SyncCreditCacheAsync(long botUin, long groupId, long qq, long newValue);
        
        Task<int> AppendAsync(long botUin, long groupId, long qq, string name, long ownerId, IDbTransaction? trans = null);
    }
}
