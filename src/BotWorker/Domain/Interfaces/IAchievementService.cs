using System.Collections.Generic;
using System.Threading.Tasks;

namespace BotWorker.Domain.Interfaces
{
    public interface IAchievementService
    {
        // Note: These types might need to be moved to Domain/Models/Achievement
        Task<List<object>> GetAllAchievementsAsync();
        Task<object?> GetProgressAsync(long userId, string achievementId);
        Task<bool> AddProgressAsync(long userId, string achievementId, int amount);
    }
}


