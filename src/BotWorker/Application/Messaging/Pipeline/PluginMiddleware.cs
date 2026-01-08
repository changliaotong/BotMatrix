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
            
            // 如果插件返回了结果且 Context 还没 Answer，则设置 Answer
            // 注意：由于 IPluginContext 的设计，插件可能已经通过 ReplyAsync 自行回复了。
            // 这里的逻辑可以根据需要调整。

            // 继续执行下一个中间件（如果有）
            await next(context);
        }
    }
}
