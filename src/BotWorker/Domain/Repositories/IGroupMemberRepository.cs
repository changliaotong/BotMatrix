using System.Threading.Tasks;
using System.Data;
using BotWorker.Domain.Entities;

namespace BotWorker.Domain.Repositories
{
    public interface IGroupMemberRepository : IBaseRepository<GroupMember>
    {
        Task<GroupMember?> GetAsync(long groupId, long userId, IDbTransaction? trans = null);
        Task<bool> ExistsAsync(long groupId, long userId, IDbTransaction? trans = null);
        Task<long> AddAsync(GroupMember member);
        Task<bool> UpdateAsync(GroupMember member);
        
        Task<long> GetCoinsAsync(int coinsType, long groupId, long userId, IDbTransaction? trans = null);
        Task<bool> AddCoinsAsync(long botUin, long groupId, long userId, int coinsType, long amount, string reason);
        Task<long> GetCoinsForUpdateAsync(int coinsType, long groupId, long userId, IDbTransaction trans);
        
        Task<long> GetLongAsync(string field, long groupId, long userId, IDbTransaction? trans = null);
        Task<T> GetValueAsync<T>(string field, long groupId, long userId, IDbTransaction? trans = null);
        Task<int> SetValueAsync(string field, object value, long groupId, long userId, IDbTransaction? trans = null);
        Task<int> UpdateAsync(string fieldsSql, long groupId, long userId, IDbTransaction? trans = null);
        Task<int> IncrementValueAsync(string field, object value, long groupId, long userId, IDbTransaction? trans = null);
        Task<long> GetForUpdateAsync(string field, long groupId, long userId, IDbTransaction trans);
        
        Task<bool> UpdateSignInfoAsync(long groupId, long userId, int signTimes, int signLevel, IDbTransaction? trans = null);
        Task<int> GetSignDateDiffAsync(long groupId, long userId);
        Task<string> GetSignListAsync(long groupId, int topN = 10);
        
        Task<int> AppendAsync(long groupId, long userId, string name, string displayName = "", long groupCredit = 0, string confirmCode = "", IDbTransaction? trans = null);
        
        Task<long> GetCoinsRankingAsync(long groupId, long userId);
        Task<long> GetCoinsRankingAllAsync(long userId);
        Task<string> GetCreditRankingAsync(long groupId, int top, string format);
        
        Task<int> GetIntAsync(string field, long groupId, long userId, IDbTransaction? trans = null);

        Task SyncCacheFieldAsync(long groupId, long userId, string field, object value);
    }
}
