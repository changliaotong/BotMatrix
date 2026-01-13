using System;
using System.Collections.Concurrent;
using System.Collections.Generic;
using System.Linq;
using System.Reflection;
using System.Text;
using System.Threading.Tasks;
using BotWorker.Domain.Interfaces;
using BotWorker.Domain.Models;
using Microsoft.Extensions.Logging;

namespace BotWorker.Modules.Games
{
    [BotPlugin(
        Id = "system.menu",
        Name = "超级多级菜单系统",
        Version = "1.1.0",
        Author = "Matrix",
        Description = "支持全系统插件自动发现、多级分类聚合、动态技能映射的智能菜单中心。",
        Category = "System"
    )]
    public class MenuService : IPlugin
    {
        private readonly ILogger<MenuService>? _logger;
        private static readonly ConcurrentDictionary<string, MenuSession> _sessions = new();
        private MenuNode _rootMenu = null!;
        private IRobot? _robot;

        public MenuService() { }
        public MenuService(ILogger<MenuService> logger) => _logger = logger;

        public List<Intent> Intents => [
            new() { Name = "主菜单", Keywords = ["菜单", "menu", "help", "帮助"] }
        ];

        public async Task InitAsync(IRobot robot)
        {
            _robot = robot;
            
            // 延迟一点初始化，确保其他插件都已加载完成
            _ = Task.Run(async () => {
                await Task.Delay(2000); 
                BuildDynamicMenuTree();
            });

            await robot.RegisterSkillAsync(new SkillCapability
            {
                Name = "菜单系统",
                Commands = ["菜单", "menu", "退出菜单", "刷新菜单"],
                Description = "输入【菜单】开启交互式导航；【刷新菜单】同步最新功能"
            }, HandleCommandAsync);
        }

        public Task StopAsync() => Task.CompletedTask;

        /// <summary>
        /// 核心：自动收集系统内所有系统模块并构建菜单
        /// </summary>
        private void BuildDynamicMenuTree()
        {
            if (_robot == null) return;

            var newRoot = new MenuNode
            {
                Id = "root",
                Title = "🤖 BotMatrix 系统主控中心",
                Description = "逻辑层已就绪，请选择需要交互的系统模块：",
                Children = new List<MenuNode>()
            };

            // 1. 获取所有系统模块元数据
            var pluginTypes = AppDomain.CurrentDomain.GetAssemblies()
                .SelectMany(s => s.GetTypes())
                .Where(p => typeof(IPlugin).IsAssignableFrom(p) && !p.IsInterface && !p.IsAbstract);

            var categoryGroups = new Dictionary<string, List<MenuNode>>();

            foreach (var type in pluginTypes)
            {
                var attr = type.GetCustomAttribute<BotPluginAttribute>();
                if (attr == null || attr.Id == "system.menu" || attr.Id == "matrix_market") continue;

                var category = attr.Category ?? "其他功能";
                if (!categoryGroups.ContainsKey(category))
                {
                    categoryGroups[category] = new List<MenuNode>();
                }

                // 为每个系统模块创建一个菜单项
                categoryGroups[category].Add(new MenuNode
                {
                    Id = attr.Id,
                    Title = attr.Name,
                    Description = attr.Description,
                    Type = MenuNodeType.Command,
                    ActionSkill = attr.Id 
                });
            }

            // 2. 将分类转换为二级菜单
            foreach (var group in categoryGroups)
            {
                var categoryNode = new MenuNode
                {
                    Id = $"cat_{group.Key}",
                    Title = GetCategoryIcon(group.Key) + " " + group.Key,
                    Description = $"包含 {group.Value.Count} 个逻辑子系统",
                    Type = MenuNodeType.Container,
                    Children = group.Value.Concat(new[] { 
                        new MenuNode { Id = "back", Title = "⬅️ 返回上一级", Type = MenuNodeType.Back } 
                    }).ToList()
                };
                newRoot.Children.Add(categoryNode);
            }

            // 3. 添加资源中心 (Matrix Market)
            newRoot.Children.Add(new MenuNode 
            { 
                Id = "market", 
                Title = "🌌 矩阵资源中心", 
                Description = "开启新系统、接入新逻辑、管理资源权限", 
                Type = MenuNodeType.Command,
                ActionSkill = "matrix_market"
            });

            // 4. 添加赛博团队 (Digital Staff)
            newRoot.Children.Add(new MenuNode 
            { 
                Id = "staff", 
                Title = "💼 赛博团队管理", 
                Description = "指挥您的数字员工进行自动化开发与销售", 
                Type = MenuNodeType.Command,
                ActionSkill = "core.digital_staff"
            });

            // 5. 添加荣耀榜单选项
            newRoot.Children.Add(new MenuNode 
            { 
                Id = "rankings", 
                Title = "🏆 荣耀榜单", 
                Description = "查看全服进化等级 Top 10", 
                Type = MenuNodeType.Command,
                ActionSkill = "menu.rankings"
            });

            // 5. 添加系统脉动 (Audit Log)
            newRoot.Children.Add(new MenuNode 
            { 
                Id = "monitor", 
                Title = "💓 系统脉动", 
                Description = "实时观察系统的事件流与审计日志", 
                Type = MenuNodeType.Command,
                ActionSkill = "menu.monitor"
            });

            // 6. 添加退出选项
            newRoot.Children.Add(new MenuNode { Id = "exit", Title = "🚪 退出系统", Type = MenuNodeType.Command, ActionSkill = "menu.exit" });

            _rootMenu = newRoot;
            _logger?.LogInformation($"系统逻辑同步完成，共接入 {categoryGroups.Count} 个分类。");
        }

        private string GetCategoryIcon(string category)
        {
            return category switch
            {
                "Games" or "游戏" => "🎮",
                "Financial" or "金融" or "积分" => "💰",
                "System" or "系统" => "⚙️",
                "Media" or "媒体" or "音乐" => "🎵",
                "Social" or "社交" => "💬",
                _ => "📦"
            };
        }

        private async Task<string> HandleCommandAsync(IPluginContext ctx, string[] args)
        {
            var text = ctx.RawMessage.Trim();
            
            if (text == "退出菜单")
            {
                _sessions.TryRemove(ctx.UserId, out _);
                return "已退出菜单模式。";
            }

            if (text == "刷新菜单")
            {
                BuildDynamicMenuTree();
                return "✅ 菜单树已实时重构，请输入【菜单】查看。";
            }

            var session = _sessions.GetOrAdd(ctx.UserId, id => {
                // 第一次进入菜单，触发系统交互事件
                if (_robot != null)
                {
                    _ = _robot.Events.PublishAsync(new SystemInteractionEvent
                    {
                        UserId = ctx.UserId,
                        InteractionType = "OpenMenu",
                        Details = "用户首次开启超级菜单"
                    });
                }
                return new MenuSession { UserId = id, Path = new List<string> { "root" } };
            });
            session.LastActiveTime = DateTime.Now;

            if (session.CurrentQuestionIndex >= 0)
            {
                return await HandleConversationAsync(ctx, session, text);
            }

            if (int.TryParse(text, out int choice))
            {
                return await HandleMenuChoiceAsync(ctx, session, choice);
            }

            return await RenderMenuAsync(session);
        }

        private async Task<string> HandleMenuChoiceAsync(IPluginContext ctx, MenuSession session, int choice)
        {
            var currentMenu = FindNodeById(_rootMenu, session.CurrentMenuId);
            if (currentMenu == null || choice < 1 || choice > currentMenu.Children.Count)
            {
                return "❌ 无效的选择，请重新输入数字。";
            }

            var selected = currentMenu.Children[choice - 1];
            
            switch (selected.Type)
            {
                case MenuNodeType.Container:
                    session.Path.Add(selected.Id);
                    return await RenderMenuAsync(session);

                case MenuNodeType.Back:
                    if (session.Path.Count > 1) session.Path.RemoveAt(session.Path.Count - 1);
                    return await RenderMenuAsync(session);

                case MenuNodeType.Command:
                    if (selected.Id == "exit")
                    {
                        _sessions.TryRemove(ctx.UserId, out _);
                        return "👋 感谢使用，再见！";
                    }
                    if (selected.Id == "rankings")
                    {
                        return await GetRankingsDisplayAsync();
                    }
                    if (selected.Id == "monitor")
                    {
                        return GetMonitorDisplay();
                    }
                    if (selected.Id == "market")
                    {
                        // 映射到 MatrixMarketService 的指令
                        return await _robot!.CallSkillAsync("matrix_market", ctx, Array.Empty<string>()) as string ?? "❌ 资源中心暂时无法连接";
                    }

                    // 检查是否是需要激活的系统
                    if (selected.Id.StartsWith("game."))
                    {
                        var access = await UserModuleAccess.QueryWhere("UserId = @p1 AND ModuleId = @p2", UserModuleAccess.SqlParams(("@p1", ctx.UserId), ("@p2", selected.Id)));
                        if (!access.Any())
                        {
                            return $"🔒 访问受限：系统检测到您尚未接入“{selected.Title}”。\n\n💡 请前往【🌌 矩阵资源中心】获取接入权限。";
                        }
                    }

                    return $"🚀 正在为您启动：{selected.Title}...\n(描述: {selected.Description})\n\n💡 请直接输入该功能的指令。";

                case MenuNodeType.Input:
                    session.Path.Add(selected.Id);
                    session.CurrentQuestionIndex = 0;
                    session.CollectedData.Clear();
                    return $"📝 开始【{selected.Title}】流程：\n\n1. {selected.Questions![0]}";

                default:
                    return "未知节点类型";
            }
        }

        private async Task<string> HandleConversationAsync(IPluginContext ctx, MenuSession session, string input)
        {
            var currentMenu = FindNodeById(_rootMenu, session.CurrentMenuId);
            var questions = currentMenu?.Questions;
            
            if (questions == null) return "对话配置错误";

            session.CollectedData[questions[session.CurrentQuestionIndex]] = input;
            session.CurrentQuestionIndex++;

            if (session.CurrentQuestionIndex < questions.Count)
            {
                return $"{session.CurrentQuestionIndex + 1}. {questions[session.CurrentQuestionIndex]}";
            }

            var sb = new StringBuilder();
            sb.AppendLine("✅ 采集完成！数据如下：");
            foreach (var kv in session.CollectedData)
            {
                sb.AppendLine($" - {kv.Key}: {kv.Value}");
            }
            
            session.CurrentQuestionIndex = -1;
            session.Path.RemoveAt(session.Path.Count - 1);
            
            sb.AppendLine("\n" + await RenderMenuAsync(session));
            return sb.ToString();
        }

        private async Task<string> RenderMenuAsync(MenuSession session)
        {
            var node = FindNodeById(_rootMenu, session.CurrentMenuId);
            if (node == null) return "❌ 菜单节点丢失，请尝试回复【刷新菜单】。";

            var sb = new StringBuilder();
            
            // 头部：标题与用户信息
            sb.AppendLine($"┏━━ {node.Title} ━━┓");
            
            if (session.CurrentMenuId == "root")
            {
                var userLevel = await UserLevel.GetByUserIdAsync(session.UserId);
                string plane = "原质";
                int level = 1;
                if (userLevel != null)
                {
                    level = userLevel.Level;
                    plane = GetPlaneName(level);
                }

                // 尝试获取用户积分
                long credit = await UserInfo.GetCreditAsync(long.Parse(session.UserId));

                sb.AppendLine($"┃ 👤 账户: {session.UserId}");
                sb.AppendLine($"┃ 🆙 等级: Lv.{level} ({plane})");
                sb.AppendLine($"┃ 💰 积分: {credit:N0}");

                // 展示活跃的全局 Buff
                double expBuff = _robot?.Events.GetActiveBuff(BuffType.ExperienceMultiplier) ?? 1.0;
                double pointsBuff = _robot?.Events.GetActiveBuff(BuffType.PointsMultiplier) ?? 1.0;
                if (expBuff > 1.0 || pointsBuff > 1.0)
                {
                    sb.AppendLine("┃ ━━━━━━━━━━━━━━━━━━");
                    if (expBuff > 1.0) sb.AppendLine($"┃ 🔥 经验加成: {expBuff}x");
                    if (pointsBuff > 1.0) sb.AppendLine($"┃ � 积分加成: {pointsBuff}x");
                }
            }

            sb.AppendLine("┃ ━━━━━━━━━━━━━━━━━━");
            sb.AppendLine($"┃ 📝 {node.Description}");
            sb.AppendLine("┃");
            
            var userAccess = await UserModuleAccess.QueryWhere("UserId = @p1", UserModuleAccess.SqlParams(("@p1", session.UserId)));
            var unlockedIds = userAccess.Select(a => a.ModuleId).ToHashSet();

            for (int i = 0; i < node.Children.Count; i++)
            {
                var child = node.Children[i];
                var icon = child.Type switch {
                    MenuNodeType.Container => "�",
                    MenuNodeType.Command => "▶️",
                    MenuNodeType.Input => "💬",
                    MenuNodeType.Back => "🔙",
                    _ => "🔹"
                };

                string title = child.Title;
                if (child.Id.StartsWith("game.") && !unlockedIds.Contains(child.Id))
                {
                    title = "🔒 " + title;
                }

                sb.AppendLine($"┃  {i + 1}. {icon} {title}");
            }

            sb.AppendLine("┃");
            sb.AppendLine("┃ 💡 回复数字选择 | 退出菜单");
            sb.AppendLine("┗━━━━━━━━━━━━━━━━━━┛");

            return sb.ToString();
        }

        private string GetPlaneName(int level)
        {
            if (level < 10) return "⚪ 原质";
            if (level < 30) return "🟢 构件";
            if (level < 60) return "🔵 逻辑";
            if (level < 90) return "🟣 协议";
            if (level < 120) return "🟡 矩阵";
            return "🔴 奇点";
        }

        private async Task<string> GetRankingsDisplayAsync()
        {
            var topList = await UserLevel.GetTopRankingsAsync(10);
            var sb = new StringBuilder();
            sb.AppendLine("🏆 【BotMatrix 进化荣耀榜】 🏆");
            sb.AppendLine("━━━━━━━━━━━━━━━━━━━━━━━━━━━━");
            
            if (topList.Count == 0)
            {
                sb.AppendLine("  暂无排名数据，快去进化吧！");
            }
            else
            {
                for (int i = 0; i < topList.Count; i++)
                {
                    var user = topList[i];
                    string medal = i switch { 0 => "🥇", 1 => "🥈", 2 => "🥉", _ => $" {i + 1}. " };
                    sb.AppendLine($"{medal} {user.UserId.PadRight(12)} Lv.{user.Level}");
                }
            }
            
            sb.AppendLine("━━━━━━━━━━━━━━━━━━━━━━━━━━━━");
            sb.AppendLine("💡 回复任意数字返回主菜单");
            return sb.ToString();
        }

        private string GetMonitorDisplay()
        {
            if (_robot == null) return "❌ 系统未就绪";
            
            var audits = _robot.Events.GetRecentAudits();
            var sb = new StringBuilder();
            sb.AppendLine("💓 【BotMatrix 系统脉动监控】 💓");
            sb.AppendLine("━━━━━━━━━━━━━━━━━━━━━━━━━━━━");
            
            if (audits.Count == 0)
            {
                sb.AppendLine("  [静默] 系统目前运行平稳，无关键事件。");
            }
            else
            {
                foreach (var log in audits.Take(15)) // 只显示最近 15 条
                {
                    string icon = log.Level switch {
                        "Success" => "✅",
                        "Warning" => "⚠️",
                        "Critical" => "🚨",
                        _ => "ℹ️"
                    };
                    sb.AppendLine($"{icon} [{log.Timestamp:HH:mm:ss}] {log.Message}");
                }
            }
            
            sb.AppendLine("━━━━━━━━━━━━━━━━━━━━━━━━━━━━");
            sb.AppendLine("💡 自动追踪最新 50 条关键审计日志");
            sb.AppendLine("💡 回复任意数字返回主菜单");
            return sb.ToString();
        }

        private MenuNode? FindNodeById(MenuNode root, string id)
        {
            if (root.Id == id) return root;
            if (root.Children == null) return null;
            foreach (var child in root.Children)
            {
                var found = FindNodeById(child, id);
                if (found != null) return found;
            }
            return null;
        }
    }
}
