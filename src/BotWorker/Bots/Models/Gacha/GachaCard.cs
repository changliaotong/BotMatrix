namespace sz84.Bots.Models.Gacha
{
    public class GachaCard
    {
        public int Id { get; set; }
        public string Name { get; set; } = default!;
        public int Rarity { get; set; }              // 稀有度：1~5
        public string Category { get; set; } = "普通";
        public string? ImageUrl { get; set; }
        public string? Description { get; set; }
    }

    public enum Rarity
    {
        Common = 1,
        Rare = 2,
        Legendary = 3
    }

    public class GachaCardRecord
    {
        public int Id { get; set; }
        public long UserId { get; set; }
        public int CardId { get; set; }
        public int Count { get; set; } = 1;
        public DateTime ObtainedAt { get; set; } = DateTime.Now;

        public GachaCard? Card { get; set; }
    }

    public class GachaDrawLog
    {
        public int Id { get; set; }
        public long UserId { get; set; }
        public int CardId { get; set; }
        public string Source { get; set; } = "normal"; // "free"/"daily"/"event"
        public DateTime Timestamp { get; set; } = DateTime.Now;
    }

}