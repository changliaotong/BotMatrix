using System.Threading.Tasks;

namespace BotWorker.Core.Repositories
{
    /// <summary>
    /// 群组数据仓库：处理群组配置、开关状态、VIP信息�?    /// </summary>
    public interface IGroupRepository
    {
        Task<bool> IsOpenAsync(long groupId);
        Task<int> SetOpenStatusAsync(long groupId, bool isOpen);
        Task<int> GetVipRestDaysAsync(long groupId);
        Task<bool> IsSz84Async(long groupId);
        // 其他群组配置相关的底层操�?    }

    /// <summary>
    /// 用户数据仓库：处理黑名单、积分、签到记录、权限等
    /// </summary>
    public interface IUserRepository
    {
        Task<bool> IsBlacklistedAsync(long userId, long groupId);
        Task<int> AddToBlacklistAsync(long userId, long groupId, string reason);
        Task<int> GetPointsAsync(long userId);
        Task<int> UpdatePointsAsync(long userId, int delta);
        Task<bool> HasSignedIdAsync(long userId, string date);
        // 其他用户相关的底层操�?    }
}


