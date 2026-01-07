using Microsoft.AspNetCore.SignalR;
using StackExchange.Redis;
using sz84.Core.Services;

namespace sz84.Infrastructure.SignalR
{
    public class RedisUserConnectionManager : IUserConnectionManager
    {
        private readonly IDatabase _db;
        private readonly IHubContext<ChatHub>? _hub;

        public RedisUserConnectionManager(IConnectionMultiplexer redis, IHubContext<ChatHub>? hub = null)
        {
            _db = redis.GetDatabase();
            _hub = hub;
        }

        private static string ConnKey(string connId) => $"conn:{connId}";
        private static string UserKey(string userId) => $"user:{userId}";
        private static string RoleKey(string role) => $"role:{role}";

        public void AddConnection(string userId, string connectionId, string role = "user")
        {
            var tran = _db.CreateTransaction();

            tran.HashSetAsync(ConnKey(connectionId), new HashEntry[]
            {
            new("userId", userId),
            new("role", role),
            new("lastActive", DateTimeOffset.UtcNow.ToUnixTimeSeconds())
            });

            tran.SetAddAsync(UserKey(userId), connectionId);
            tran.SetAddAsync(RoleKey(role), connectionId);
            tran.SetAddAsync("online:all", connectionId);

            tran.Execute();

            // 触发上线事件
            _hub?.Clients.All.SendAsync("UserOnline", userId);
        }

        public void RemoveConnection(string connectionId)
        {
            var userId = _db.HashGet(ConnKey(connectionId), "userId");
            var role = _db.HashGet(ConnKey(connectionId), "role");

            var tran = _db.CreateTransaction();

            if (!userId.IsNullOrEmpty)
                tran.SetRemoveAsync(UserKey(userId.AsString()), connectionId);

            if (!role.IsNullOrEmpty)
                tran.SetRemoveAsync(RoleKey(role.AsString()), connectionId);

            tran.SetRemoveAsync("online:all", connectionId);
            tran.KeyDeleteAsync(ConnKey(connectionId));
            tran.Execute();

            // 检查是否还有其他连接
            if (!userId.IsNullOrEmpty && _db.SetLength(UserKey(userId.AsString())) == 0)
            {
                _hub?.Clients.All.SendAsync("UserOffline", userId.ToString());
            }
        }

        public void UpdateActivity(string connectionId)
        {
            _db.HashSet(ConnKey(connectionId), "lastActive", DateTimeOffset.UtcNow.ToUnixTimeSeconds());
        }

        public bool IsOnline(string userId)
        {
            return _db.SetLength(UserKey(userId)) > 0;
        }

        public List<string> GetConnections(string userId)
        {
            return [.. _db.SetMembers(UserKey(userId.AsString())).Select(x => x.AsString())];
        }

        public string? GetSingleConnection(string userId)
        {
            return GetConnections(userId).FirstOrDefault();
        }

        public List<string> GetConnectionsByRole(string role)
        {
            return [.. _db.SetMembers(RoleKey(role)).Select(x => x.AsString())];
        }

        public List<string> GetOnlineUserIds()
        {
            var allConns = _db.SetMembers("online:all");
            var userIds = new HashSet<string>();

            foreach (var connId in allConns)
            {
                var uid = _db.HashGet(ConnKey(connId.AsString()), "userId");
                if (!uid.IsNullOrEmpty) userIds.Add(uid!);
            }

            return userIds.ToList();
        }

        public int OnlineUserCount => (int)_db.SetLength("online:all");

        public int GetRoleCount(string role)
        {
            return (int)_db.SetLength(RoleKey(role));
        }

        public void RemoveInactiveConnections(TimeSpan timeout)
        {
            var now = DateTimeOffset.UtcNow.ToUnixTimeSeconds();
            var threshold = now - (long)timeout.TotalSeconds;

            var conns = _db.SetMembers("online:all");

            foreach (var connId in conns)
            {
                var lastActive = _db.HashGet(ConnKey(connId.AsString()), "lastActive");

                if (lastActive.TryParse(out long ts) && ts < threshold)
                {
                    RemoveConnection(connId!);
                }
            }
        }

        public bool CanSend(string connectionId)
        {
            return _db.KeyExists(ConnKey(connectionId));
        }
    }


}
