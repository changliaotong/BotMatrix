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
            robot.Events?.Subscribe<PointTransactionEvent>(OnPointTransactionAsync);
            
            // 订阅系统交互事件，实现新手引导任务
            robot.Events?.Subscribe<SystemInteractionEvent>(OnSystemInteractionAsync);

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
            await UserLevel.EnsureTableCreatedAsync();
        }

        private async Task OnPointTransactionAsync(PointTransactionEvent ev)
        {
            try
            {
                var userLevel = await GetOrCreateLevelAsync(ev.UserId);
                
                // 天才数值模型：收入 0.8 倍经验，支出 1.2 倍经验 (鼓励流动)
                double weight = ev.TransactionType == "Income" ? 0.8 : 1.2;
                
                // 应用全局 Buff 加成
                double globalBuff = _robot?.Events.GetActiveBuff(BuffType.ExperienceMultiplier) ?? 1.0;
                
                long expGain = (long)(Math.Abs(ev.Amount) * (decimal)weight * (decimal)globalBuff);
                
                if (expGain <= 0) return;

                userLevel.Experience += expGain;
                userLevel.LastUpdateTime = DateTime.Now;
                
                // 详细日志记录
                string buffInfo = globalBuff > 1.0 ? $" (含 {globalBuff}x 全服加成)" : "";
                _logger?.LogInformation($"[进化] 用户 {ev.UserId} 产生行为({ev.TransactionType})，获得经验: {expGain}{buffInfo}");

                // 检查是否升级
                int oldLevel = userLevel.Level;
                int newLevel = CalculateLevel(userLevel.Experience);
                
                bool medalsChanged = CheckAndAwardMedals(userLevel);

                if (newLevel > oldLevel || medalsChanged)
                {
                    userLevel.Level = newLevel;
                    await userLevel.UpdateAsync();
                    
                    if (newLevel > oldLevel)
                    {
                        _logger?.LogInformation($"[进化] 用户 {ev.UserId} 升级至 Lv.{newLevel} ({GetRankName(newLevel)})");
                        
                        // 发布审计事件
                        if (_robot != null)
                        {
                            await _robot.Events.PublishAsync(new SystemAuditEvent {
                                Level = "Success",
                                Source = "Evolution",
                                Message = $"用户 {ev.UserId} 晋升位面: {GetRankName(newLevel)} (Lv.{newLevel})",
                                TargetUser = ev.UserId
                            });

                            // 发送升级通知
                            await _robot.SendMessageAsync("system", "bot", null, ev.UserId, 
                                $"🎊 恭喜！您已进化至位面：{GetRankName(newLevel)} (Lv.{newLevel})！\n解锁了更多系统特权，请前往超级菜单查看。");
                        }
                    }
                }
                else
                {
                    await userLevel.UpdateAsync();
                }
            }
            catch (Exception ex)
            {
                _logger?.LogError(ex, $"处理积分经验转化失败: {ev.UserId}");
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
            if (exp <= 0) return 1;
            // 对应公式 Exp = 50L^2 + 150L - 200
            // 反函数 L = (-150 + sqrt(22500 + 200 * (200 + exp))) / 100
            double l = (-150.0 + Math.Sqrt(22500.0 + 200.0 * (200.0 + exp))) / 100.0;
            return (int)Math.Max(1, Math.Floor(l));
        }

        private string GetRankName(int level)
        {
            if (level < 10) return "原质";
            if (level < 30) return "构件";
            if (level < 60) return "逻辑";
            if (level < 90) return "协议";
            if (level < 120) return "矩阵";
            return "奇点";
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
            return (await QueryWhere("UserId = @p1", SqlParams(("@p1", userId)))).FirstOrDefault();
        }

        public static async Task<List<UserLevel>> GetTopRankingsAsync(int limit = 10)
        {
            return await GetListAsync($"SELECT TOP {limit} * FROM {FullName} ORDER BY Experience DESC");
        }
    }
}
