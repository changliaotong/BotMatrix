using Microsoft.AspNetCore.SignalR;
using BotWorker.Core.Services;

namespace BotWorker.Infrastructure.SignalR
{
    public static class SignalRHelper
    {
        public static async Task SafeReplyAsync(IHubContext<ChatHub> hub, string connectionId, string method, object? arg1 = null, object? arg2 = null)
        {
            try
            {
                if (!string.IsNullOrWhiteSpace(connectionId))
                {
                    await hub.Clients.Client(connectionId).SendAsync(method, arg1, arg2);
                }
            }
            catch (Exception ex)
            {
                Console.WriteLine($"[SafeReply] 消息发送失败：{ex.Message}");
            }
        }
    }

}
