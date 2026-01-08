using Microsoft.Extensions.Configuration;

namespace BotWorker.Common.Config
{
    public class AppConfig
    {
        private static IConfiguration? _configuration;

        public static void Initialize(IConfiguration configuration)
        {
            _configuration = configuration;
        }

        internal static string _url => _configuration?["sz84:url"] ?? "https://sz84.com";
        internal static string _apiKey => _configuration?["sz84:api_key"] ?? "AFCDE195E9EE00DCFCB5E0ED44D129EB";

        public static string RetryMsgTooFast => _configuration?["Messages:RetryMsgTooFast"] ?? "速度太快了，请稍后再试";
        public static string OwnerOnlyMsg => _configuration?["Messages:OwnerOnlyMsg"] ?? "此命令仅机器人主人可用";
        public static string YearOnlyMsg => _configuration?["Messages:YearOnlyMsg"] ?? "非年费版不能使用此功能";
        public static long[] OfficalBots { get; set; } = [3889418604, 3889420782, 3889411042, 3889610970, 3889535978, 3889494926, 3889699720, 3889699721, 3889699722, 3889699723];
    }
}
