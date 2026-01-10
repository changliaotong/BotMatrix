using BotWorker.Infrastructure.Communication.OneBot;

namespace BotWorker.Application.Messaging.Pipeline
{
    /// <summary>
    /// 问答中间件：在正则指令识别之后，尝试匹配问答库
    /// </summary>
    public class QaMiddleware : IMiddleware
    {
        public async Task InvokeAsync(IPluginContext context, RequestDelegate next)
        {
            if (context is PluginContext pluginCtx && pluginCtx.Event is BotMessageEvent botMsgEvent)
            {
                var botMsg = botMsgEvent.BotMessage;

                // 如果已经有回答了，或者已经是指令了，则跳过
                if (!string.IsNullOrEmpty(botMsg.Answer) || botMsg.IsCmd)
                {
                    await next(context);
                    return;
                }

                // 尝试匹配问答
                context.Logger?.LogInformation("[QaMiddleware] Attempting QA for message {MessageId}", context.EventId);
                await botMsg.GetAnswerAsync();

                if (!string.IsNullOrEmpty(botMsg.Answer))
                {
                    context.Logger?.LogInformation("[QaMiddleware] Handled by GetAnswerAsync, Answer: {Answer}", botMsg.Answer);
                    return; // 拦截，直接返回
                }
            }

            await next(context);
        }
    }
}
