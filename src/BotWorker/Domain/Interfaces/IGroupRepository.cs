using System.Threading.Tasks;

namespace BotWorker.Domain.Interfaces
{
    public interface IGroupRepository
    {
        Task<long> GetGroupOwnerAsync(long groupId);
        Task<long> GetRobotOwnerAsync(long groupId);
        Task<int> SetRobotOwnerAsync(long groupId, long ownerId);
        Task<int> SetInGameAsync(long groupId, int isInGame);
        Task<int> UpdateGroupAsync(long groupId, string name, long selfId, long groupOwner = 0, long robotOwner = 0);
        Task<bool> IsOpenAsync(long groupId);
        Task<int> SetOpenStatusAsync(long groupId, bool isOpen);
        Task<int> GetVipRestDaysAsync(long groupId);
        Task<bool> IsSz84Async(long groupId);
    }
}
