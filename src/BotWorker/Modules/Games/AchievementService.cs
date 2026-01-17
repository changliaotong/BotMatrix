using BotWorker.Domain.Interfaces;
using BotWorker.Domain.Repositories;
using BotWorker.Modules.Games;
using System;
using System.Collections.Generic;
using System.Linq;
using System.Threading.Tasks;

namespace BotWorker.Modules.Games
{
    public class AchievementService : IAchievementService
    {
        private readonly IUserMetricRepository _metricRepo;
        private readonly IUserAchievementRepository _achievementRepo;

        public AchievementService(IUserMetricRepository metricRepo, IUserAchievementRepository achievementRepo)
        {
            _metricRepo = metricRepo;
            _achievementRepo = achievementRepo;
        }

        public async Task<List<string>> ReportMetricAsync(string userId, string key, double delta, bool isAbsolute = false)
        {
            var metric = await _metricRepo.GetOrCreateAsync(userId, key);
            if (isAbsolute) metric.Value = delta;
            else metric.Value += delta;
            metric.LastUpdateTime = DateTime.Now;
            await _metricRepo.UpdateAsync(metric);

            var newUnlocks = new List<string>();
            var relatedAchievements = AchievementPlugin.Definitions.Where(d => d.MetricKey == key);

            foreach (var def in relatedAchievements)
            {
                if (metric.Value >= def.Threshold)
                {
                    if (!await _achievementRepo.IsUnlockedAsync(userId, def.Id))
                    {
                        await _achievementRepo.InsertAsync(new UserAchievement 
                        { 
                            Id = $"{userId}_{def.Id}", 
                            UserId = userId, 
                            AchievementId = def.Id,
                            UnlockTime = DateTime.Now 
                        });
                        newUnlocks.Add(def.Name);
                    }
                }
            }
            return newUnlocks;
        }
    }
}
