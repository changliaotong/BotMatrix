using System;
using System.Collections.Generic;
using System.Collections.Concurrent;
using System.Threading.Tasks;
using BotWorker.Infrastructure.Communication.OneBot;
using BotWorker.Application.Services;
using BotWorker.Services;
using BotWorker.Domain.Entities;
using BotWorker.Domain.Interfaces;
using Microsoft.Extensions.Logging;

namespace BotWorker.Modules.Plugins
{
    public class SkillCapability
    {
        public string Name { get; set; } = string.Empty;
        public string Description { get; set; } = string.Empty;
        public string[] Commands { get; set; } = Array.Empty<string>();
        public List<BotWorker.Domain.Interfaces.Intent> Intents { get; set; } = new();
        public string Usage { get; set; } = string.Empty;

        public SkillCapability() { }
        public SkillCapability(string name, string[] commands, string description = "")
        {
            Name = name;
            Commands = commands;
            Description = description;
        }
    }

    public class PluginContext : IPluginContext
    {
        private readonly EventBase _ev;
        public EventBase Event => _ev;
        private readonly Func<string, Task>? _replyDelegate;
        private readonly Func<string, string, string, string, string, Task>? _musicReplyDelegate;
        private readonly ConcurrentDictionary<string, object?> _state = new();

        // 基础信息
        public string Message => _ev.RawMessage;
        public string? GroupId => _ev.GroupId;
        public string UserId => _ev.UserId;
        public string UserName => User?.Name ?? string.Empty;
        public string Platform { get; }
        public string BotId { get; }
        public string EventType => _ev.PostType; 
        public bool IsMessage => _ev.PostType == "message";
        public string RawMessage { get; set; } = string.Empty;
        public string? GroupName => Group?.GroupName;

        // 提及解析
        private List<MentionedUser>? _mentionedUsers;
        public List<MentionedUser> MentionedUsers
        {
            get
            {
                if (_mentionedUsers == null)
                {
                    _mentionedUsers = new List<MentionedUser>();
                    // 解析 CQ 码中的 at: [CQ:at,qq=123456]
                    var matches = System.Text.RegularExpressions.Regex.Matches(RawMessage, @"\[CQ:at,qq=(\d+)\]");
                    foreach (System.Text.RegularExpressions.Match match in matches)
                    {
                        var userId = match.Groups[1].Value;
                        _mentionedUsers.Add(new MentionedUser { UserId = userId, Name = "" }); // 名字可能需要后续从缓存或 API 获取
                    }
                }
                return _mentionedUsers;
            }
        }

        // 丰富实体
        public UserInfo? User { get; }
        public GroupInfo? Group { get; }
        public GroupMember? Member { get; }
        public BotInfo? Bot { get; }

        // 服务接口
        public IAIService AI { get; }
        public II18nService I18n { get; }
        public ILogger Logger { get; }

        public PluginContext(
            EventBase ev, 
            string platform, 
            string botId,
            IAIService ai,
            II18nService i18n,
            ILogger logger,
            UserInfo? user = null,
            GroupInfo? group = null,
            GroupMember? member = null,
            BotInfo? bot = null,
            Func<string, Task>? replyDelegate = null,
            Func<string, string, string, string, string, Task>? musicReplyDelegate = null)
        {
            _ev = ev;
            Platform = platform;
            BotId = botId;
            RawMessage = ev.RawMessage;

            AI = ai;
            I18n = i18n;
            Logger = logger;
            User = user;
            Group = group;
            Member = member;
            Bot = bot;
            _replyDelegate = replyDelegate;
            _musicReplyDelegate = musicReplyDelegate;
        }

        public async Task ReplyAsync(string message)
        {
            if (_replyDelegate != null)
            {
                await _replyDelegate(message);
            }
            else
            {
                Logger.LogWarning($"[PluginContext] Reply to {UserId} on {Platform} failed: No reply delegate provided. Message: {message}");
            }
        }

        public async Task SendMusicAsync(string title, string artist, string jumpUrl, string coverUrl, string audioUrl)
        {
            if (_musicReplyDelegate != null)
            {
                await _musicReplyDelegate(title, artist, jumpUrl, coverUrl, audioUrl);
            }
            else
            {
                // 如果不支持音乐消息，退而求其次发送链接
                await ReplyAsync($"🎵 {title} - {artist}\n🔗 {audioUrl}");
            }
        }

        public void SetState<T>(string key, T value) => _state[key] = value;

        public T? GetState<T>(string key) => _state.TryGetValue(key, out var value) && value is T t ? t : default;

        // 会话支持
        public bool IsConfirmed { get; set; }
        public string? SessionAction { get; set; }
        public string? SessionStep { get; set; }
        public object? SessionData { get; set; }
    }

    public class Skill
    {
        public string PluginId { get; set; } = string.Empty;
        public SkillCapability Capability { get; set; } = new();
        public Func<IPluginContext, string[], Task<string>> Handler { get; set; } = (_, _) => Task.FromResult(string.Empty);

        public Skill() { }
        public Skill(SkillCapability capability, Func<IPluginContext, string[], Task<string>> handler)
        {
            Capability = capability;
            Handler = handler;
        }
    }

    // --- 插件配置模型 (plugin.json) ---

    public class PluginConfig : IModuleMetadata
    {
        public string Id { get; set; } = string.Empty;
        public string Name { get; set; } = string.Empty;
        public string Version { get; set; } = "1.0.0";
        public string Author { get; set; } = "Unknown";
        public string Description { get; set; } = string.Empty;
        public string Category { get; set; } = "General";
        public string[] Permissions { get; set; } = Array.Empty<string>();
        public string[] Dependencies { get; set; } = Array.Empty<string>();
        public bool IsEssential { get; set; } = false;
        
        // Go 兼容字段
        public List<Intent> Intents { get; set; } = new();
        public List<UIComponent> UI { get; set; } = new();
        public string[] Events { get; set; } = Array.Empty<string>();

        // 运行配置
        public string? Type { get; set; }          // native, process, remote
        public string? Executable { get; set; }    // 仅 process 类型使用
        public string? Endpoint { get; set; }      // 仅 remote 类型使用
        public string? EntryPoint { get; set; }    // 仅 native 类型使用 (DLL 路径)
        public string[]? Actions { get; set; }     // 兼容旧字段
    }
}


