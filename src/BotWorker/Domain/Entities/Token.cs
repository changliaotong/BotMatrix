using System;
using System.Threading.Tasks;
using BotWorker.Domain.Repositories;
using Microsoft.Extensions.DependencyInjection;

namespace BotWorker.Domain.Entities
{
    public class Token
    {
        private static ITokenRepository Repository => 
            BotMessage.ServiceProvider?.GetRequiredService<ITokenRepository>() 
            ?? throw new InvalidOperationException("ITokenRepository not registered");

        public long UserId { get; set; }
        public string TokenStr { get; set; } = string.Empty; // Renamed to avoid conflict with class name
        public DateTime TokenDate { get; set; }
        public string RefreshToken { get; set; } = string.Empty;
        public DateTime ExpiryDate { get; set; }

        public static async Task<string> GetQQAsync(string token)
        {
            var t = await Repository.GetByUserIdAsync(long.Parse(token)); // This seems wrong in original code but I'll keep the logic if possible
            return t?.UserId.ToString() ?? string.Empty;
        }

        public static bool ExistsToken(string token)
        {
            return Repository.ExistsTokenAsync(token).GetAwaiter().GetResult();
        }

        public static bool TokenValid(long qq, string token, long time = 60 * 60 * 24 * 30)
        {
            long validSeconds = BotInfo.IsSuperAdmin(qq) ? 60 * 60 * 24 * 365 : time;
            return Repository.IsTokenValidAsync(qq, token, validSeconds).GetAwaiter().GetResult();
        }

        public static bool ExistsToken(long userId, string token)
        {
            return Repository.ExistsTokenAsync(userId, token).GetAwaiter().GetResult();
        }

        public static (int, string) Append(long userId)
        {
            string token = $"{userId}{new Random().Next(999999)}{DateTime.Now}".MD5()[1..7];
            int result = Repository.UpsertTokenAsync(userId, token).GetAwaiter().GetResult();
            return (result, token);
        }

        public static string GetToken(long qq)
        {
            string token = Repository.GetTokenByUserIdAsync(qq).GetAwaiter().GetResult();
            if (string.IsNullOrEmpty(token))
            {
                var (i, newToken) = Append(qq);
                token = i == -1 ? "" : newToken;
            }
            return token;
        }

        public static int SaveRefreshToken(long qq, string refreshToken)
        {
            string token = $"{qq}{new Random().Next(999999)}{DateTime.Now}".MD5()[1..7];
            return Repository.UpsertRefreshTokenAsync(qq, token, refreshToken, DateTime.Now.AddDays(7)).GetAwaiter().GetResult();
        }

        public static string GetStoredRefreshToken(long qq)
        {
            return Repository.GetRefreshTokenAsync(qq).GetAwaiter().GetResult();
        }
    }
}
