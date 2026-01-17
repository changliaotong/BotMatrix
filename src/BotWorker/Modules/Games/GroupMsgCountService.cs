using System;
using System.Threading.Tasks;
using BotWorker.Domain.Entities;
using BotWorker.Domain.Interfaces;
using BotWorker.Domain.Repositories;

namespace BotWorker.Modules.Games
{
    public class GroupMsgCountService : IGroupMsgCountService
    {
        private readonly IGroupMsgCountRepository _repository;
        private readonly IGroupRepository _groupRepository;

        public GroupMsgCountService(IGroupMsgCountRepository repository, IGroupRepository groupRepository)
        {
            _repository = repository;
            _groupRepository = groupRepository;
        }

        public async Task<bool> ExistTodayAsync(long groupId, long userId)
        {
            return await _repository.ExistTodayAsync(groupId, userId);
        }

        public async Task<int> AppendAsync(long botUin, long groupId, string groupName, long userId, string name)
        {
            return await _repository.AppendAsync(botUin, groupId, groupName, userId, name);
        }

        public async Task<int> UpdateAsync(long botUin, long groupId, string groupName, long userId, string name)
        {
            return await _repository.UpdateAsync(botUin, groupId, groupName, userId, name);
        }

        public async Task<int> GetMsgCountAsync(long groupId, long qq)
        {
            return await _repository.GetMsgCountAsync(groupId, qq);
        }

        public async Task<int> GetMsgCountYAsync(long groupId, long qq)
        {
            return await _repository.GetMsgCountAsync(groupId, qq, true);
        }

        public async Task<int> GetCountOrderAsync(long groupId, long userId)
        {
            return await _repository.GetCountOrderAsync(groupId, userId);
        }

        public async Task<int> GetCountOrderYAsync(long groupId, long userId)
        {
            return await _repository.GetCountOrderAsync(groupId, userId, true);
        }

        public async Task<string> GetCountListAsync(long botUin, long groupId, long userId, long top)
        {
            if (!await _groupRepository.IsOwnerAsync(groupId, userId) && !BotInfo.IsAdmin(botUin, userId))
                return "此命令仅限管理员使用";

            string res = await _repository.GetCountListAsync(groupId, false, top);
            if (!res.Contains(userId.ToString()))
            {
                int order = await GetCountOrderAsync(groupId, userId);
                int count = await GetMsgCountAsync(groupId, userId);
                res += $"【第{order}名】 你 发言：{count}";
            }
            res += "\n进入 后台 查看更多内容";
            return res;
        }

        public async Task<string> GetCountListYAsync(long botUin, long groupId, long userId, long top)
        {
            if (!await _groupRepository.IsOwnerAsync(groupId, userId) && !BotInfo.IsAdmin(botUin, userId))
                return "此命令仅限管理员使用";

            string res = await _repository.GetCountListAsync(groupId, true, top);
            if (!res.Contains(userId.ToString()))
            {
                int order = await GetCountOrderYAsync(groupId, userId);
                int count = await GetMsgCountYAsync(groupId, userId);
                res += $"【第{order}名】 你 发言：{count}";
            }
            res += "\n进入 后台 查看更多内容";
            return res;
        }
    }
}
