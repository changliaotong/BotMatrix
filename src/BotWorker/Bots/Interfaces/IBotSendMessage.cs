using sz84.Bots.BotMessages;

namespace sz84.Bots.Interfaces
{
    public interface IBotSendMessage
    {
        Task SendFinalMessageAsync(BotMessage ctx);
    }
}
