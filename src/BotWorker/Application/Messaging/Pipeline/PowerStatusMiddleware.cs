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
                    // 除非是开启指令，否则拦截
                    if (msg != "开机" && msg != "#开机")
                    {
                        // 这里可以根据需要决定是否回复“已关机”
                        // 为了减少打扰，通常只在被 @ 或输入指令时回复
                        return; // 拦截
                    }
                }
            }

            await next(context);
        }
    }
}
