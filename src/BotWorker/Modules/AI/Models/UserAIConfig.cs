using BotWorker.Infrastructure.Persistence.ORM;

namespace BotWorker.Modules.AI.Models
{
    public class UserAIConfig : MetaDataGuid<UserAIConfig>
    {
        public override string TableName => "UserAIConfig";
        public override string KeyField => "Id";

        public long UserId { get; set; }
        public string ProviderName { get; set; } = string.Empty;
        public string ApiKey { get; set; } = string.Empty;
        public string BaseUrl { get; set; } = string.Empty;
        public bool IsLeased { get; set; } = false;
        public int UseCount { get; set; } = 0;
        public DateTime LastUsedAt { get; set; }

        public static async Task<UserAIConfig?> GetUserConfigAsync(long userId, string providerName)
        {
            var sql = $"SELECT * FROM {FullName} WHERE UserId = {userId} AND ProviderName = {providerName.Quotes()}";
            var list = await QueryListAsync<UserAIConfig>(sql);
            return list.FirstOrDefault();
        }

        public static async Task<List<UserAIConfig>> GetLeasedConfigsAsync(string providerName)
        {
            var sql = $"SELECT * FROM {FullName} WHERE IsLeased = 1 AND ProviderName = {providerName.Quotes()} AND ApiKey <> ''";
            return await QueryListAsync<UserAIConfig>(sql);
        }

        public static async Task<int> UpdateUsageAsync(long id)
        {
            var sql = $"UPDATE {FullName} SET UseCount = UseCount + 1, LastUsedAt = GETDATE() WHERE Id = {id}";
            return await ExecAsync(sql);
        }
    }
}
