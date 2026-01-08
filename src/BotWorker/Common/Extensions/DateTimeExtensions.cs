namespace BotWorker.BotWorker.BotWorker.Common.Exts
{
    public static class DateTimeExtensions
    {
        public static string ToRelativeTime(this DateTime dt)
        {
            var span = DateTime.Now - dt;
            if (span.TotalMinutes < 1) return "�ո�";
            if (span.TotalHours < 1) return $"{(int)span.TotalMinutes}����ǰ";
            if (span.TotalDays < 1) return $"{(int)span.TotalHours}Сʱǰ";
            return $"{(int)span.TotalDays}��ǰ";
        }
        public static long ToUnixTimestamp(this DateTime dt)
        {
            var offset = new DateTimeOffset(dt.ToUniversalTime());
            return offset.ToUnixTimeSeconds();
        }

        public static DateTime ToStartOfDay(this DateTime dt)
            => dt.Date;

        // �򵥵�ʱ�����໯������ ��2Сʱǰ������5��ǰ��
        public static string HumanizeTimeSpan(this TimeSpan ts)
        {
            if (ts.TotalDays >= 1)
                return $"{(int)ts.TotalDays} ��ǰ";
            if (ts.TotalHours >= 1)
                return $"{(int)ts.TotalHours} Сʱǰ";
            if (ts.TotalMinutes >= 1)
                return $"{(int)ts.TotalMinutes} ����ǰ";
            return "�ո�";
        }

        public static bool IsWeekend(this DateTime dt)
            => dt.DayOfWeek == DayOfWeek.Saturday || dt.DayOfWeek == DayOfWeek.Sunday;

        // ����ָ�����������ܵ���ֹ���ڣ���һ�����գ�
        public static (DateTime Start, DateTime End) GetWeekRange(this DateTime dt)
        {
            int diff = (7 + (dt.DayOfWeek - DayOfWeek.Monday)) % 7;
            var start = dt.AddDays(-diff).Date;
            var end = start.AddDays(6).Date;
            return (start, end);
        }

        // �Ƿ�Ϊͬһ��
        public static bool IsSameDay(this DateTime dt1, DateTime dt2)
        {
            return dt1.Date == dt2.Date;
        }

        // ��ȡĳ��Ŀ�ʼ/����ʱ��
        public static DateTime StartOfDay(this DateTime dt) => dt.Date;
        public static DateTime EndOfDay(this DateTime dt) => dt.Date.AddDays(1).AddTicks(-1);
    }
}


