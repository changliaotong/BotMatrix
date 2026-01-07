using BotWorker.Common.Exts;
using sz84.Core.MetaDatas;
using sz84.Infrastructure.Utils;

namespace sz84.Bots.Entries
{
    public class Token : MetaData<Token>
    {
        public override string TableName => "Token";
        public override string KeyField => "UserId";

        public static string GetQQ(string token)
        {
            return GetWhere("UserId", $"Token = {token.Quotes()}");
        }

        public static bool ExistsToken(string token)
        {
            return ExistsField("Token", token);
        }

        public static bool TokenValid(long qq, string token, long time = 60 * 60 * 24 * 30)
        {
            string res = GetValueAandB<string>($"ABS(DATEDIFF(SECOND, TokenDate, GETDATE()))", "UserId", qq, "Token", token);
            return !res.IsNull() && res.AsLong() < (BotInfo.IsSuperAdmin(qq) ? 60 * 60 * 24 * 365 : time);
        }

        public static bool ExistsToken(long userId, string token)
        {
            return ExistsAandB("UserId", userId, "Token", token);
        }

        public static (int, string) Append(long userId)
        {
            string token = $"{userId}{RandomInt(999999)}{DateTime.Now}".MD5()[1..7];
            var (sql, parameters) = Exists(userId)
                ? SqlSetValues($"Token={token.Quotes()}, TokenDate=GETDATE()", userId)
                : SqlInsert([
                    new Cov("UserId", userId),
                    new Cov("Token", token),
                ]);
            return (Exec(sql, parameters), token);
        }

        public static string GetToken(long qq)
        {
            string token = GetValue("Token", qq);
            int i = 0;
            if (token.IsNull())
                (i, token) = Append(qq);
            return i == -1 ? "" : token;
        }

        public static int SaveRefreshToken(long qq, string refreshToken)
        {
            string token = $"{qq}{RandomInt(999999)}{DateTime.Now}".MD5()[1..7];
            var (sql, parameters) = Exists(qq)
                ? SqlSetValues($"RefreshToken = {refreshToken.Quotes()}, ExpiryDate=GETDATE()+7", qq)
                : SqlInsert([
                                new Cov("UserId", qq),
                                new Cov("Token", token),
                                new Cov("RefreshToken", refreshToken),
                            ]);
            return Exec(sql, parameters);
        }

        public static string GetStoredRefreshToken(long qq)
        {
            return Get<string>("RefreshToken", qq);
        }
    }
}
