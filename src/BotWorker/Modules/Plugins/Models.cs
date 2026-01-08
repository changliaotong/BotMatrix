using System;
using System.Collections.Generic;
using System.Threading.Tasks;
using BotWorker.Infrastructure.Communication.OneBot;
using BotWorker.Application.Services;
using BotWorker.Services;
using BotWorker.Domain.Entities;
using BotWorker.Domain.Interfaces;

namespace BotWorker.Modules.Plugins
{
    public class SkillCapability
    {
        public string Name { get; set; } = string.Empty;
        public string Description { get; set; } = string.Empty;
        public string[] Commands { get; set; } = Array.Empty<string>();
        public string Usage { get; set; } = string.Empty;
    }

    public class PluginContext : IPluginContext
    {
        private readonly EventBase _ev;
        public EventBase Event => _ev;
        private readonly Func<string, Task>? _replyDelegate;

        // 基础信息
        public string Message => _ev.RawMessage;
        public string? GroupId => _ev.GroupId;
        public string UserId => _ev.UserId;
        public string Platform { get; }
        public string BotId { get; }
        public string EventType => _ev.PostType; // 或者是更具体的事件类型
        public bool IsMessage => _ev.PostType == "message";
        public string RawMessage { get; set; } = string.Empty;

        // 丰富实体
        public UserInfo? User { get; }
        public GroupInfo? Group { get; }
        public GroupMember? Member { get; }
        public BotInfo? Bot { get; }

        // 服务接口
        public IAIService AI { get; }
        public II18nService I18n { get; }

        public PluginContext(
            EventBase ev, 
            string platform, 
            string botId,
            IAIService ai,
            II18nService i18n,
            UserInfo? user = null,
            GroupInfo? group = null,
            GroupMember? member = null,
            BotInfo? bot = null,
            Func<string, Task>? replyDelegate = null)
        {
            _ev = ev;
            Platform = platform;
            BotId = botId;
            RawMessage = ev.RawMessage;

            AI = ai;
            I18n = i18n;
            User = user;
            Group = group;
            Member = member;
            Bot = bot;
            _replyDelegate = replyDelegate;
        }

        public async Task ReplyAsync(string message)
        {            if (_replyDelegate != null)
            {
                await _replyDelegate(message);
            }
            else
            {
                // 回退逻辑，如果没提供委托则记录日�?                Console.WriteLine($"[PluginContext] Reply to {UserId} on {Platform}: {message}");
            }
        }
    }

    public class Skill
    {
        public SkillCapability Capability { get; set; } = new();
        public Func<IPluginContext, string[], Task<string>> Handler { get; set; } = (_, _) => Task.FromResult(string.Empty);
    }
}


