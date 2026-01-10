using Microsoft.AspNetCore.SignalR;
using BotWorker.Application.Services;

namespace BotWorker.Infrastructure.SignalR
{
    /// <summary>
    /// 统一处理 Hub 调用的日志拦截器
    /// </summary>
    public class HubLoggingFilter : IHubFilter
    {
        public async ValueTask<object?> InvokeMethodAsync(
            HubInvocationContext invocationContext, Func<HubInvocationContext, ValueTask<object?>> next)
        {
            // 统一处理 ChatHub 的日志输出
            if (invocationContext.Hub is ChatHub chatHub)
            {
                var methodName = invocationContext.HubMethodName;
                var staticMsg = chatHub.StaticMessage;
                
                // 排除一些过于频繁的调用，如 Ping
                if (methodName != "Ping")
                {
                    Logger.Show($"[Hub] {methodName} {staticMsg}");
                }
            }

            try
            {
                return await next(invocationContext);
            }
            catch (Exception ex)
            {
                Logger.Error($"[Hub] 调用 {invocationContext.HubMethodName} 出错: {ex.Message}", ex);
                throw;
            }
        }
    }
}
