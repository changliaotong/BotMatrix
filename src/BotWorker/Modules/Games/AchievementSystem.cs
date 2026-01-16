using BotWorker.Domain.Interfaces;
using BotWorker.Domain.Repositories;
using Microsoft.Extensions.DependencyInjection;
using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;

namespace BotWorker.Modules.Games
{
    #region 数据模型

    [Dapper.Contrib.Extensions.Table("UserMetrics")]
    public class UserMetric
    { 
        [Dapper.Contrib.Extensions.ExplicitKey]
        public string Id { get; set; } = string.Empty; // Format: UserId_MetricKey
        public string UserId { get; set; } = string.Empty;
        public string MetricKey { get; set; } = string.Empty;
        public double Value { get; set; } = 0;
        public DateTime LastUpdateTime { get; set; } = DateTime.Now;
    }

    [Dapper.Contrib.Extensions.Table("UserAchievements")]
    public class UserAchievement
    {
        [Dapper.Contrib.Extensions.ExplicitKey]
        public string Id { get; set; } = string.Empty; // Format: UserId_AchievementId
        public string UserId { get; set; } = string.Empty;
        public string AchievementId { get; set; } = string.Empty;
        public DateTime UnlockTime { get; set; } = DateTime.Now;
    }

    public class AchievementDef
    {
        public string Id { get; set; } = string.Empty;
        public string Name { get; set; } = string.Empty;
        public string Description { get; set; } = string.Empty;
        public string MetricKey { get; set; } = string.Empty;
        public double Threshold { get; set; }
        public int RewardGold { get; set; }
        public string Category { get; set; } = "General";
    }

    #endregion

    [BotPlugin(
        Id = "sys.achievement",
        Name = "成就系统",
        Version = "1.0.0",
        Author = "Matrix",
        Description = "跨模块成就追踪与奖励系统",
        Category = "System"
    )]
    public class AchievementPlugin : IPlugin
    {
        private readonly IUserMetricRepository _metricRepo;
        private readonly IUserAchievementRepository _achievementRepo;

        public AchievementPlugin(IUserMetricRepository metricRepo, IUserAchievementRepository achievementRepo)
        {
            _metricRepo = metricRepo;
            _achievementRepo = achievementRepo;
        }

        public static List<AchievementDef> Definitions = new()
        {
            // 钓鱼成就
            new AchievementDef { Id = "fish_10", Name = "初学者", Description = "累计钓到 10 条鱼", MetricKey = "fishing.catch_count", Threshold = 10, RewardGold = 100, Category = "Fishing" },
            new AchievementDef { Id = "fish_100", Name = "钓鱼达人", Description = "累计钓到 100 条鱼", MetricKey = "fishing.catch_count", Threshold = 100, RewardGold = 1000, Category = "Fishing" },
            new AchievementDef { Id = "fish_gold_10000", Name = "渔业大亨", Description = "卖鱼累计获得 10,000 金币", MetricKey = "fishing.total_gold", Threshold = 10000, RewardGold = 2000, Category = "Fishing" },
            
            // 宠物成就
            new AchievementDef { Id = "pet_adopt", Name = "爱心大使", Description = "成功领养第一只宠物", MetricKey = "pet.adopt_count", Threshold = 1, RewardGold = 200, Category = "Pet" },
            new AchievementDef { Id = "pet_level_10", Name = "金牌教练", Description = "宠物等级达到 10 级", MetricKey = "pet.max_level", Threshold = 10, RewardGold = 500, Category = "Pet" },
            
            // 婚姻与育儿成就
            new AchievementDef { Id = "marry_1", Name = "成家立业", Description = "成功与心仪的对象结婚", MetricKey = "marriage.count", Threshold = 1, RewardGold = 520, Category = "Social" },
            new AchievementDef { Id = "baby_1", Name = "初为人父/母", Description = "领养第一个宝宝", MetricKey = "baby.adopt_count", Threshold = 1, RewardGold = 666, Category = "Baby" },
            
            // 通用成就
            new AchievementDef { Id = "msg_1000", Name = "话痨", Description = "累计发送 1,000 条消息", MetricKey = "sys.msg_count", Threshold = 1000, RewardGold = 500, Category = "Social" }
        };

        public async Task InitAsync(IRobot robot)
        {
            await robot.RegisterSkillAsync(new SkillCapability
            {
                Name = "我的成就",
                Commands = ["成就", "我的成就", "勋章"],
                Description = "查看已解锁的成就与勋章"
            }, async (ctx, args) => await GetUserAchievementsAsync(ctx.UserId));
        }

        public Task StopAsync() => Task.CompletedTask;

        private async Task<string> GetUserAchievementsAsync(string userId)
        {
            var unlocked = await _achievementRepo.GetByUserIdAsync(userId);
            var unlockedIds = unlocked.Select(a => a.AchievementId).ToHashSet();

            var sb = new StringBuilder();
            sb.AppendLine("🏆 【我的成就勋章】");
            sb.AppendLine("━━━━━━━━━━━━━━");

            foreach (var category in Definitions.GroupBy(d => d.Category))
            {
                sb.AppendLine($"\n📂 {category.Key}");
                foreach (var def in category)
                {
                    var isDone = unlockedIds.Contains(def.Id);
                    var icon = isDone ? "✅" : "🔒";
                    sb.AppendLine($"{icon} {def.Name}: {def.Description}");
                    if (!isDone)
                    {
                        var metric = await _metricRepo.GetOrCreateAsync(userId, def.MetricKey);
                        sb.AppendLine($"   进度: {metric.Value}/{def.Threshold}");
                    }
                }
            }

            return sb.ToString();
        }

        /// <summary>
        /// 报告指标并检查成就
        /// </summary>
        public static async Task<List<string>> ReportMetricAsync(string userId, string key, double delta, bool isAbsolute = false)
        {
            var metricRepo = BotMessage.ServiceProvider?.GetRequiredService<IUserMetricRepository>() 
                ?? throw new InvalidOperationException("IUserMetricRepository not registered");
            var achievementRepo = BotMessage.ServiceProvider?.GetRequiredService<IUserAchievementRepository>() 
                ?? throw new InvalidOperationException("IUserAchievementRepository not registered");

            var metric = await metricRepo.GetOrCreateAsync(userId, key);
            if (isAbsolute) metric.Value = delta;
            else metric.Value += delta;
            metric.LastUpdateTime = DateTime.Now;
            await metricRepo.UpdateAsync(metric);

            var newUnlocks = new List<string>();
            var relatedAchievements = Definitions.Where(d => d.MetricKey == key);

            foreach (var def in relatedAchievements)
            {
                if (metric.Value >= def.Threshold)
                {
                    if (!await achievementRepo.IsUnlockedAsync(userId, def.Id))
                    {
                        await achievementRepo.InsertAsync(new UserAchievement 
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
