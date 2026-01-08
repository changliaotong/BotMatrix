using System.Security.Cryptography;
using System.Text;

namespace BotWorker.Common.Extensions
{
    public static class SecurityExtensions
    {
        public static string ToSha256(this string str)
        {
            var bytes = Encoding.UTF8.GetBytes(str);
            var hash = SHA256.HashData(bytes);
            return BitConverter.ToString(hash).Replace("-", "").ToLower();
        }

        public static string ToBase64(this string str) =>
            Convert.ToBase64String(Encoding.UTF8.GetBytes(str));

        public static string FromBase64(this string base64) =>
            Encoding.UTF8.GetString(Convert.FromBase64String(base64));
    }

}


