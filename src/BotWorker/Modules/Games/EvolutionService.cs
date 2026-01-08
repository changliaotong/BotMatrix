using BotWorker.Domain.Interfaces;
using BotWorker.Domain.Models;
using BotWorker.Infrastructure.Utils.Schema;
using BotWorker.Infrastructure.Persistence.ORM;
using Microsoft.Extensions.Logging;
using System;
using System.Collections.Generic;
using System.Linq;
using System.Threading.Tasks;

namespace BotWorker.Modules.Games
{
    [BotPlugin(
        Id = "evolution",
        Name = "进化与等级系统",
        Version = "1.0.0",
        Author = "BotMatrix Evolution",
        Description = "基于积分交易自动增长经验，实现全自动等级晋升系统。",
        Category = "Social"
    )]
    public class EvolutionService : IPlugin
    {
        private readonly ILogger<EvolutionService>? _logger;
        private IRobot? _robot;

        public EvolutionService() { }
        public EvolutionService(ILogger<EvolutionService> logger)
        {
            _logger = logger;
        }

        public List<Intent> Intents => [
            new() { Name = "等级查询", Keywords = ["等级", "经验", "level", "exp"] }
        ];

        public async Task InitAsync(IRobot robot)
        {
            _robot = robot;

            // 自动同步表结构
            await EnsureTablesCreatedAsync();

            // 注册指令
            await robot.RegisterSkillAsync(new SkillCapability
            {
                Name = "等级系统",
                Commands = ["等级", "经验", "level", "exp"],
                Description = "查看您的当前等级与经验值"
            }, HandleCommandAsync);

            // 订阅积分交易事件，实现自动经验增长
            robot.Events.Subscribe<PointTransactionEvent>(OnPointTransactionAsync);
            
            // 订阅系统交互事件，实现新手引导任务
            robot.Events.Subscribe<SystemInteractionEvent>(OnSystemInteractionAsync);

            _logger?.LogInformation("EvolutionService 已启动并成功订阅事件中枢");
        }

        public Task StopAsync()
        {
            _robot?.Events.Unsubscribe<PointTransactionEvent>(OnPointTransactionAsync);
            _robot?.Events.Unsubscribe<SystemInteractionEvent>(OnSystemInteractionAsync);
            return Task.CompletedTask;
        }

        private async Task EnsureTablesCreatedAsync()
        {
            try
            {
                var checkTable = await UserLevel.QueryScalarAsync<int>("SELECT COUNT(*) FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_NAME = 'UserLevels'");
                if (checkTable == 0)
                {
                    await UserLevel.ExecAsync(SchemaSynchronizer.GenerateCreateTableSql<UserLevel>());
                }
            }
            catch (Exception ex)
            {
                _logger?.LogError(ex, "EvolutionService 数据库初始化失败");
            }
        }

        private async Task OnPointTransactionAsync(PointTransactionEvent ev)
        {
            // 只有正向收入才增加经验 (Income)
            if (ev.TransactionType != "Income" || ev.Amount <= 0) return;

            try
            {
                var userLevel = await GetOrCreateLevelAsync(ev.UserId);
                
                // 经验算法：1 积分 = 1 经验 (可调)
                long expGain = (long)ev.Amount;
                userLevel.Experience += expGain;
                userLevel.LastUpdateTime = DateTime.Now;

                // 检查是否升级
                int oldLevel = userLevel.Level;
                int newLevel = CalculateLevel(userLevel.Experience);

                bool isLevelUp = newLevel > oldLevel;
                bool isMedalAwarded = CheckAndAwardMedals(userLevel);

                if (isLevelUp)
                {
                    userLevel.Level = newLevel;
                    _logger?.LogInformation($"[进化] 用户 {ev.UserId} 升级了！ {oldLevel} -> {newLevel}");

                    // 触发升级事件
                    if (_robot != null)
                    {
                        await _robot.Events.PublishAsync(new LevelUpEvent
                        {
                            UserId = ev.UserId,
                            OldLevel = oldLevel,
                            NewLevel = newLevel,
                            RankName = GetRankName(newLevel)
                        });

                        // 升级奖励：赠送等级*100的积分
                        await _robot.CallSkillAsync("points.transfer", null!, new string[] { 
                            ev.UserId, 
                            "SYSTEM_RESERVE", 
                            (newLevel * 100).ToString(), 
                            $"等级提升至 Lv.{newLevel} 奖励" 
                        });
                    }
                }

                if (isLevelUp || isMedalAwarded)
                {
                    await userLevel.UpdateAsync();
                }
            }
            catch (Exception ex)
            {
                _logger?.LogError(ex, "处理经验增长时发生异常");
            }
        }

        private async Task OnSystemInteractionAsync(SystemInteractionEvent ev)
        {
            if (ev.InteractionType == "OpenMenu")
            {
                var userLevel = await GetOrCreateLevelAsync(ev.UserId);
                
                // 检查是否已获得过新手勋章
                var medals = (userLevel.Medals ?? "").Split(',', StringSplitOptions.RemoveEmptyEntries);
                if (!medals.Contains("⛵ 新手启航"))
                {
                    _logger?.LogInformation($"[任务] 用户 {ev.UserId} 完成了新手启航任务");
                    
                    // 奖励 50 经验
                    userLevel.Experience += 50;
                    CheckAndAwardMedals(userLevel);
                    await userLevel.UpdateAsync();

                    // 发送通知消息（如果可能）
                    if (_robot != null)
                    {
                        await _robot.SendMessageAsync("system", "bot", null, ev.UserId, 
                            "🎉 恭喜完成新手任务：【新手启航】！\n获得奖励：50 经验值 & ⛵ 新手启航勋章");
                    }
                }
            }
        }

        private bool CheckAndAwardMedals(UserLevel user)
        {
            var currentMedals = (user.Medals ?? "").Split(',', StringSplitOptions.RemoveEmptyEntries).ToList();
            bool changed = false;

            // 勋章规则
            var rules = new Dictionary<string, Func<UserLevel, bool>>
            {
                { "⛵ 新手启航", u => true }, // 只要触发 Check 就代表开启了旅程
                { "💰 第一桶金", u => u.Experience > 0 },
                { "🏅 崭露头角", u => u.Level >= 5 },
                { "🔥 矩阵精英", u => u.Level >= 10 },
                { "👑 进化主宰", u => u.Level >= 50 },
                { "💎 积分大亨", u => u.Experience >= 10000 }
            };

            foreach (var rule in rules)
            {
                if (!currentMedals.Contains(rule.Key) && rule.Value(user))
                {
                    currentMedals.Add(rule.Key);
                    changed = true;
                    _logger?.LogInformation($"[勋章] 用户 {user.UserId} 获得了勋章：{rule.Key}");
                }
            }

            if (changed)
            {
                user.Medals = string.Join(",", currentMedals);
            }
            return changed;
        }

        private int CalculateLevel(long exp)
        {
            // 简单等级算法：Level = sqrt(exp / 100)
            if (exp <= 0) return 1;
            return (int)Math.Floor(Math.Sqrt(exp / 100.0)) + 1;
        }

        private string GetRankName(int level)
        {
            if (level < 5) return "萌新机器人";
            if (level < 10) return "初级助理";
            if (level < 20) return "高级特工";
            if (level < 50) return "矩阵专家";
            return "进化终结者";
        }

        private async Task<UserLevel> GetOrCreateLevelAsync(string userId)
        {
            var level = await UserLevel.GetByUserIdAsync(userId);
            if (level == null)
            {
                level = new UserLevel
                {
                    UserId = userId,
                    Level = 1,
                    Experience = 0,
                    LastUpdateTime = DateTime.Now
                };
                await level.InsertAsync();
            }
            return level;
        }

        private async Task<string> HandleCommandAsync(IPluginContext ctx, string[] args)
        {
            var userLevel = await GetOrCreateLevelAsync(ctx.UserId);
            int nextLevel = userLevel.Level + 1;
            long nextLevelExp = (long)(Math.Pow(nextLevel - 1, 2) * 100);
            long needed = nextLevelExp - userLevel.Experience;
            
            var medals = string.IsNullOrEmpty(userLevel.Medals) ? "暂无勋章" : userLevel.Medals.Replace(",", "  ");

            return $"🆙 您的进化状态：\n" +
                   $"----------------\n" +
                   $"当前等级：Lv.{userLevel.Level} ({GetRankName(userLevel.Level)})\n" +
                   $"当前经验：{userLevel.Experience}\n" +
                   $"升级进度：{userLevel.Experience}/{nextLevelExp}\n" +
                   $"距离下级还需：{Math.Max(0, needed)} 经验\n" +
                   $"已获勋章：{medals}\n" +
                   $"----------------\n" +
                   $"提示：通过签到、完成任务获得积分可同步提升经验！";
        }
    }

    /// <summary>
    /// 用户等级数据模型
    /// </summary>
    public class UserLevel : MetaDataGuid<UserLevel>
    {
        public string UserId { get; set; } = string.Empty;

        public int Level { get; set; } = 1;

        public long Experience { get; set; } = 0;

        public string Medals { get; set; } = string.Empty; // 以逗号分隔的勋章列表

        public DateTime LastUpdateTime { get; set; } = DateTime.Now;

        public override string TableName => "UserLevels";
        public override string KeyField => "Id";

        public static async Task<UserLevel?> GetByUserIdAsync(string userId)
        {
            return await GetSingleAsync("WHERE UserId = @UserId", new { UserId = userId });
        }

        public static async Task<List<UserLevel>> GetTopRankingsAsync(int limit = 10)
        {
            return (await QueryAsync($"ORDER BY Experience DESC LIMIT {limit}")).ToList();
        }
    }
}
