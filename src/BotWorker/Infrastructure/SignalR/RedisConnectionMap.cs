using StackExchange.Redis;
using BotWorker.Common;

namespace sz84.Infrastructure.SignalR
{
    public static class RedisConnectionMap
    {
        private static readonly Lazy<ConnectionMultiplexer> _redis = new(() =>
            ConnectionMultiplexer.Connect(GlobalConfig.RedisConnection)); 

        private static IDatabase Db => _redis.Value.GetDatabase();

        private static string GetKey(string username) => $"signalr:user:{username}";

        // 添加连接ID
        public static async Task AddConnectionAsync(string username, string connectionId)
        {
            await Db.SetAddAsync(GetKey(username), connectionId);
        }

        // 移除连接ID
        public static async Task RemoveConnectionAsync(string username, string connectionId)
        {
            await Db.SetRemoveAsync(GetKey(username), connectionId);

            // 如果连接为空，可选：删除整个键
            if (await Db.SetLengthAsync(GetKey(username)) == 0)
            {
                await Db.KeyDeleteAsync(GetKey(username));
            }
        }

        // 获取某用户的所有连接ID
        public static async Task<IEnumerable<string>> GetConnectionsAsync(string username)
        {
            var members = await Db.SetMembersAsync(GetKey(username));
            return members.Select(m => m.ToString() ?? "");
        }
    }

}
