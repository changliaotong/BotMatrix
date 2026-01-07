using System;
using System.Collections.Generic;
using System.Threading.Tasks;
using BotWorker.Core.OneBot;

namespace BotWorker.Core.Plugin
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
        public string Message { get; set; } = string.Empty;
        public string? GroupId { get; set; }
        public string UserId { get; set; } = string.Empty;
        public string Platform { get; set; } = string.Empty;
        public string BotId { get; set; } = string.Empty;
        public string RawMessage { get; set; } = string.Empty;

        public PluginContext(EventBase ev, string platform, string botId)
        {
            Message = ev.RawMessage;
            UserId = ev.UserId;
            GroupId = ev.GroupId;
            Platform = platform;
            BotId = botId;
        }

        public Task ReplyAsync(string message)
        {
            // TODO: 实现发送逻辑
            Console.WriteLine($"Reply to {UserId} in {GroupId}: {message}");
            return Task.CompletedTask;
        }
    }

    public class Skill
    {
        public SkillCapability Capability { get; set; } = new();
        public Func<IPluginContext, string[], Task<string>> Handler { get; set; } = (_, _) => Task.FromResult(string.Empty);
    }
}
