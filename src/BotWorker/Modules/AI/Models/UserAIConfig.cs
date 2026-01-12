using BotWorker.Infrastructure.Persistence.ORM;
using BotWorker.Infrastructure.Utils;

namespace BotWorker.Modules.AI.Models
{
    public class UserAIConfig : MetaDataGuid<UserAIConfig>
    {
        private static readonly string _encryptKey = "AI_KEY_SECRET_2024_BOT_MATRIX"; // 建议从配置读取

        public override string TableName => "UserAIConfig";
        public override string KeyField => "Id";

        public long UserId { get; set; }
        public string ProviderName { get; set; } = string.Empty;
        public string ApiKey { get; set; } = string.Empty; // 数据库存储加密后的值
        public string BaseUrl { get; set; } = string.Empty;
        public bool IsLeased { get; set; } = false;
        public int UseCount { get; set; } = 0;
        public DateTime LastUsedAt { get; set; }

        /// <summary>
        /// 获取解密后的 API Key
        /// </summary>
        public string GetDecryptedApiKey()
        {
            if (string.IsNullOrEmpty(ApiKey)) return string.Empty;
            try
            {
                // 如果是明文（非 Base64 或 解密失败），则返回原值（兼容旧数据）
                var decrypted = ApiKey.Decrypt3DES(_encryptKey.MD5().Substring(0, 24));
                return string.IsNullOrEmpty(decrypted) ? ApiKey : decrypted;
            }
            catch
            {
                return ApiKey;
            }
        }

        /// <summary>
        /// 设置并加密 API Key
        /// </summary>
        public void SetEncryptedApiKey(string plainKey)
        {
            if (string.IsNullOrEmpty(plainKey))
            {
                ApiKey = string.Empty;
                return;
            }
            ApiKey = plainKey.Encrypt3DES(_encryptKey.MD5().Substring(0, 24));
        }

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
            var sql = $"UPDATE {FullName} SET UseCount = UseCount + 1, LastUsedAt = {SqlDateTime} WHERE Id = {id}";
            return await ExecAsync(sql);
        }
    }
}
