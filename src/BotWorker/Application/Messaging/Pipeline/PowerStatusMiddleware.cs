using System.Threading.Tasks;
using BotWorker.Domain.Interfaces;
using BotWorker.Modules.Plugins;

namespace BotWorker.Application.Messaging.Pipeline
{
    /// <summary>
    /// 开关机状态中间件：处理机器人 PowerOn/Off 逻辑
    /// </summary>
    public class PowerStatusMiddleware : IMiddleware
    {
        public async Task InvokeAsync(IPluginContext context, RequestDelegate next)
        {
            if (context is PluginContext pluginCtx && pluginCtx.Event is Infrastructure.Communication.OneBot.BotMessageEvent botMsgEvent)
            {
                var botMsg = botMsgEvent.BotMessage;

                // 如果是在群组中，且机器人处于关机状态
                if (botMsg.IsGroup && !botMsg.Group.IsPowerOn)
                {
                    var msg = botMsg.CurrentMessage.Trim();
                    // 除非是开启/开机指令，否则拦截
                    if (msg != "开机" && msg != "#开机" && msg != "开启" && msg != "#开启")
                    {
                        context.Logger?.LogInformation("[PowerStatus] Intercepted: Bot is PowerOff in group {GroupId}, message {MessageId}", botMsg.GroupId, context.EventId);
                        return; // 拦截
                    }
                }
            }

            await next(context);
        }
    }
}
