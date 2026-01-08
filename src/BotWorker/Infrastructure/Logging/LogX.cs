using System.Runtime.CompilerServices;
using System.Text.Json;

namespace BotWorker.Common
{
    public static class LogX
    {
        public enum LogLevel { Debug, Info, Warn, Error, Fatal }

        private static readonly object _lock = new();

        public static void Log2(
            LogLevel level,
            string message,
            Exception? ex = null,
            object? data = null,
            [CallerFilePath] string file = "",
            [CallerMemberName] string member = "",
            [CallerLineNumber] int line = 0)
        {
            lock (_lock)
            {
                var timestamp = DateTime.Now.ToString("yyyy-MM-dd HH:mm:ss.fff");
                var fileName = Path.GetFileName(file);
                var levelStr = level.ToString().ToUpper().PadRight(5);
                var color = GetConsoleColor(level);

                Console.ForegroundColor = color;

                Console.WriteLine($"[{timestamp}] [{levelStr}] {message}");
                Console.WriteLine($" => at {fileName}:{member} (line {line})");

                if (data != null)
                {
                    try
                    {
                        var json = JsonSerializer.Serialize(data, new JsonSerializerOptions
                        {
                            WriteIndented = true
                        });
                        Console.WriteLine($" => data: {json}");
                    }
                    catch
                    {
                        Console.WriteLine(" => data: [Failed to serialize]");
                    }
                }

                if (ex != null)
                {
                    Console.WriteLine($" => exception: {ex.GetType().Name}: {ex.Message}");
                    Console.WriteLine(ex.StackTrace);
                }

                Console.ResetColor();
            }
        }

        public static void Debug(string message, object? data = null,
            [CallerFilePath] string file = "", [CallerMemberName] string member = "", [CallerLineNumber] int line = 0)
            => Log2(LogLevel.Debug, message, null, data, file, member, line);

        public static void Info(string message, object? data = null,
            [CallerFilePath] string file = "", [CallerMemberName] string member = "", [CallerLineNumber] int line = 0)
            => Log2(LogLevel.Info, message, null, data, file, member, line);

        public static void Warn(string message, object? data = null,
            [CallerFilePath] string file = "", [CallerMemberName] string member = "", [CallerLineNumber] int line = 0)
            => Log2(LogLevel.Warn, message, null, data, file, member, line);

        public static void Error(string message, Exception? ex = null, object? data = null,
            [CallerFilePath] string file = "", [CallerMemberName] string member = "", [CallerLineNumber] int line = 0)
            => Log2(LogLevel.Error, message, ex, data, file, member, line);

        public static void Fatal(string message, Exception? ex = null, object? data = null,
            [CallerFilePath] string file = "", [CallerMemberName] string member = "", [CallerLineNumber] int line = 0)
            => Log2(LogLevel.Fatal, message, ex, data, file, member, line);

        private static ConsoleColor GetConsoleColor(LogLevel level) =>
            level switch
            {
                LogLevel.Debug => ConsoleColor.Gray,
                LogLevel.Info => ConsoleColor.Green,
                LogLevel.Warn => ConsoleColor.Yellow,
                LogLevel.Error => ConsoleColor.Red,
                LogLevel.Fatal => ConsoleColor.Magenta,
                _ => ConsoleColor.White
            };
    }

}


