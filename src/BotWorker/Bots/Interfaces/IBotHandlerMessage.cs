using sz84.Bots.BotMessages;

namespace sz84.Bots.Interfaces
{
    public interface IBotHandlerMessage
    {
        Task HandleBotMessageAsync(BotMessage context);
    }

}
