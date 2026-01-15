using System.Threading.Tasks;
using BotWorker.Domain.Entities;

namespace BotWorker.Domain.Repositories
{
    public interface IGroupSendMessageRepository : IBaseRepository<GroupSendMessage>
    {
        Task<int> UserCountAsync(long groupId);
        Task<int> AppendAsync(GroupSendMessage entity);
    }
}
