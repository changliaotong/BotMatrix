namespace sz84.Infrastructure.Utils
{
    class TimeHelper
    {       
        public static string ToFriendlyTime(DateTime dt)
        {
            var span = DateTime.Now - dt;
            if (span.TotalSeconds < 60) return "刚刚";
            if (span.TotalMinutes < 60) return $"{(int)span.TotalMinutes}分钟前";
            if (span.TotalHours < 24) return $"{(int)span.TotalHours}小时前";
            return dt.ToString("yyyy-MM-dd HH:mm");
        }
    }
}
