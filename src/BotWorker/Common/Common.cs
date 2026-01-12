namespace BotWorker.Core.Configurations
{
    public class AppConfig
    {
        private static IConfiguration? _configuration;

        public static void Initialize(IConfiguration configuration)
        {
            _configuration = configuration;
        }

        internal static string _url => _configuration?["sz84:url"] ?? "https://sz84.com";
        internal static string _apiKey => _configuration?["sz84:api_key"] ?? "";

        public static string AzureTranslateSubscriptionKey => _configuration?["azure:translate:subscription_key"] ?? "";
        public static string AzureTranslateEndpoint => _configuration?["azure:translate:endpoint"] ?? "https://api.cognitive.microsofttranslator.com";
        public static string AzureTranslateLocation => _configuration?["azure:translate:location"] ?? "global";

        public static string RetryMsg => "操作失败，请稍后重试";
        public static string RetryMsgTooFast => "速度太快了，请稍后再试";
        public static string YearOnlyMsg => "非年费版不能使用此功能";
        public static string SetupUrl => _configuration?["sz84:SetupUrl"] ?? _url;
        public static string NoAnswer => "这个问题我不会，输入【教学】了解如何教我说话";
        public static string AnswerExists => "这个我已经学过了，再教我点别的吧~";
        public static string BlackListMsg => "该号码已被官方拉黑";
        public static string CreditSystemClosed => "积分系统已关闭";

        public static string OwnerOnlyMsg => "此命令仅机器人主人可用";
        public const long SystemPromptGroup = 320;
        public const long C2CMessageGroupId = 990000000003;
        public static int RandomInt(int max) => new Random().Next(max + 1);
        public static int RandomInt(int min, int max) => new Random().Next(min, max + 1);
        public static long RandomInt64(long max) => new Random().NextInt64(max + 1);
        public static long RandomInt64(long min, long max) => new Random().NextInt64(min, max + 1);
        public static long[] OfficalBots { get; set; } = [3889418604, 3889420782, 3889411042, 3889610970, 3889535978, 3889494926, 3889699720, 3889699721, 3889699722, 3889699723];
    }
}

