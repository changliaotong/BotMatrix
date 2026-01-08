namespace BotWorker.BotWorker.BotWorker.Common.Exts
{
    public static class LoggingExtensions
    {
        public static void Log(this string message) =>
            Console.WriteLine($"[{DateTime.Now:HH:mm:ss}] {message}");

        public static void LogError(this Exception ex) =>
            Console.WriteLine($"[ERROR] {ex.Message}\n{ex.StackTrace}");
    }
}


