namespace BotWorker.Bots.Models.Ranking
{
    public interface IRankingService
    {
        List<RankingEntry> GetTop(string key, string groupId, int topN = 10);

        void UpdateValue(string key, string groupId, long userId, int value);

        void Increment(string key, string groupId, long userId, int delta = 1);
    }

}
