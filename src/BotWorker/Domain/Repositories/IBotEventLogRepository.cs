using System.Threading.Tasks;

namespace BotWorker.Domain.Repositories
{
    public interface IBotEventLogRepository
    {
        Task<int> AppendAsync(long botUin, string eventName, long groupId, string groupName, long userId, string userName);
    }
}
