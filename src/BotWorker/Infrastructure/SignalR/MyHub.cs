using Microsoft.AspNetCore.SignalR;
using Microsoft.AspNetCore.SignalR.Client;
using Microsoft.Extensions.Logging;
using BotWorker.Modules.AI.Plugins;
using BotWorker.Infrastructure.Caching;

namespace BotWorker.Infrastructure.SignalR
{
    public class MyHub(ICacheService cache, KnowledgeBaseService qaService, IUserConnectionManager userConnectionManager, Microsoft.Extensions.Logging.ILogger<MyHub> logger) : Hub
    {
        private readonly IUserConnectionManager _userConnectionManager = userConnectionManager;
        public readonly ICacheService _cache = cache;
        public readonly KnowledgeBaseService _qaService = qaService;
        public readonly Microsoft.Extensions.Logging.ILogger<MyHub> _logger = logger;


        public static async Task NotifyUpgrade(IHubContext<MyHub> context)
        {
            await context.Clients.All.SendAsync("ServerUpgrade");
        }

        public static class ConnectionMap
        {
            private static readonly Dictionary<string, string> _userToConnection = [];

            public static Dictionary<string, string> UserToConnection
            {
                get { return _userToConnection; }
            }
        }

        public override async Task OnConnectedAsync()
        {
            var userId = Context.UserIdentifier;   
            if (string.IsNullOrEmpty(userId))
                userId = Context.ConnectionId;

            var role = Context.User?.IsInRole("admin") == true ? "admin" : "user";
            InfoMessage($"用户连接: {Context.ConnectionId} userId：{userId}");

            await AddToGroupAsync(userId);
            await AddToGroupAsync(role);

            InfoMessage($"已加入 {userId} {role} 组");

            _userConnectionManager.AddConnection(userId, Context.ConnectionId, role);

            await base.OnConnectedAsync();
        }

        public override async Task OnDisconnectedAsync(Exception? exception)
        {
            var userId = Context.User?.Identity?.Name ?? Context.ConnectionId;
            var role = Context.User?.IsInRole("admin") == true ? "admin" : "user";
            InfoMessage($"用户断开: {Context.ConnectionId} 用户名：{userId}");
            if (!string.IsNullOrEmpty(userId))
            {
                _userConnectionManager.RemoveConnection(Context.ConnectionId);
                await RemoveFromGroupAsync(userId); 
                await RemoveFromGroupAsync(role);
                InfoMessage($"已从{userId}组中移除");
            }    

            await base.OnDisconnectedAsync(exception);
        }

        public Task<bool> Ping()
        {
            _userConnectionManager.UpdateActivity(Context.ConnectionId);
            //InfoMessage($"用户活跃: {Context.ConnectionId}");
            return Task.FromResult(true); 
        }

        public async Task SendToUser(string targetUsername, string message)
        {
            var connections = await RedisConnectionMap.GetConnectionsAsync(targetUsername);

            foreach (var connId in connections)
            {
                await Clients.Client(connId).SendAsync("ReceiveMessage", message);
            }
        }

        public static string GetConnectionId(string userId)
        {
            return ConnectionMap.UserToConnection.GetValueOrDefault(userId) ?? "";
        }

        public async Task AddToGroupAsync(string roomId)
        {
            await Groups.AddToGroupAsync(Context.ConnectionId, roomId);
        }

        public async Task RemoveFromGroupAsync(string roomId)
        {
            await Groups.RemoveFromGroupAsync(Context.ConnectionId, roomId);
        }
    }
}
