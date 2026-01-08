namespace BotWorker.Domain.Entities.Achievement
{
    public interface IAchievementService
    {
        Task<List<Achievement>> GetAllAchievementsAsync();
        Task<UserAchievement?> GetProgressAsync(long userId, string achievementId);
        Task<bool> AddProgressAsync(long userId, string achievementId, int amount);
    }
}