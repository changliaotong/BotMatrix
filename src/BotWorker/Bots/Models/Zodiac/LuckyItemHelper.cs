namespace sz84.Bots.Models.Zodiac
{
    public static class LuckyItemHelper
    {
        private static readonly Dictionary<string, string[]> ItemsMap = new()
        {
            ["大吉"] = new[] { "四叶草", "金元宝", "红珊瑚" },
            ["吉"] = new[] { "幸运星", "翡翠吊坠", "兔子脚" },
            ["平"] = new[] { "黑曜石", "风铃", "绿松石" },
            ["凶"] = new[] { "护身符", "护符", "白水晶" },
            ["大凶"] = new[] { "平安符", "黑曜石戒指", "守护石" }
        };

        private static readonly Random _rand = new();

        public static string GetLuckyItem(string fortuneLevel)
        {
            if (!ItemsMap.ContainsKey(fortuneLevel)) return "幸运物品";

            var items = ItemsMap[fortuneLevel];
            return items[_rand.Next(items.Length)];
        }
    }

}
