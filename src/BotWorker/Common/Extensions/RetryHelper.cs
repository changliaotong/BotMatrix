namespace BotWorker.Common.Extensions
{
    public static class RetryHelper
    {
        // Ĭ�����Լ����룩
        private static readonly int[] DefaultRetryDelaysSeconds =
        [
            1, 2, 4, 8, 16, 30, 60, 120, 300, 600, 900, 1800
        ];

        /// <summary>
        /// ����ѡ�Զ������Լ����룩��ͬ������ִ�з�����
        /// ��ִ�е� <paramref name="action"/> �׳��쳣ʱ������� <paramref name="retryDelaysSeconds"/> �ж����ʱ�������εȴ������ԣ�
        /// �������������Դ�����δ�ɹ����쳣�ᱻ�׳��������ߡ�
        /// </summary>
        /// <param name="action">��Ҫִ���ҿ���ʧ�ܵĲ���ί�С�</param>
        /// <param name="retryDelaysSeconds">���Լ�����飨��λ���룩��Ĭ�ϲ���ָ���˱ܲ��ԣ����� [1, 2, 4, 8, 16]��</param>
        /// <param name="onRetryDelay">ÿ�����Եȴ�ǰ�Ļص�������Ϊ��ǰ�ȴ���������������־�������ʾ��Ĭ��Ϊ null��</param>
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

            throw new Exception("����ʧ�ܣ��Ѻľ����Դ���");
        }
    }

}


