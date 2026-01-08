using System;
using System.Threading.Tasks;
using BotWorker.Core.Plugin;
using BotWorker.Core.Database;

namespace BotWorker.Core.Pipeline
{
    /// <summary>
    /// 系统维护中间件：处理凌晨 4:00 - 4:01 的系统维护逻辑
    /// </summary>
    public class MaintenanceMiddleware : IMiddleware
    {
        public async Task InvokeAsync(IPluginContext context, RequestDelegate next)
        {
            // 凌晨 4 点零 1 �?数据维护
            DateTime dt = SQLConn.GetDate();
            if (context.Platform != "guild" && dt.Hour == 4 && dt.Minute < 1)
            {
                if (context.IsMessage)
                {
                    await context.ReplyAsync($"每天 04:00-04:01 系统维护中，请稍后再�?..\n当前时间：{dt:MM-dd HH:mm:ss}");
                    return; // 拦截后续逻辑
                }
            }

            await next(context);
        }
    }
}


