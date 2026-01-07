using Microsoft.Extensions.Logging;
using System.Diagnostics;

namespace BotWorker.Common.Exts
{
    public static class DiagnosticsExtensions
    {
        // 计时器包装，执行代码并输出耗时日志
        public static void TimeAction(this ILogger logger, string actionName, Action action)
        {
            ArgumentNullException.ThrowIfNull(logger);
            ArgumentNullException.ThrowIfNull(action);

            var sw = Stopwatch.StartNew();
            try
            {
                action();
            }
            finally
            {
                sw.Stop();
                logger.LogInformation($"{actionName} 耗时: {sw.ElapsedMilliseconds} ms");
            }
        }

        // 计时器包装异步
        public static async Task TimeActionAsync(this ILogger logger, string actionName, Func<Task> func)
        {
            ArgumentNullException.ThrowIfNull(logger);
            ArgumentNullException.ThrowIfNull(func);

            var sw = Stopwatch.StartNew();
            try
            {
                await func();
            }
            finally
            {
                sw.Stop();
                logger.LogInformation($"{actionName} 耗时: {sw.ElapsedMilliseconds} ms");
            }
        }
    }
}
