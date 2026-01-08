using System;
using System.Threading.Tasks;
using BotWorker.Core.Plugin;
using Microsoft.Extensions.Logging;

namespace BotWorker.Core.Pipeline
{
    /// <summary>
    /// 全局异常处理中间件：捕获管道内所有异常并记录日志
    /// </summary>
    public class ExceptionMiddleware : IMiddleware
    {
        private readonly ILogger<ExceptionMiddleware> _logger;

        public ExceptionMiddleware(ILogger<ExceptionMiddleware> logger)
        {
            _logger = logger;
        }

        public async Task InvokeAsync(IPluginContext context, RequestDelegate next)
        {
            try
            {
                await next(context);
            }
            catch (Exception ex)
            {
                _logger.LogError(ex, "处理插件请求时发生未捕获的异常。Context: {ContextId}", context.GetHashCode());

                if (context is PluginContext pluginCtx && pluginCtx.Event is Core.OneBot.BotMessageEvent botMsgEvent)
                {
                    var botMsg = botMsgEvent.BotMessage;
                    botMsg.Answer = "⚠️ 抱歉，处理您的请求时发生了内部错误，请稍后再试。";
                    botMsg.Reason += $"[异常: {ex.Message}]";
                }
            }
        }
    }
}
