using BotWorker.Infrastructure.Persistence.ORM;
using BotWorker.Domain.Interfaces;
using System.Text;

namespace BotWorker.Modules.Games
{
    #region 数据模型

    public class UserMetric : MetaData<UserMetric>
    {
        public override string TableName => "UserMetrics";
        public override string KeyField => "Id";

        [BotWorker.Infrastructure.Utils.Schema.Attributes.PrimaryKey]
        public string Id { get; set; } = string.Empty; // Format: UserId_MetricKey
        public string UserId { get; set; } = string.Empty;
        public string MetricKey { get; set; } = string.Empty;
        public double Value { get; set; } = 0;
        public DateTime LastUpdateTime { get; set; } = DateTime.Now;

        public static async Task<UserMetric> GetOrCreateAsync(string userId, string key)
        {
            string id = $"{userId}_{key}";
            var metric = await GetSingleAsync(id);
            if (metric == null)
            {
                metric = new UserMetric { Id = id, UserId = userId, MetricKey = key, Value = 0 };
                await InsertAsync([
                    new Cov("Id", id),
                    new Cov("UserId", userId),
                    new Cov("MetricKey", key),
                    new Cov("Value", 0),
                    new Cov("LastUpdateTime", DateTime.Now)
                ]);
            }
            return metric;
        }
    }

    public class UserAchievement : MetaData<UserAchievement>
    {
        public override string TableName => "UserAchievements";
        public override string KeyField => "Id";

        [BotWorker.Infrastructure.Utils.Schema.Attributes.PrimaryKey]
        public string Id { get; set; } = string.Empty; // Format: UserId_AchievementId
        public string UserId { get; set; } = string.Empty;
        public string AchievementId { get; set; } = string.Empty;
        public DateTime UnlockTime { get; set; } = DateTime.Now;

        public static async Task<bool> IsUnlockedAsync(string userId, string achievementId)
        {
            return await GetSingleAsync($"{userId}_{achievementId}") != null;
        }
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
            await EnsureTablesCreatedAsync();
            await robot.RegisterSkillAsync(new SkillCapability
            {
                Name = "我的成就",
                Commands = ["我的成就", "成就排行", "成就详情"],
                Description = "查看已解锁的成就与进度"
            }, HandleCommandAsync);
        }

        public Task StopAsync() => Task.CompletedTask;

        private async Task EnsureTablesCreatedAsync()
        {
            try
            {
                var checkMetric = await UserMetric.QueryScalarAsync<int>($"SELECT COUNT(*) FROM {UserMetric.DbName}.INFORMATION_SCHEMA.TABLES WHERE TABLE_NAME = 'UserMetrics'");
                if (checkMetric == 0)
                {
                    await UserMetric.ExecAsync(BotWorker.Infrastructure.Utils.Schema.SchemaSynchronizer.GenerateCreateTableSql<UserMetric>());
                }

                var checkAch = await UserAchievement.QueryScalarAsync<int>($"SELECT COUNT(*) FROM {UserAchievement.DbName}.INFORMATION_SCHEMA.TABLES WHERE TABLE_NAME = 'UserAchievements'");
                if (checkAch == 0)
                {
                    await UserAchievement.ExecAsync(BotWorker.Infrastructure.Utils.Schema.SchemaSynchronizer.GenerateCreateTableSql<UserAchievement>());
                }
            }
            catch (Exception ex)
            {
                Console.WriteLine($"[Achievement] Table init failed: {ex.Message}");
                throw;
            }
        }

        private async Task<string> HandleCommandAsync(IPluginContext ctx, string[] args)
        {
            var cmd = ctx.RawMessage.Trim().Split(' ')[0];
            return cmd switch
            {
                "我的成就" => await GetUserAchievementsAsync(ctx.UserId),
                _ => "未知成就指令"
            };
        }

        private async Task<string> GetUserAchievementsAsync(string userId)
        {
            var unlocked = await UserAchievement.QueryWhere("UserId = @p1", UserAchievement.SqlParams(("@p1", userId)));
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
                        var metric = await UserMetric.GetOrCreateAsync(userId, def.MetricKey);
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
            var metric = await UserMetric.GetOrCreateAsync(userId, key);
            if (isAbsolute) metric.Value = delta;
            else metric.Value += delta;
            metric.LastUpdateTime = DateTime.Now;
            await metric.UpdateAsync();

            var newUnlocks = new List<string>();
            var relatedAchievements = Definitions.Where(d => d.MetricKey == key);

            foreach (var def in relatedAchievements)
            {
                if (metric.Value >= def.Threshold)
                {
                    if (!await UserAchievement.IsUnlockedAsync(userId, def.Id))
                    {
                        await new UserAchievement 
                        { 
                            Id = $"{userId}_{def.Id}", 
                            UserId = userId, 
                            AchievementId = def.Id,
                            UnlockTime = DateTime.Now 
                        }.InsertAsync();
                        newUnlocks.Add(def.Name);
                    }
                }
            }
            return newUnlocks;
        }
    }
}
