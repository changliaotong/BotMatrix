using System.Data;
using BotWorker.Domain.Models.BotMessages;

namespace BotWorker.Domain.Repositories
{
    public interface IBotMessageRepository
    {
        Task<string> InsertAsync(BotMessage message, IDbTransaction? trans = null);
        Task<bool> UpdateAsync(BotMessage message, IDbTransaction? trans = null);
        Task<bool> DeleteAsync(string msgId, IDbTransaction? trans = null);
        Task<BotMessage?> GetByMsgIdAsync(string msgId, IDbTransaction? trans = null);
    }
}
