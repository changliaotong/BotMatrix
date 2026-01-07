namespace BotWorker.Common.Exts
{
    public static class DateTimeExtensions
    {
        public static string ToRelativeTime(this DateTime dt)
        {
            var span = DateTime.Now - dt;
            if (span.TotalMinutes < 1) return "刚刚";
            if (span.TotalHours < 1) return $"{(int)span.TotalMinutes}分钟前";
            if (span.TotalDays < 1) return $"{(int)span.TotalHours}小时前";
            return $"{(int)span.TotalDays}天前";
        }
        public static long ToUnixTimestamp(this DateTime dt)
        {
            var offset = new DateTimeOffset(dt.ToUniversalTime());
            return offset.ToUnixTimeSeconds();
        }

        public static DateTime ToStartOfDay(this DateTime dt)
            => dt.Date;

        // 简单的时间差“人类化”表达，如 “2小时前”，“5天前”
        public static string HumanizeTimeSpan(this TimeSpan ts)
        {
            if (ts.TotalDays >= 1)
                return $"{(int)ts.TotalDays} 天前";
            if (ts.TotalHours >= 1)
                return $"{(int)ts.TotalHours} 小时前";
            if (ts.TotalMinutes >= 1)
                return $"{(int)ts.TotalMinutes} 分钟前";
            return "刚刚";
        }

        public static bool IsWeekend(this DateTime dt)
            => dt.DayOfWeek == DayOfWeek.Saturday || dt.DayOfWeek == DayOfWeek.Sunday;

        // 返回指定日期所在周的起止日期（周一到周日）
        public static (DateTime Start, DateTime End) GetWeekRange(this DateTime dt)
        {
            int diff = (7 + (dt.DayOfWeek - DayOfWeek.Monday)) % 7;
            var start = dt.AddDays(-diff).Date;
            var end = start.AddDays(6).Date;
            return (start, end);
        }

        // 是否为同一天
        public static bool IsSameDay(this DateTime dt1, DateTime dt2)
        {
            return dt1.Date == dt2.Date;
        }

        // 获取某天的开始/结束时间
        public static DateTime StartOfDay(this DateTime dt) => dt.Date;
        public static DateTime EndOfDay(this DateTime dt) => dt.Date.AddDays(1).AddTicks(-1);
    }
}
