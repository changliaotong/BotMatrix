using System.Collections.Generic;

namespace BotWorker.Domain.Interfaces
{
    public class RankingEntry
    {
        public long UserId { get; set; }
        public int Value { get; set; }
    }

    public interface IRankingService
    {
        List<RankingEntry> GetTop(string key, string groupId, int topN = 10);

        void UpdateValue(string key, string groupId, long userId, int value);

        void Increment(string key, string groupId, long userId, int delta = 1);
    }
}


