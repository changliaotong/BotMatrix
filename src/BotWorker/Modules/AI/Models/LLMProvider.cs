using System;
using System.ComponentModel.DataAnnotations.Schema;
using BotWorker.Infrastructure.Utils;

namespace BotWorker.Modules.AI.Models
{
    [Table("ai_providers")]
    public class LLMProvider
    {
        private static readonly string _encryptKey = "AI_KEY_SECRET_2024_BOT_MATRIX";

        public long Id { get; set; }
        public string Name { get; set; } = string.Empty;
        public string Type { get; set; } = string.Empty; // openai, azure, ollama, etc.
        public string? Endpoint { get; set; }
        public string? ApiKey { get; set; }
        public string Config { get; set; } = "{}"; // JSONB
        public bool IsActive { get; set; } = true;
        public long OwnerId { get; set; }
        public bool IsShared { get; set; }
        public DateTime CreatedAt { get; set; } = DateTime.UtcNow;
        public DateTime UpdatedAt { get; set; } = DateTime.UtcNow;

        /// <summary>
        /// 获取解密后的 API Key
        /// </summary>
        public string GetDecryptedApiKey()
        {
            if (string.IsNullOrEmpty(ApiKey)) return string.Empty;
            try
            {
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
    }
}
