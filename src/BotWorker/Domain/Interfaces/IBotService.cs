namespace BotWorker.Domain.Interfaces
{
    public interface IBotService
    {
        bool IsSuperAdmin(long userId);
    }
}
