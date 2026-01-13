using Microsoft.AspNetCore.SignalR;

namespace BotWorker.Infrastructure.SignalR
{
    public class HeaderUserIdProvider : IUserIdProvider
    {
        public string? GetUserId(HubConnectionContext connection)
        {
            var httpContext = connection.GetHttpContext();

            // 从 Header 中取你传过来的 userId（就是 QQ）
            if (httpContext != null && httpContext.Request.Headers.TryGetValue("userId", out var userId))
            {
                return userId.FirstOrDefault();
            }

            return null;
        }
    }
}
