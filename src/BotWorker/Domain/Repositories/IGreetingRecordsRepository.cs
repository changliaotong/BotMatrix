using System;
using System.Threading.Tasks;

namespace BotWorker.Domain.Repositories
{
    public interface IGreetingRecordsRepository
    {
        Task<int> AppendAsync(long botQQ, long groupId, string groupName, long qq, string name, int greetingType = 0);
        Task<bool> ExistsAsync(long groupId, long qq, int greetingType = 0);
        Task<int> GetCountAsync(int greetingType = 0);
        Task<int> GetCountAsync(long groupId, int greetingType = 0);
    }
}
