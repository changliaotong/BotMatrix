using System;
using System.Threading.Tasks;
using BotWorker.Core.OneBot;

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
        Task<string> HandleMessageAsync(EventBase ev);
    }

    public interface IPluginContext
    {
        string Message { get; }
        string? GroupId { get; }
        string UserId { get; }
        string Platform { get; }
        string BotId { get; }
        string RawMessage { get; set; }
        Task ReplyAsync(string message);
    }
}
