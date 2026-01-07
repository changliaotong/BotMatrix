using System.Diagnostics;

namespace BotWorker.Common.Exts
{
    public static class BenchmarkExtensions
    {
        public static TimeSpan Benchmark(this Action action, out string msg)
        {
            var sw = Stopwatch.StartNew();
            action();
            sw.Stop();
            msg = $"耗时: {sw.Elapsed.TotalMilliseconds} ms";
            return sw.Elapsed;
        }

        public static async Task<TimeSpan> BenchmarkAsync(this Func<Task> func, Action<TimeSpan>? onCompleted = null)
        {
            var sw = Stopwatch.StartNew();
            await func();
            sw.Stop();
            onCompleted?.Invoke(sw.Elapsed);
            return sw.Elapsed;
        }
    }

}
