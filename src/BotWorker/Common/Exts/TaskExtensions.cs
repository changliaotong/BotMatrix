namespace BotWorker.Common.Exts
{
    public static class TaskExtensions
    {
        public static async Task RetryAsync(this Func<Task> action, int retries = 3, int delayMs = 200)
        {
            for (int i = 0; i < retries; i++)
            {
                try { await action(); return; }
                catch when (i < retries - 1) { await Task.Delay(delayMs); }
            }
        }

        public static async Task<T> WithTimeout<T>(this Task<T> task, TimeSpan timeout)
        {
            if (await Task.WhenAny(task, Task.Delay(timeout)) == task)
                return await task;
            throw new TimeoutException();
        }

        // 火并忘记，安全捕获异常
        public static async void FireAndForgetSafeAsync(this Task task, Action<Exception>? onException = null)
        {
            try
            {
                await task;
            }
            catch (Exception ex)
            {
                onException?.Invoke(ex);
            }
        }

        // 忽略任务异常（适用于 fire-and-forget）
        public static void FireAndForget(this Task task, Action<Exception>? onError = null)
        {
            task.ContinueWith(t => onError?.Invoke(t.Exception!), TaskContinuationOptions.OnlyOnFaulted);
        }

        // 添加取消支持
        public static async Task<T> WithCancellation<T>(this Task<T> task, CancellationToken token)
        {
            using var reg = token.Register(() => throw new OperationCanceledException(token));
            return await task;
        }
    }
}
