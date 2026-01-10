namespace BotWorker.Common.Extensions
{
    public static class ConsoleExtensions
    {
        public static void WriteColor(this string message, ConsoleColor color)
        {
            var previous = Console.ForegroundColor;
            Console.ForegroundColor = color;
            Console.WriteLine(message);
            Console.ForegroundColor = previous;
        }

        public static void Info(this string message) => message.WriteColor(ConsoleColor.Cyan);
        public static void Success(this string message) => message.WriteColor(ConsoleColor.Green);
        public static void Warn(this string message) => message.WriteColor(ConsoleColor.Yellow);
        public static void Error(this string message) => message.WriteColor(ConsoleColor.Red);
    }
}


