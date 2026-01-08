namespace BotWorker.Domain.Interfaces
{
    public interface IBotSendMessage
    {
        Task SendFinalMessageAsync(BotMessage ctx);
    }
}


