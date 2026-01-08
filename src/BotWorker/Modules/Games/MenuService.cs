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
        /// 核心：自动收集系统内所有插件并构建菜单
        /// </summary>
        private void BuildDynamicMenuTree()
        {
            if (_robot == null) return;

            var newRoot = new MenuNode
            {
                Id = "root",
                Title = "🤖 智能机器人主控中心 (自动发现版)",
                Description = "系统已自动扫描并聚合所有功能，请选择：",
                Children = new List<MenuNode>()
            };

            // 1. 获取所有插件元数据
            // 注意：这里假设 IRobot 提供了获取已加载插件列表的能力
            // 如果接口受限，我们通过反射当前程序集获取所有 IPlugin 实现
            var pluginTypes = AppDomain.CurrentDomain.GetAssemblies()
                .SelectMany(s => s.GetTypes())
                .Where(p => typeof(IPlugin).IsAssignableFrom(p) && !p.IsInterface && !p.IsAbstract);

            var categoryGroups = new Dictionary<string, List<MenuNode>>();

            foreach (var type in pluginTypes)
            {
                var attr = type.GetCustomAttribute<BotPluginAttribute>();
                if (attr == null || attr.Id == "system.menu") continue;

                var category = attr.Category ?? "其他功能";
                if (!categoryGroups.ContainsKey(category))
                {
                    categoryGroups[category] = new List<MenuNode>();
                }

                // 为每个插件创建一个菜单项
                categoryGroups[category].Add(new MenuNode
                {
                    Id = attr.Id,
                    Title = attr.Name,
                    Description = attr.Description,
                    Type = MenuNodeType.Command,
                    ActionSkill = attr.Id // 约定：动作技能 ID 与插件 ID 一致
                });
            }

            // 2. 将分类转换为二级菜单
            foreach (var group in categoryGroups)
            {
                var categoryNode = new MenuNode
                {
                    Id = $"cat_{group.Key}",
                    Title = GetCategoryIcon(group.Key) + " " + group.Key,
                    Description = $"包含 {group.Value.Count} 个相关功能",
                    Type = MenuNodeType.Container,
                    Children = group.Value.Concat(new[] { 
                        new MenuNode { Id = "back", Title = "⬅️ 返回上一级", Type = MenuNodeType.Back } 
                    }).ToList()
                };
                newRoot.Children.Add(categoryNode);
            }

            // 3. 添加荣耀榜单选项
            newRoot.Children.Add(new MenuNode 
            { 
                Id = "rankings", 
                Title = "🏆 荣耀榜单", 
                Description = "查看全服进化等级 Top 10", 
                Type = MenuNodeType.Command,
                ActionSkill = "menu.rankings"
            });

            // 4. 添加退出选项
            newRoot.Children.Add(new MenuNode { Id = "exit", Title = "🚪 退出系统", Type = MenuNodeType.Command, ActionSkill = "menu.exit" });

            _rootMenu = newRoot;
            _logger?.LogInformation($"菜单系统已完成自动发现，共聚合了 {categoryGroups.Count} 个分类。");
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

            return RenderMenu(session);
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
                    return RenderMenu(session);

                case MenuNodeType.Back:
                    if (session.Path.Count > 1) session.Path.RemoveAt(session.Path.Count - 1);
                    return RenderMenu(session);

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
            
            sb.AppendLine("\n" + RenderMenu(session));
            return sb.ToString();
        }

        private string RenderMenu(MenuSession session)
        {
            var node = FindNodeById(_rootMenu, session.CurrentMenuId);
            if (node == null) return "❌ 菜单节点丢失，请尝试回复【刷新菜单】。";

            var sb = new StringBuilder();
            sb.AppendLine("╔════════════════════════════╗");
            sb.AppendLine($"║  {node.Title.PadRight(24)}║");
            
            if (session.CurrentMenuId == "root")
            {
                sb.AppendLine("╟────────────────────────────╢");
                sb.AppendLine($"║ 👤 用户: {session.UserId.PadRight(18)}║");
            }

            sb.AppendLine("╟────────────────────────────╢");
            sb.AppendLine($"║ 📝 {node.Description.PadRight(24)}║");
            sb.AppendLine("║                            ║");
            
            for (int i = 0; i < node.Children.Count; i++)
            {
                var child = node.Children[i];
                var icon = child.Type switch {
                    MenuNodeType.Container => "📁",
                    MenuNodeType.Command => "⚡",
                    MenuNodeType.Input => "⌨️",
                    MenuNodeType.Back => "🔙",
                    _ => "🔹"
                };
                var line = $" {i + 1}. {icon} {child.Title}";
                sb.AppendLine($"║ {line.PadRight(25)}║");
            }

            sb.AppendLine("║                            ║");
            sb.AppendLine("║ 💡 输入数字选择 | 退出菜单 ║");
            sb.AppendLine("╚════════════════════════════╝");

            return sb.ToString();
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
