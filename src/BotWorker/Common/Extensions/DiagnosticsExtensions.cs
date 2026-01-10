using Microsoft.Extensions.Logging;
using System.Diagnostics;

namespace BotWorker.Common.Extensions
{
    public static class DiagnosticsExtensions
    {
        // ��ʱ����װ��ִ�д��벢�����ʱ��־
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
                logger.LogInformation($"{actionName} ��ʱ: {sw.ElapsedMilliseconds} ms");
            }
        }

        // ��ʱ����װ�첽
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
                logger.LogInformation($"{actionName} ��ʱ: {sw.ElapsedMilliseconds} ms");
            }
        }
    }
}


