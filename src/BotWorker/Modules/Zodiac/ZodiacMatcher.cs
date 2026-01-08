namespace BotWorker.Domain.Entities.Zodiac
{
    public static class ZodiacMatcher
    {
        private static readonly Dictionary<string, (string Best, string[] Good, string[] Bad)> Matches = new()
        {
            ["白羊座"] = ("狮子座", new[] { "射手座", "天秤座" }, new[] { "巨蟹座", "摩羯座" }),
            ["金牛座"] = ("处女座", new[] { "摩羯座", "巨蟹座" }, new[] { "狮子座", "水瓶座" }),
            ["双子座"] = ("天秤座", new[] { "水瓶座", "白羊座" }, new[] { "处女座", "双鱼座" }),
            ["巨蟹座"] = ("天蝎座", new[] { "双鱼座", "金牛座" }, new[] { "白羊座", "天秤座" }),
            ["狮子座"] = ("射手座", new[] { "白羊座", "双子座" }, new[] { "金牛座", "天蝎座" }),
            ["处女座"] = ("金牛座", new[] { "摩羯座", "巨蟹座" }, new[] { "双子座", "射手座" }),
            ["天秤座"] = ("双子座", new[] { "水瓶座", "狮子座" }, new[] { "巨蟹座", "摩羯座" }),
            ["天蝎座"] = ("巨蟹座", new[] { "双鱼座", "处女座" }, new[] { "白羊座", "狮子座" }),
            ["射手座"] = ("白羊座", new[] { "狮子座", "天秤座" }, new[] { "处女座", "双鱼座" }),
            ["摩羯座"] = ("金牛座", new[] { "处女座", "双鱼座" }, new[] { "白羊座", "天秤座" }),
            ["水瓶座"] = ("双子座", new[] { "天秤座", "射手座" }, new[] { "金牛座", "巨蟹座" }),
            ["双鱼座"] = ("巨蟹座", new[] { "天蝎座", "摩羯座" }, new[] { "双子座", "射手座" }),
        };

        public static string GetMatchInfo(string zodiac1, string zodiac2)
        {
            if (!Matches.ContainsKey(zodiac1)) return "未知星座配对";

            var match = Matches[zodiac1];
            if (match.Best == zodiac2)
                return "最佳搭配，天作之合！";
            if (match.Good.Contains(zodiac2))
                return "相当般配，配合默契。";
            if (match.Bad.Contains(zodiac2))
                return "配对较差，小心摩擦。";

            return "普通搭配，互相理解很重要。";
        }
    }

}
