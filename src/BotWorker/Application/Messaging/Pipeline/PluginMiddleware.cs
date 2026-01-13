using System.Threading.Tasks;
using BotWorker.Domain.Interfaces;
using BotWorker.Modules.Plugins;

namespace BotWorker.Application.Messaging.Pipeline
{
    /// <summary>
    /// 插件中间件：负责将消息分发给插件系统。通常作为管道的最后一个中间件。
    /// </summary>
    public class PluginMiddleware : IMiddleware
    {
        private readonly PluginManager _pluginManager;

        public PluginMiddleware(PluginManager pluginManager)
        {
            _pluginManager = pluginManager;
        }

        public async Task InvokeAsync(IPluginContext context, RequestDelegate next)
        {
            // 插件中间件通常是管道的末端，它负责实际的业务分发
            var result = await _pluginManager.DispatchAsync(context);
            
            // 如果插件返回了结果且当前还没有设置 Answer，则尝试回复
            if (!string.IsNullOrEmpty(result))
            {
                await context.ReplyAsync(result);
            }

            // 继续执行下一个中间件（如果有）
            await next(context);
        }
    }
}
