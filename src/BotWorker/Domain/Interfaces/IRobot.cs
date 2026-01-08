using System;
using System.Threading.Tasks;
using BotWorker.Infrastructure.Communication.OneBot; // Assuming EventBase moved here

namespace BotWorker.Domain.Interfaces
{
    public interface IRobot
    {
        Task RegisterSkillAsync(object capability, Func<IPluginContext, string[], Task<string>> handler);
        Task RegisterEventAsync(string eventType, Func<IPluginContext, Task> handler);
        Task<string> HandleEventAsync(EventBase ev, Func<string, Task>? replyDelegate = null);
    }
}


