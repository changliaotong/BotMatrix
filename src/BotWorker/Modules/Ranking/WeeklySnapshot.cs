namespace BotWorker.Domain.Entities.Ranking
{
    public class WeeklySnapshot
    {
        public string Key { get; set; } = string.Empty;
        public long GroupId { get; set; }
        public string WeekKey { get; set; } = string.Empty;
        public List<RankingEntry> TopUsers { get; set; } = new();
        public DateTime ArchivedAt { get; set; } = DateTime.Now;
    }
}
