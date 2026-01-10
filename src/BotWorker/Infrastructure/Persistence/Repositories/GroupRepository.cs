using System.Threading.Tasks;
using BotWorker.Domain.Interfaces;
using BotWorker.Domain.Entities;
using BotWorker.Infrastructure.Persistence.ORM;

namespace BotWorker.Infrastructure.Persistence.Repositories
{
    public class GroupRepository : IGroupRepository
    {
        public async Task<long> GetGroupOwnerAsync(long groupId)
        {
            return await GroupInfo.GetLongAsync("GroupOwner", groupId);
        }

        public async Task<long> GetRobotOwnerAsync(long groupId)
        {
            return await GroupInfo.GetLongAsync("RobotOwner", groupId);
        }

        public async Task<int> SetRobotOwnerAsync(long groupId, long ownerId)
        {
            return await GroupInfo.SetValueAsync("RobotOwner", ownerId, groupId);
        }

        public async Task<int> SetInGameAsync(long groupId, int isInGame)
        {
            // Note: IsInGame is marked as [DbIgnore] in GroupInfo.cs, but it might be needed for in-memory status or a separate table.
            // For now, we follow the interface. If it's not in DB, we might need to handle it differently.
            return await GroupInfo.SetValueAsync("IsInGame", isInGame, groupId);
        }

        public async Task<int> UpdateGroupAsync(long groupId, string name, long selfId, long groupOwner = 0, long robotOwner = 0)
        {
            var fields = new List<Cov>
            {
                new Cov("GroupName", name),
                new Cov("BotUin", selfId)
            };
            if (groupOwner != 0) fields.Add(new Cov("GroupOwner", groupOwner));
            if (robotOwner != 0) fields.Add(new Cov("RobotOwner", robotOwner));

            if (await GroupInfo.ExistsAsync(groupId))
            {
                return await GroupInfo.UpdateAsync(fields, groupId);
            }
            else
            {
                fields.Add(new Cov("Id", groupId));
                return await GroupInfo.InsertAsync(fields);
            }
        }

        public async Task<bool> IsOpenAsync(long groupId)
        {
            return await GroupInfo.GetBoolAsync("IsOpen", groupId);
        }

        public async Task<int> SetOpenStatusAsync(long groupId, bool isOpen)
        {
            return await GroupInfo.SetValueAsync("IsOpen", isOpen, groupId);
        }

        public async Task<int> GetVipRestDaysAsync(long groupId)
        {
            return await GroupVip.RestDaysAsync(groupId);
        }

        public async Task<bool> IsSz84Async(long groupId)
        {
            return await GroupInfo.GetBoolAsync("IsSz84", groupId);
        }
    }
}
