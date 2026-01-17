using BotWorker.Domain.Models.BotMessages;

namespace BotWorker.Domain.Interfaces
{
    public interface IRedBlueService
    {
        Task<string> GetRedBlueResAsync(BotMessage botMsg, bool isDetail = true);
    }
}
