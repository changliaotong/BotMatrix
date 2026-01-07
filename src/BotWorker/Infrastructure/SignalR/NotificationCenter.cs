using Microsoft.AspNetCore.SignalR;

namespace BotWorker.Infrastructure.SignalR
{
    public class NotificationCenter(IHubContext<MyHub> hub, IUserConnectionManager manager)
    {
        private readonly IHubContext<MyHub> _hub = hub;
        private readonly IUserConnectionManager _manager = manager;

        public async Task SendToUser(string userId, string method, object message)
        {
            var conns = _manager.GetConnections(userId);
            foreach (var conn in conns)
            {
                await _hub.Clients.Client(conn).SendAsync(method, message);
            }
        }

        public async Task SendToRole(string role, string method, object message)
        {
            var conns = _manager.GetConnectionsByRole(role);
            foreach (var conn in conns)
            {
                await _hub.Clients.Client(conn).SendAsync(method, message);
            }
        }

        public async Task Broadcast(string method, object message)
        {
            await _hub.Clients.All.SendAsync(method, message);
        }

        public async Task BroadcastExcept(string excludedConnId, string method, object message)
        {
            await _hub.Clients.AllExcept(excludedConnId).SendAsync(method, message);
        }


    }
}
