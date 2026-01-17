using System.Collections.Generic;
using System.Threading.Tasks;
using BotWorker.Domain.Entities;
using BotWorker.Modules.AI.Models;

namespace BotWorker.Domain.Repositories
{
    public interface IGroupSendMessageRepository : IBaseRepository<GroupSendMessage>
    {
        Task<int> UserCountAsync(long groupId);
        Task<int> AppendAsync(GroupSendMessage entity);
        Task<IEnumerable<ChatHistoryItem>> GetChatHistoryAsync(long groupId, long userId, bool isMultAI, int context);
    }
}
