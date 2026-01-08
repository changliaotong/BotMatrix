using System;
using System.Threading.Tasks;

namespace BotWorker.Core.Repositories
{
    public interface IUserRepository
    {
        Task<bool> IsSuperAdminAsync(long userId);
        Task<bool> IsBlackAsync(long userId);
        Task<int> SetIsBlackAsync(long userId, bool isBlack, string reason = "");
        Task<int> AddUserAsync(long botUin, long groupId, long userId, string name, long refUserId, string userOpenId = "", string groupOpenId = "");
        Task<int> SetStateAsync(long userId, int state);
        Task<int> GetStateAsync(long userId);
        Task<bool> ExistsAsync(long userId);
        Task<long> GetCreditAsync(long groupId, long userId);
        Task<int> AddCreditAsync(long botUin, long groupId, long userId, long credit, string reason);
        Task<long> GetSaveCreditAsync(long userId);
        Task<int> AddSaveCreditAsync(long userId, long credit, string reason);
        Task<long> GetCoinsAsync(int coinsType, long groupId, long userId);
        Task<int> AddCoinsAsync(int coinsType, long coinsValue, long groupId, long userId, string reason);

        /// <summary>
        /// 获取积分排行榜
        /// </summary>
        Task<System.Collections.Generic.IEnumerable<(long UserId, long Credit)>> GetCreditRankAsync(long groupId, int top = 10);
    }
}
