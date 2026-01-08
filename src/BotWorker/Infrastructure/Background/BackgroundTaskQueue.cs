using System.Threading.Channels;

namespace BotWorker.Infrastructure.Background
{
    public class BackgroundTaskQueue : IBackgroundTaskQueue
    {
        private readonly Channel<Func<CancellationToken, Task>> _queue;

        public BackgroundTaskQueue(int capacity = 100)
        {
            // Bounded channel for limited queue size
            var options = new BoundedChannelOptions(capacity)
            {
                FullMode = BoundedChannelFullMode.Wait
            };
            _queue = Channel.CreateBounded<Func<CancellationToken, Task>>(options);
        }

        public void QueueBackgroundWorkItem(Func<CancellationToken, Task> workItem)
        {
            ArgumentNullException.ThrowIfNull(workItem);

            _queue.Writer.TryWrite(workItem);
        }

        public async Task<Func<CancellationToken, Task>> DequeueAsync(CancellationToken cancellationToken)
        {
            var workItem = await _queue.Reader.ReadAsync(cancellationToken);
            return workItem;
        }
    }

}
