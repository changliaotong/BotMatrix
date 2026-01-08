namespace BotWorker.BotWorker.BotWorker.Common.Exts
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

        // �����ǣ���ȫ�����쳣
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

        // ���������쳣�������� fire-and-forget��
        public static void FireAndForget(this Task task, Action<Exception>? onError = null)
        {
            task.ContinueWith(t => onError?.Invoke(t.Exception!), TaskContinuationOptions.OnlyOnFaulted);
        }

        // ���ȡ��֧��
        public static async Task<T> WithCancellation<T>(this Task<T> task, CancellationToken token)
        {
            using var reg = token.Register(() => throw new OperationCanceledException(token));
            return await task;
        }
    }
}


