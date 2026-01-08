using System.Runtime.CompilerServices;
using System.Text.RegularExpressions;
using Serilog;

namespace BotWorker.Core.Logging
{
    /// <summary>
    /// 基于 Serilog 的统一日志管理�?    /// </summary>
    public static class Logger
    {
        private static readonly object _lock = new();

        #region 基础日志方法 (转发�?Serilog)

        public static void Debug(string message, object? data = null, [CallerFilePath] string file = "", [CallerMemberName] string member = "", [CallerLineNumber] int line = 0)
            => WriteLog(Serilog.Events.LogEventLevel.Debug, message, null, data, file, member, line);

        public static void Info(string message, object? data = null, [CallerFilePath] string file = "", [CallerMemberName] string member = "", [CallerLineNumber] int line = 0)
            => WriteLog(Serilog.Events.LogEventLevel.Information, message, null, data, file, member, line);

        public static void Warn(string message, object? data = null, [CallerFilePath] string file = "", [CallerMemberName] string member = "", [CallerLineNumber] int line = 0)
            => WriteLog(Serilog.Events.LogEventLevel.Warning, message, null, data, file, member, line);

        public static void Error(string message, Exception? ex = null, object? data = null, [CallerFilePath] string file = "", [CallerMemberName] string member = "", [CallerLineNumber] int line = 0)
            => WriteLog(Serilog.Events.LogEventLevel.Error, message, ex, data, file, member, line);

        public static void Fatal(string message, Exception? ex = null, object? data = null, [CallerFilePath] string file = "", [CallerMemberName] string member = "", [CallerLineNumber] int line = 0)
            => WriteLog(Serilog.Events.LogEventLevel.Fatal, message, ex, data, file, member, line);

        private static void WriteLog(
            Serilog.Events.LogEventLevel level,
            string message,
            Exception? ex = null,
            object? data = null,
            string file = "",
            string member = "",
            int line = 0)
        {
            var fileName = Path.GetFileName(file);
            var contextMessage = $"{message} (at {fileName}:{member}, line {line})";

            if (data != null)
            {
                Log.Write(level, ex, "{@Message} | Data: {@Data}", contextMessage, data);
            }
            else
            {
                Log.Write(level, ex, contextMessage);
            }
        }

        #endregion

        #region 业务彩色显示逻辑 (兼容�?ShowMessage)

        public static void Show(string text, ConsoleColor color = ConsoleColor.Cyan, bool displayTime = true)
        {
            // 同时记录到文件日�?            Log.Information("[Show] {Text}", text);

            lock (_lock)
            {
                if (displayTime)
                {
                    Console.ForegroundColor = ConsoleColor.White;
                    Console.Write(DateTime.Now.ToLongTimeString() + " ");
                }

                var colorMap = new Dictionary<string, ConsoleColor>
                {
                    {"red", ConsoleColor.Red},
                    {"green", ConsoleColor.Green},
                    {"blue", ConsoleColor.Blue},
                    {"yellow", ConsoleColor.Yellow},
                    {"cyan", ConsoleColor.Cyan},
                    {"white", ConsoleColor.Gray}
                };

                var pattern = @"\<(.*?)\>(.*?)\<(.*?)\>";
                var matches = Regex.Matches(text, pattern);
                var textParts = Regex.Split(text, pattern);

                for (int i = 0; i < textParts.Length; i++)
                {
                    if (i < matches.Count)
                    {
                        var colorText = matches[i].Groups[1].Value.ToLower();
                        var textPart = matches[i].Groups[2].Value;

                        if (colorMap.TryGetValue(colorText, out var consoleColor))
                        {
                            Console.ForegroundColor = consoleColor;
                        }
                        Console.Write(textPart);
                    }
                    else
                    {
                        Console.ForegroundColor = color;
                        Console.Write(textParts[i]);
                    }
                }

                Console.ResetColor();
                Console.WriteLine();
            }
        }

        #endregion
    }
}


