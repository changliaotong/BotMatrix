namespace BotWorker.Common.Extensions
{
    public static class CacheExtensions
    {
        private static readonly Dictionary<string, object> _memory = [];

        public static void SetCache<T>(this string key, T value) => _memory[key] = value!;
        public static T? GetCache<T>(this string key) => _memory.TryGetValue(key, out var val) ? (T)val : default;

        public static bool HasCache(this string key) => _memory.ContainsKey(key);
    }
}


