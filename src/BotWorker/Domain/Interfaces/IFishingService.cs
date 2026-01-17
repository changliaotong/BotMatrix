using System.Threading.Tasks;

namespace BotWorker.Domain.Interfaces
{
    public interface IFishingService
    {
        Task<string> GetStatusAsync(long userId, string nickname);
        Task<string> CastAsync(long userId);
        Task<string> ReelInAsync(long userId);
        Task<string> GetBagAsync(long userId);
        Task<string> SellFishAsync(long userId);
        Task<string> GetShopAsync(long userId);
        Task<string> UpgradeRodAsync(long userId);
        Task<string> HandleFishingAsync(long userId, string userName, string cmd);
    }
}
