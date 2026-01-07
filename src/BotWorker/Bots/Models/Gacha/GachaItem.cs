namespace BotWorker.Bots.Models.Gacha
{
    public class GachaItem
    {
        public string Name { get; set; } = default!;
        public string Rarity { get; set; } = "R"; // SSR / SR / R
        public double Probability { get; set; } // 例如 0.01 表示 1%
    }

}
