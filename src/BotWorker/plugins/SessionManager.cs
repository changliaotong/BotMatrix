using System;
using System.Text.Json;
using System.Threading.Tasks;
using StackExchange.Redis;

namespace BotWorker.Modules.Plugins
{
    public class UserSession
    {
        public string PluginId { get; set; } = string.Empty;
        public string Action { get; set; } = string.Empty;
        public string? Step { get; set; }
        public string? WaitingFor { get; set; }
        public string? DataJson { get; set; }
        public string? ConfirmationCode { get; set; }

        public T? GetData<T>()
        {
            if (string.IsNullOrEmpty(DataJson)) return default;
            try
            {
                return JsonSerializer.Deserialize<T>(DataJson);
            }
            catch
            {
                return default;
            }
        }
    }

    public class SessionManager
    {
        private readonly IDatabase _db;
        private const string Prefix = "bot:session:";

        public SessionManager(IConnectionMultiplexer redis)
        {
            _db = redis.GetDatabase();
        }

        private string GetSessionKey(string userId, string? groupId) 
            => $"{Prefix}{(string.IsNullOrEmpty(groupId) ? $"user:{userId}" : $"group:{groupId}:user:{userId}")}";

        public async Task SetSessionAsync(string userId, string? groupId, string pluginId, string action, object? data = null, int durationSeconds = 60, string? confirmationCode = null, string? step = null, string? waitingFor = null)
        {
            var key = GetSessionKey(userId, groupId);
            var session = new UserSession
            {
                PluginId = pluginId,
                Action = action,
                DataJson = data != null ? JsonSerializer.Serialize(data) : null,
                ConfirmationCode = confirmationCode,
                Step = step,
                WaitingFor = waitingFor
            };

            var json = JsonSerializer.Serialize(session);
            await _db.StringSetAsync(key, json, TimeSpan.FromSeconds(durationSeconds));
        }

        public async Task StartDialogAsync(string userId, string? groupId, string pluginId, string action, string? step = null, object? data = null, int durationSeconds = 300)
        {
            await SetSessionAsync(userId, groupId, pluginId, action, data, durationSeconds, null, step, "message");
        }

        public async Task<string> StartConfirmationAsync(string userId, string? groupId, string pluginId, string action, object? data = null, int durationSeconds = 60)
        {
            var code = GenerateConfirmationCode();
            await SetSessionAsync(userId, groupId, pluginId, action, data, durationSeconds, code);
            return code;
        }

        public async Task<UserSession?> GetSessionAsync(string userId, string? groupId)
        {
            var key = GetSessionKey(userId, groupId);
            var json = await _db.StringGetAsync(key);
            
            if (json.IsNullOrEmpty) return null;

            try
            {
                return JsonSerializer.Deserialize<UserSession>(json.ToString());
            }
            catch
            {
                return null;
            }
        }

        public async Task ClearSessionAsync(string userId, string? groupId)
        {
            var key = GetSessionKey(userId, groupId);
            await _db.KeyDeleteAsync(key);
        }

        public string GenerateConfirmationCode()
        {
            return new Random().Next(100, 999).ToString();
        }
    }
}
