namespace BotWorker.Domain.Interfaces
{
    public interface IBotHandlerMessage
    {
        Task HandleBotMessageAsync(BotMessage context);
    }
}


