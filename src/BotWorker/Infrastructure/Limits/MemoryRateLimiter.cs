namespace BotWorker.Infrastructure.Limits
{
    // MemoryRateLimiter.cs
    public class MemoryRateLimiter : IRateLimiter
    {
        private readonly Dictionary<string, List<DateTime>> _accessRecords = new();

        private readonly object _lock = new();

        public bool CheckLimit(string key, int maxCount, TimeSpan period)
        {
            lock (_lock)
            {
                if (!_accessRecords.TryGetValue(key, out var times))
                {
                    times = [];
                    _accessRecords[key] = times;
                }

                var now = DateTime.UtcNow;
                times.RemoveAll(t => now - t > period);

                if (times.Count >= maxCount)
                    return false;

                times.Add(now);
                return true;
            }
        }
    }


}
