using System.Threading.Tasks;
using BotWorker.Domain.Entities;
using BotWorker.Domain.Repositories;

namespace BotWorker.Infrastructure.Persistence.Repositories
{
    public class BotEventLogRepository : BaseRepository<BotEventLog>, IBotEventLogRepository
    {
        public BotEventLogRepository() : base("Event")
        {
        }

        public async Task<int> AppendAsync(long botUin, string eventName, long groupId, string groupName, long userId, string userName)
        {
            var log = new BotEventLog
            {
                BotUin = botUin,
                EventName = eventName,
                GroupId = groupId,
                GroupName = groupName,
                UserId = userId,
                UserName = userName
            };
            await InsertAsync(log);
            return 1;
        }
    }
}
