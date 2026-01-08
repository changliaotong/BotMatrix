using System;
using System.Threading.Tasks;
using System.Collections.Generic;

namespace BotWorker.Core.Repositories
{
    public interface IGroupRepository
    {
        Task<long> GetGroupOwnerAsync(long groupId);
        Task<long> GetRobotOwnerAsync(long groupId);
        Task<int> SetRobotOwnerAsync(long groupId, long ownerId);
        Task<int> SetInGameAsync(long groupId, int isInGame);
        Task<int> UpdateGroupAsync(long groupId, string name, long selfId, long groupOwner = 0, long robotOwner = 0);
        Task<bool> GetIsOpenAsync(long groupId);
        Task<int> SetIsOpenAsync(long groupId, bool isOpen);
    }
}


