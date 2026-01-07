namespace BotWorker.Bots.Models.Ranking
{
    public interface IWeeklyRankingService
    {
        void Increment(string key, string groupId, long userId, int delta = 1);

        List<RankingEntry> GetTop(string key, string groupId, string weekKey = "", int topN = 10);

        void ResetWeek(string weekKey = "");

        List<WeeklySnapshot> GetLastWeekTop(string key, string groupId, int topN = 10);
    }
}
