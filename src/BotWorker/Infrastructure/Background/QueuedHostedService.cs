using Microsoft.Extensions.Hosting;
using Microsoft.Extensions.Logging;
using BotWorker.Core.Interfaces;

namespace BotWorker.Infrastructure.Background
{
    public class QueuedHostedService(IBackgroundTaskQueue taskQueue, ILogger<QueuedHostedService> logger) : BackgroundService
    {
        private readonly IBackgroundTaskQueue _taskQueue = taskQueue;
        private readonly ILogger<QueuedHostedService> _logger = logger;
        private readonly int _concurrentWorkers = 16; 

        protected override Task ExecuteAsync(CancellationToken stoppingToken)
        {
            // 启动多个并发消费者
            var workers = Enumerable.Range(0, _concurrentWorkers).Select(_ => Task.Run(() => ConsumeQueue(stoppingToken)));
            return Task.WhenAll(workers);
        }

        private async Task ConsumeQueue(CancellationToken stoppingToken)
        {
            while (!stoppingToken.IsCancellationRequested)
            {
                try
                {
                    var workItem = await _taskQueue.DequeueAsync(stoppingToken);
                    await workItem(stoppingToken);
                }
                catch (Exception ex)
                {
                    _logger.LogError(ex, "后台任务执行失败");
                }
            }
        }
    }


}
