using BotWorker.Bots.BotMessages;

namespace BotWorker.Bots.Interfaces
{
    public interface IBotSendMessage
    {
        Task SendFinalMessageAsync(BotMessage ctx);
    }
}
