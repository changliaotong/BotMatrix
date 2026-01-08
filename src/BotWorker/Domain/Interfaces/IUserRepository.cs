using System.Threading.Tasks;

namespace BotWorker.Domain.Interfaces
{
    public interface IUserRepository
    {
        Task<bool> IsBlacklistedAsync(long userId, long groupId);
        Task<int> AddToBlacklistAsync(long userId, long groupId, string reason);
        Task<int> GetPointsAsync(long userId);
        Task<int> UpdatePointsAsync(long userId, int delta);
        Task<bool> HasSignedIdAsync(long userId, string date);
    }
}
