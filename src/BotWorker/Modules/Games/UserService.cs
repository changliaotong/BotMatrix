using System.Data;
using System.Linq;
using System.Threading.Tasks;
using BotWorker.Common.Extensions;
using BotWorker.Domain.Entities;
using BotWorker.Domain.Interfaces;
using BotWorker.Domain.Repositories;

namespace BotWorker.Modules.Games
{
    public class UserService : IUserService
    {
        private readonly IUserRepository _userRepository;
        private readonly IGroupRepository _groupRepository;
        private readonly IGroupMemberRepository _groupMemberRepository;
        private readonly IFriendRepository _friendRepository;

        public UserService(
            IUserRepository userRepository,
            IGroupRepository groupRepository,
            IGroupMemberRepository groupMemberRepository,
            IFriendRepository friendRepository)
        {
            _userRepository = userRepository;
            _groupRepository = groupRepository;
            _groupMemberRepository = groupMemberRepository;
            _friendRepository = friendRepository;
        }

        public async Task<int> SetStateAsync(int state, long userId)
        {
            return await _userRepository.SetValueAsync("state", state, userId);
        }

        public async Task<int> AppendAsync(long botUin, long groupId, long userId, string name, long refUserId, string userOpenId = "", string groupOpenId = "", IDbTransaction? trans = null)
        {
            if (await _userRepository.ExistsAsync(userId, null, trans)) return 0;

            return await _userRepository.InsertAsync(new
            {
                BotUin = botUin,
                UserOpenid = userOpenId,
                GroupOpenid = groupOpenId,
                GroupId = groupId,
                Id = userId,
                Credit = string.IsNullOrEmpty(userOpenId) ? 0 : 5000,
                Name = name,
                RefUserId = refUserId,
            }, trans);
        }

        public async Task<long> GetCreditAsync(long userId)
        {
            return await _userRepository.GetCreditAsync(0, 0, userId);
        }

        public async Task<long> GetCreditAsync(long groupId, long userId)
        {
            if (groupId != 0 && await _groupRepository.GetIsCreditAsync(groupId))
            {
                // This logic might need to be refined if GroupMember.GetCoinsAsync is moved to a service too
                // For now, using Repository directly if available
                return await _groupMemberRepository.GetCoinsAsync(1, groupId, userId); // 1 is groupCredit
            }
            return await GetCreditAsync(userId);
        }

        public async Task<int> AppendUserAsync(long botUin, long groupId, long userId, string name, string userOpenId = "", string groupOpenId = "")
        {
            long[] specialIds = { 2107992324, 3677524472, 3662527857, 2174158062, 2188157235, 3375620034, 1611512438, 3227607419, 3586811032,
                        3835195413, 3527470977, 3394199803, 2437953621, 3082166471, 2375832958, 1807139582, 2704647312, 1420694846, 3788007880 };

            if (specialIds.Contains(userId)) return 0;

            var groupOwner = await _groupRepository.GetGroupOwnerAsync(groupId);
            int i = await AppendAsync(botUin, groupId, userId, name, groupOwner, userOpenId, groupOpenId);
            if (i == -1) return i;

            i = await _groupMemberRepository.AppendAsync(groupId, userId, name, "");
            if (i == -1) return i;

            if (await BotInfo.GetIsCreditAsync(botUin))
            {
                i = await _friendRepository.AppendAsync(botUin, userId, name);
                if (i == -1) return i;
            }

            return i;
        }

        public async Task<string> GetResetDefaultGroupAsync(long userId)
        {
            // BotInfo.GroupCrm is static, so it's fine
            return await _userRepository.SetValueAsync("DefaultGroup", BotInfo.GroupCrm, userId) == -1 
                ? "" 
                : $"\n默认群已重置为 {BotInfo.GroupCrm}";
        }
    }
}
