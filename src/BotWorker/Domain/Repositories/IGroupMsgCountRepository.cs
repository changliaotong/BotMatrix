using System.Collections.Generic;
using System.Threading.Tasks;
using BotWorker.Domain.Entities;

namespace BotWorker.Domain.Repositories
{
    public interface IGroupMsgCountRepository
    {
        Task<bool> ExistTodayAsync(long groupId, long userId);
        Task<int> AppendAsync(long botUin, long groupId, string groupName, long userId, string name);
        Task<int> UpdateAsync(long botUin, long groupId, string groupName, long userId, string name);
        Task<int> GetMsgCountAsync(long groupId, long userId, bool yesterday = false);
        Task<int> GetCountOrderAsync(long groupId, long userId, bool yesterday = false);
        Task<string> GetCountListAsync(long groupId, bool yesterday = false, long top = 10);
    }
}
