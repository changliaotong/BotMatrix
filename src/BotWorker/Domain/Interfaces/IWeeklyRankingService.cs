using System;
using System.Collections.Generic;

namespace BotWorker.Domain.Interfaces
{
    public class WeeklySnapshot
    {
        public string Key { get; set; } = string.Empty;
        public long GroupId { get; set; }
        public string WeekKey { get; set; } = string.Empty;
        public List<RankingEntry> TopUsers { get; set; } = new();
        public DateTime ArchivedAt { get; set; } = DateTime.Now;
    }

    public interface IWeeklyRankingService
    {
        void Increment(string key, string groupId, long userId, int delta = 1);

        List<RankingEntry> GetTop(string key, string groupId, string weekKey = "", int topN = 10);

        void ResetWeek(string weekKey = "");

        List<WeeklySnapshot> GetLastWeekTop(string key, string groupId, int topN = 10);
    }
}


