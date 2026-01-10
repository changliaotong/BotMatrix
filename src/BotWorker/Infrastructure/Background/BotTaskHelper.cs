using Microsoft.Extensions.DependencyInjection;
using Microsoft.Extensions.Logging;
using BotWorker.Domain.Models.Messages.BotMessages;
using BotWorker.Domain.Interfaces;

namespace BotWorker.Infrastructure.Background
{
    public static class BotTaskHelper
    {
        public static void EnqueueBotTask(
            IServiceProvider provider,
            BotMessage context,
            Func<BotMessage, Task> action,
            ILogger logger,
            string taskName = "BotTask")
        {
            var queue = provider.GetRequiredService<IBackgroundTaskQueue>();

            queue.QueueBackgroundWorkItem(async token =>
            {
                try
                {
                    await action(context);
                }
                catch (Exception ex)
                {
                    logger.LogError(ex, "[BotTask] 异步任务 [{taskName}] 执行异常", taskName);
                }
            });
        }
    }

}
