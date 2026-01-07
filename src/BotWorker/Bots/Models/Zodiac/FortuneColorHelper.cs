namespace sz84.Bots.Models.Zodiac
{
    public static class FortuneColorHelper
    {
        private static readonly Dictionary<string, string> ColorMap = new()
        {
            ["大吉"] = "#FFD700",  // 金色
            ["吉"] = "#32CD32",    // 绿色
            ["平"] = "#808080",    // 灰色
            ["凶"] = "#FF4500",    // 橙红色
            ["大凶"] = "#8B0000"   // 深红色
        };

        public static string GetColor(string fortuneLevel)
        {
            return ColorMap.TryGetValue(fortuneLevel, out var color) ? color : "#000000"; // 默认黑色
        }
    }

}
