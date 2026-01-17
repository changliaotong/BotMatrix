using BotWorker.Domain.Models.BotMessages;
using BotWorker.Infrastructure.Communication.OneBot;

namespace BotWorker.Infrastructure.Messaging
{
    public interface IBotMessageMapper
    {
        Task<BotMessage?> MapToOneBotEventAsync(string json, IOneBotApiClient? apiClient = null);
    }
}
