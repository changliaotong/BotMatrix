using System.Text.RegularExpressions;
using Microsoft.Extensions.Configuration;

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
        internal static string _apiKey => _configuration?["sz84:api_key"] ?? "AFCDE195E9EE00DCFCB5E0ED44D129EB";

        public static string RetryMsgTooFast => _configuration?["Messages:RetryMsgTooFast"] ?? "速度太快了，请稍后再试";
        public static string YearOnlyMsg => _configuration?["Messages:YearOnlyMsg"] ?? "非年费版不能使用此功能";
        public static string SetupUrl => _configuration?["sz84:SetupUrl"] ?? _url;
        public static long[] OfficalBots { get; set; } = [3889418604, 3889420782, 3889411042, 3889610970, 3889535978, 3889494926, 3889699720, 3889699721, 3889699722, 3889699723];
    }
}

namespace BotWorker.Common
{
    // 改为一个极其简单的名字，避免与命名空间冲突
    public static class C
    {
        public static string url => BotWorker.Core.Configurations.AppConfig._url;
        public static string apiKey => BotWorker.Core.Configurations.AppConfig._apiKey;
        public const string NoAnswer = "这个问题我不会，输入【教学】了解如何教我说话";
        public const string AnswerExists = "这个我已经学过了，再教我点别的吧~";
        public const string BlackListMsg = "该号码已被官方拉黑";
        public const string CreditSystemClosed = "积分系统已关闭";
        public const string RetryMsg = "操作失败，请稍后重试";
        public static string OwnerOnlyMsg => "此命令仅机器人主人可用";
        public const long SystemPromptGroup = 320;
        public const long C2CMessageGroupId = 990000000003;

        public static int RandomInt(int max) => new Random().Next(max + 1);
        public static int RandomInt(int min, int max) => new Random().Next(min, max + 1);
        public static long RandomInt64(long max) => new Random().NextInt64(max + 1);
        public static long RandomInt64(long min, long max) => new Random().NextInt64(min, max + 1);
        
        public static void ShowMessage(string message) => Console.WriteLine(message);
        public static void ShowMessage(string message, ConsoleColor color, bool displayTime = true) => Console.WriteLine(message);
        public static void InfoMessage(string message) => Console.WriteLine($"[INFO] {message}");
        public static void InfoMessage(string message, ConsoleColor color) => Console.WriteLine($"[INFO] {message}");
        public static void ErrorMessage(string message) => Console.Error.WriteLine($"[ERROR] {message}");
        public static void ErrorMessage(string message, bool displayTime = true, ConsoleColor color = ConsoleColor.Red) => Console.Error.WriteLine($"[ERROR] {message}");
    }
}
