using System;
using System.Threading.Tasks;
using BotWorker.Infrastructure.Communication.OneBot;
using BotWorker.Modules.Plugins;

namespace BotWorker.Domain.Interfaces
{
    public interface IRobot
    {
        Task RegisterSkillAsync(SkillCapability capability, Func<IPluginContext, string[], Task<string>> handler);
        Task RegisterEventAsync(string eventType, Func<IPluginContext, Task> handler);
        Task<string> HandleEventAsync(EventBase ev, Func<string, Task>? replyDelegate = null);
    }
}


