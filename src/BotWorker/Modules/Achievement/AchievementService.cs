using BotWorker.Domain.Interfaces;
using BotWorker.Domain.Constants;
using BotWorker.Core.Data;

namespace BotWorker.Modules.Achievement
{
    public partial class AchievementService(AppDbContext db) : IAchievementService
    {

        private readonly AppDbContext _db = db;

        // 事件：用户签到
        public async Task OnUserSignedIn(long userId)
        {
            const string achievementId = "signin_1";
            await AddProgressAsync(userId, achievementId, 1);
        }

        // 事件：用户发言（增加发言次数）
        public async Task OnUserSentMessage(long userId)
        {
            const string achievementId = "chat_100";
            await AddProgressAsync(userId, achievementId, 1);
        }

        // 事件：用户揍群主
        public async Task OnUserPokedBoss(long userId)
        {
            const string achievementId = "poke_boss";
            await AddProgressAsync(userId, achievementId, 1);
        }

        private async Task<string> ProcessAchievementProgress(long userId, string achievementId, int increment, string baseMsg)
        {
            var achievement = await _db.Achievements.FindAsync(achievementId);
            if (achievement == null) return $"成就 {achievementId} 未配置。";

            // 模拟从规则配置解析，演示时硬编码规则
            var rules = achievement.Rules.OrderBy(r => r.Level).ToList();
            if (!rules.Any())
            {
                // 简单升级规则，兼容旧逻辑
                rules = new List<AchievementRule>
        {
            new AchievementRule { Level = 1, RequiredCount = 1, Reward = achievement.Reward }
        };
            }

            var progress = await _db.UserAchievement.FindAsync(userId, achievementId);
            if (progress == null)
            {
                progress = new UserAchievement
                {
                    UserId = userId,
                    AchievementId = achievementId,
                    CurrentLevel = 0,
                    CurrentValue = 0,
                    CurrentStreakDays = 0,
                    LastActionDate = DateTime.MinValue,
                    LastActionDatePrev = null
                };
                _db.UserAchievement.Add(progress);
            }

            var now = DateTime.UtcNow.Date;
            var yesterday = now.AddDays(-1);

            // 判断是否连续签到（以签到举例）
            if (achievementId == "signin_1")
            {
                if (progress.LastActionDatePrev == yesterday)
                {
                    progress.CurrentStreakDays++;
                }
                else if (progress.LastActionDatePrev == now)
                {
                    // 当天已签到，不累加
                }
                else
                {
                    progress.CurrentStreakDays = 1; // 断连重置
                }
            }

            progress.CurrentValue += increment;
            progress.LastActionDatePrev = now;
            progress.LastActionDate = DateTime.UtcNow;

            // 判断是否满足升级条件
            var nextLevel = progress.CurrentLevel + 1;
            var nextRule = rules.FirstOrDefault(r => r.Level == nextLevel);
            if (nextRule == null)
            {
                await _db.SaveChangesAsync();
                return baseMsg + " （已达最高等级）";
            }

            // 判断连续天数是否满足（如果有要求）
            if (nextRule.RequiredStreakDays.HasValue && progress.CurrentStreakDays < nextRule.RequiredStreakDays)
            {
                await _db.SaveChangesAsync();
                return baseMsg + $" （未满足连续签到天数 {nextRule.RequiredStreakDays}）";
            }

            // 判断累计次数是否满足
            if (progress.CurrentValue >= nextRule.RequiredCount)
            {
                progress.CurrentLevel = nextLevel;
                var rewards = ParseReward(nextRule.Reward, nextLevel);
                await _db.SaveChangesAsync();
            }

            await _db.SaveChangesAsync();
            return baseMsg + $" 当前进度：{progress.CurrentValue}/{nextRule.RequiredCount}";
        }

        private List<(string Type, int Amount, string Value)> ParseReward(string rewardStr, int level)
        {
            var list = new List<(string, int, string)>();
            if (string.IsNullOrEmpty(rewardStr)) return list;

            var parts = rewardStr.Split(';');
            foreach (var part in parts)
            {
                var subParts = part.Split(':');
                if (subParts.Length >= 2)
                {
                    var type = subParts[0];
                    if (type == "point" && int.TryParse(subParts[1], out int amount))
                    {
                        list.Add((type, amount * level, ""));
                    }
                    else if (type == "title")
                    {
                        list.Add((type, 0, subParts[1]));
                    }
                }
            }
            return list;
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

        internal void DailyRefreshShop()
        {
            throw new NotImplementedException();
        }

        internal string UserSignIn(long userId)
        {
            throw new NotImplementedException();
        }

        internal string UserPokeBoss(long userId)
        {
            throw new NotImplementedException();
        }

        internal object GetUserPointsAsync(long userId)
        {
            throw new NotImplementedException();
        }

        internal string GetTitles(long userId)
        {
            throw new NotImplementedException();
        }

        internal string ShowShop()
        {
            throw new NotImplementedException();
        }

        internal string SetUserTitle(long userId, string title)
        {
            throw new NotImplementedException();
        }

        internal string Redeem(long userId, int id)
        {
            throw new NotImplementedException();
        }
    }
}
