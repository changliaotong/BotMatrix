using System.Collections.Concurrent;
using System.Threading.Tasks;

namespace BotWorker.Application.Services
{
    public interface IBotStatsService
    {
        void RecordMessage(long botUin, bool isGroup);
        Task<long> GetTotalMessagesAsync();
    }

    public class BotStatsService : IBotStatsService
    {
        private long _totalMessages = 0;

        public void RecordMessage(long botUin, bool isGroup)
        {
            Interlocked.Increment(ref _totalMessages);
        }

        public async Task<long> GetTotalMessagesAsync()
        {
            return await Task.FromResult(Interlocked.Read(ref _totalMessages));
        }
    }
}


