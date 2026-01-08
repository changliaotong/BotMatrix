using System;
using System.Collections.Generic;
using System.Linq;
using System.Threading.Tasks;
using BotWorker.Infrastructure.Communication.OneBot;
using BotWorker.Application.Services;
using BotWorker.Domain.Entities;
using BotWorker.Domain.Interfaces;

namespace BotWorker.Modules.Plugins
{
    public class PluginManager : IRobot
    {
        private readonly List<Skill> _skills = new();
        private readonly Dictionary<string, List<Func<IPluginContext, Task>>> _eventHandlers = new();
        private readonly List<IPlugin> _plugins = new();
        private readonly IAIService _aiService;
        private readonly II18nService _i18nService;

        public IReadOnlyList<Skill> Skills => _skills;
        public IReadOnlyList<IPlugin> Plugins => _plugins;

        public IAIService AI => _aiService;
        public II18nService I18n => _i18nService;

        public PluginManager(IAIService aiService, II18nService i18nService)
        {
            _aiService = aiService;
            _i18nService = i18nService;
        }

        public Task RegisterSkillAsync(SkillCapability capability, Func<IPluginContext, string[], Task<string>> handler)
        {
            _skills.Add(new Skill { Capability = capability, Handler = handler });
            return Task.CompletedTask;
        }

        public Task RegisterEventAsync(string eventType, Func<IPluginContext, Task> handler)
        {
            if (!_eventHandlers.ContainsKey(eventType))
            {
                _eventHandlers[eventType] = new List<Func<IPluginContext, Task>>();
            }
            _eventHandlers[eventType].Add(handler);
            return Task.CompletedTask;
        }

        public async Task LoadPluginAsync(IPlugin plugin)
        {
            _plugins.Add(plugin);
            await plugin.InitAsync(this);
        }

        public async Task<string> HandleEventAsync(EventBase ev, Func<string, Task>? replyDelegate = null)
        {
            // 1. 异步加载上下文数�?
            var userIdStr = ev.UserId;
            var groupIdStr = ev.GroupId;
            
            long userId = 0;
            long.TryParse(userIdStr, out userId);
            
            long groupId = 0;
            if (!string.IsNullOrEmpty(groupIdStr)) long.TryParse(groupIdStr, out groupId);
            
            var botId = ev.SelfId;

            var userTask = userId != 0 ? UserInfo.GetSingleAsync(userId) : Task.FromResult<UserInfo?>(null);
            var botTask = BotInfo.GetSingleAsync(botId);
            var groupTask = groupId != 0 ? GroupInfo.GetSingleAsync(groupId) : Task.FromResult<GroupInfo?>(null);
            var memberTask = (groupId != 0 && userId != 0) ? GroupMember.GetSingleAsync(groupId, userId) : Task.FromResult<GroupMember?>(null);

            await Task.WhenAll(userTask, botTask, groupTask, memberTask);

            var ctx = new PluginContext(
                ev, 
                "onebot", 
                botId.ToString(),
                _aiService,
                _i18nService,
                await userTask,
                await groupTask,
                await memberTask,
                await botTask,
                replyDelegate)
            {
                RawMessage = ev.RawMessage
            };

            return await DispatchAsync(ctx);
        }

        public async Task<string> DispatchAsync(IPluginContext ctx)
        {
            // 1. 处理通用事件分发
            if (_eventHandlers.TryGetValue(ctx.EventType, out var handlers))
            {
                foreach (var handler in handlers)
                {
                    try { await handler(ctx); } catch { /* 忽略插件错误 */ }
                }
            }

            // 2. 处理消息指令 (仅限 PostType �?message 的情�?
            if (ctx.IsMessage)
            {
                var message = ctx.RawMessage.Trim();
                if (string.IsNullOrEmpty(message)) return string.Empty;

                foreach (var skill in _skills)
                {
                    foreach (var cmd in skill.Capability.Commands)
                    {
                        if (message.StartsWith(cmd, StringComparison.OrdinalIgnoreCase))
                        {
                            var args = message.Substring(cmd.Length).Trim().Split(new[] { ' ' }, StringSplitOptions.RemoveEmptyEntries);
                            return await skill.Handler(ctx, args);
                        }
                    }
                }
            }

            return string.Empty;
        }
    }
}


