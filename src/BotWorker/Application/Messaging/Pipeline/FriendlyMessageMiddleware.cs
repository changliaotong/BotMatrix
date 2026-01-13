using BotWorker.Infrastructure.Communication.OneBot;
using BotWorker.Domain.Interfaces;
using BotWorker.Modules.Plugins;

namespace BotWorker.Application.Messaging.Pipeline
{
    /// <summary>
    /// 友好化消息中间件：在消息发送前进行人性化处理（占位符替换、URL拦截等）
    /// </summary>
    public class FriendlyMessageMiddleware : IMiddleware
    {
        public async Task InvokeAsync(IPluginContext context, RequestDelegate next)
        {
            // 这是一个后置处理中间件，先执行后续逻辑
            await next(context);

            if (context is PluginContext pluginCtx && pluginCtx.Event is BotMessageEvent botMsgEvent)
            {
                var botMsg = botMsgEvent.BotMessage;

                // 如果有回答，则进行友好化处理
                if (!string.IsNullOrEmpty(botMsg.Answer))
                {
                    await botMsg.GetFriendlyResAsync();
                }
            }
        }
    }
}
