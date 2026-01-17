using System.Threading.Tasks;

namespace BotWorker.Domain.Interfaces
{
    public interface IGroupGiftService
    {
        Task<string> GetGiftAsync(long groupId, long userId);
        Task<string> GetGiftResAsync(long botUin, long groupId, string groupName, long userId, string name, long qqGift, string giftName, int giftCount);
        bool IsFans(long groupId, long userId);
        Task<bool> IsFansAsync(long groupId, long userId);
        Task<long> GetFansValueAsync(long groupId, long userId);
        Task<long> GetFansRankingAsync(long groupId, long userId);
        Task<int> GetFansLevelAsync(long groupId, long userId);
        Task<GroupMemberInfo?> GetMemberInfoAsync(long groupId, long userId);
        int LampMinutes(long groupId, long userId);
        (string sql, object paras) SqlLightLamp(long groupId, long userId);
        (string sql, object paras) SqlBingFans(long groupId, long userId);
        int GetFansCount(long groupId);
    }
}
