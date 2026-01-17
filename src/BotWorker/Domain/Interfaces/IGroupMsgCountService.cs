using System.Threading.Tasks;

namespace BotWorker.Domain.Interfaces
{
    public interface IGroupMsgCountService
    {
        Task<bool> ExistTodayAsync(long groupId, long userId);
        Task<int> AppendAsync(long botUin, long groupId, string groupName, long userId, string name);
        Task<int> UpdateAsync(long botUin, long groupId, string groupName, long userId, string name);
        Task<int> GetMsgCountAsync(long groupId, long qq);
        Task<int> GetMsgCountYAsync(long groupId, long qq);
        Task<int> GetCountOrderAsync(long groupId, long userId);
        Task<int> GetCountOrderYAsync(long groupId, long userId);
        Task<string> GetCountListAsync(long botUin, long groupId, long userId, long top);
        Task<string> GetCountListYAsync(long botUin, long groupId, long userId, long top);
    }
}
