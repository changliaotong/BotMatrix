using BotWorker.Bots.BotMessages;

namespace BotWorker.Bots.Interfaces
{
    public interface IBotHandlerMessage
    {
        Task HandleBotMessageAsync(BotMessage context);
    }

}
