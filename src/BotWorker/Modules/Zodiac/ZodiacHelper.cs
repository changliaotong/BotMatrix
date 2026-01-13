namespace BotWorker.Domain.Entities.Zodiac
{
    public static class ZodiacHelper
    {
        // 星座日期区间（不含年份）
        private static readonly (int startMonth, int startDay, int endMonth, int endDay, string name)[] ZodiacRanges = new[]
        {
        (3, 21, 4, 19, "白羊座"),
        (4, 20, 5, 20, "金牛座"),
        (5, 21, 6, 21, "双子座"),
        (6, 22, 7, 22, "巨蟹座"),
        (7, 23, 8, 22, "狮子座"),
        (8, 23, 9, 22, "处女座"),
        (9, 23, 10, 23, "天秤座"),
        (10, 24, 11, 22, "天蝎座"),
        (11, 23, 12, 21, "射手座"),
        (12, 22, 1, 19, "摩羯座"),
        (1, 20, 2, 18, "水瓶座"),
        (2, 19, 3, 20, "双鱼座"),
    };

        public static string GetZodiac(DateTime birthday)
        {
            foreach (var (startM, startD, endM, endD, name) in ZodiacRanges)
            {
                if (birthday.Month == startM && birthday.Day >= startD || birthday.Month == endM && birthday.Day <= endD ||
                    startM > endM && (birthday.Month == startM && birthday.Day >= startD || birthday.Month == endM && birthday.Day <= endD))
                    return name;
            }
            return "未知星座";
        }
    }

}
