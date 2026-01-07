using System;
using System.Threading.Tasks;
using BotWorker.Core.OneBot;
using BotWorker.Services;
using BotWorker.Bots.Entries;
using BotWorker.Bots.Groups;

namespace BotWorker.Core.Plugin
{
    public interface IPlugin
    {
        string Name { get; }
        string Description { get; }
        Task InitAsync(IRobot robot);
    }

    public interface IRobot
    {
        Task RegisterSkillAsync(SkillCapability capability, Func<IPluginContext, string[], Task<string>> handler);
        Task RegisterEventAsync(string eventType, Func<IPluginContext, Task> handler);
        Task<string> HandleEventAsync(EventBase ev, Func<string, Task>? replyDelegate = null);
    }

    public interface IPluginContext
    {
        // 基础信息
        string Message { get; }
        string? GroupId { get; }
        string UserId { get; }
        string Platform { get; }
        string BotId { get; }
        string EventType { get; }
        bool IsMessage { get; }
        string RawMessage { get; set; }

        // 丰富实体
        UserInfo? User { get; }
        GroupInfo? Group { get; }
        GroupMember? Member { get; }
        BotInfo? Bot { get; }

        // 服务接口
        IAIService AI { get; }
        II18nService I18n { get; }

        // 动作
        Task ReplyAsync(string message);
    }
}
