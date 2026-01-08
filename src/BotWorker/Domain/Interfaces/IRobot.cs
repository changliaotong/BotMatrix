using System;
using System.Threading.Tasks;
using BotWorker.Infrastructure.Communication.OneBot;
using BotWorker.Modules.Plugins;

namespace BotWorker.Domain.Interfaces
{
    public interface IRobot
    {
        Task RegisterSkillAsync(SkillCapability capability, Func<IPluginContext, string[], Task<string>> handler);
        
        // 为了兼容旧插件的同步调用
        void RegisterSkill(SkillCapability capability, Func<IPluginContext, string[], Task<string>> handler)
            => RegisterSkillAsync(capability, handler).GetAwaiter().GetResult();

        Task RegisterEventAsync(string eventType, Func<IPluginContext, Task> handler);
        Task<string> HandleEventAsync(EventBase ev, Func<string, Task>? replyDelegate = null);
        
        /// <summary>
        /// 发送消息（支持主动发送）
        /// </summary>
        Task SendMessageAsync(string platform, string botId, string? groupId, string userId, string message);

        /// <summary>
        /// 调用指定技能
        /// </summary>
        Task<string> CallSkillAsync(string skillName, IPluginContext ctx, string[] args);

        /// <summary>
        /// 获取插件管理器实例（用于兼容旧插件访问 robot.PluginManager）
        /// </summary>
        IRobot PluginManager => this;

        /// <summary>
        /// 会话管理器
        /// </summary>
        SessionManager Sessions { get; }
    }
}


