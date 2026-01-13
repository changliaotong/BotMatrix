using Microsoft.Extensions.Logging;

namespace BotWorker.Infrastructure.Tasks
{
    public static class AsyncHelper
    {
        public static void FireAndForget(Func<Task> taskFunc, ILogger logger, string taskName = "FireAndForget")
        {
            _ = Task.Run(async () =>
            {
                try
                {
                    await taskFunc();
                }
                catch (Exception ex)
                {
                    logger.LogError(ex, "[AsyncHelper] 异步任务 [{taskName}] 执行异常", taskName);
                }
            });
        }
    }

}
