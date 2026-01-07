namespace BotWorker.Common.Exts
{
    public static class RetryHelper
    {
        // 默认重试间隔（秒）
        private static readonly int[] DefaultRetryDelaysSeconds =
        [
            1, 2, 4, 8, 16, 30, 60, 120, 300, 600, 900, 1800
        ];

        /// <summary>
        /// 带可选自定义重试间隔（秒）的同步重试执行方法。
        /// 当执行的 <paramref name="action"/> 抛出异常时，会根据 <paramref name="retryDelaysSeconds"/> 中定义的时间间隔依次等待并重试，
        /// 如果超过最大重试次数仍未成功，异常会被抛出至调用者。
        /// </summary>
        /// <param name="action">需要执行且可能失败的操作委托。</param>
        /// <param name="retryDelaysSeconds">重试间隔数组（单位：秒），默认采用指数退避策略，例如 [1, 2, 4, 8, 16]。</param>
        /// <param name="onRetryDelay">每次重试等待前的回调，参数为当前等待秒数，可用于日志或界面提示，默认为 null。</param>
        public static void Retry(
            Action action,
            int[]? retryDelaysSeconds = null,
            Action<int>? onRetryDelay = null)
        {
            ArgumentNullException.ThrowIfNull(action);

            var delays = retryDelaysSeconds ?? DefaultRetryDelaysSeconds;

            int attempt = 0;
            while (true)
            {
                try
                {
                    action();
                    return;
                }
                catch
                {
                    if (attempt >= delays.Length)
                        throw;

                    int delay = delays[attempt];
                    onRetryDelay?.Invoke(delay);

                    Thread.Sleep(delay * 1000);
                    attempt++;
                }
            }
        }

        public static async Task RetryAsync(Func<Task> action, int[] delaysInSeconds, Action<int>? onRetry = null)
        {
            foreach (var delay in delaysInSeconds)
            {
                try
                {
                    await action();
                    return;
                }
                catch
                {
                    onRetry?.Invoke(delay);
                    await Task.Delay(TimeSpan.FromSeconds(delay));
                }
            }

            throw new Exception("重试失败，已耗尽重试次数");
        }
    }

}
