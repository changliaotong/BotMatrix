using System.Threading.Tasks;
using BotWorker.Core.Plugin;

namespace BotWorker.Core.Pipeline
{
    /// <summary>
    /// 开关机状态中间件：处理机器人的 PowerOn/Off 逻辑
    /// </summary>
    public class PowerStatusMiddleware : IMiddleware
    {
        public async Task InvokeAsync(IPluginContext context, RequestDelegate next)
        {
            // 如果是在群组中，且机器人处于关机状态
            if (context.Group != null && !context.Group.IsPowerOn)
            {
                var msg = context.RawMessage.Trim();
                // 除非是开启指令，否则拦截
                if (msg != "开启" && msg != "#开启")
                {
                    if (context.IsMessage)
                    {
                        // 这里可以根据需要决定是否回复“已关机”
                        // 为了减少打扰，通常只在被 @ 或输入指令时回复
                        // await context.ReplyAsync("机器人已关机，请发送“开启”来启动。");
                    }
                    return; // 拦截
                }
            }

            await next(context);
        }
    }
}
