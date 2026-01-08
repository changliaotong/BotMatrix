namespace BotWorker.BotWorker.BotWorker.Common.Exts
{
    public static class CommandExtensions
    {
        public static Dictionary<string, string> ParseOptions(this string[] args)
        {
            return args
                .Where(x => x.Contains('='))
                .Select(x => x.Split('='))
                .ToDictionary(x => x[0], x => x[1]);
        }

        public static string Arg(this string[] args, int index, string defaultValue = "")
            => args.Length > index ? args[index] : defaultValue;
    }
}


