using BotWorker.Core.Interfaces;

namespace BotWorker.Domain.Entities.Achievement
{
    public class AchievementManager : IAchievementService
    {
        private readonly Dictionary<string, Achievement> _definitions = new();

        public AchievementManager()
        {
        }

        public void Register(Achievement def)
        {
            _definitions[def.Id] = def;
        }

        public void TryUnlockAchievements(long userId)
        {

            foreach (var def in _definitions.Values)
            {



            }
        }

 

        public Task<List<Achievement>> GetAllAchievementsAsync()
        {
            throw new NotImplementedException();
        }

        public Task<UserAchievement?> GetProgressAsync(long userId, string achievementId)
        {
            throw new NotImplementedException();
        }

        public Task<bool> AddProgressAsync(long userId, string achievementId, int amount)
        {
            throw new NotImplementedException();
        }
    }

}
