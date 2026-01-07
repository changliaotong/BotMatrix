namespace sz84.Bots.Models.Gacha
{
    public class GachaCollection
    {
        public int Id { get; set; }
        public long UserId { get; set; }
        public string CardName { get; set; } = default!;
        public Rarity Rarity { get; set; }
        public int Count { get; set; }
    }
}
