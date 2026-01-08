using System.Text;
using System.Text.RegularExpressions;

namespace BotWorker.Common.Extensions
{
    public static class Extensions
    {
        // ---------- string 扩展 ----------

        public static bool IsNullOrBlank(this string input)
        {
            return string.IsNullOrWhiteSpace(input);
        }

        public static int ToInt(this string input, int defaultValue = 0)
        {
            return int.TryParse(input, out var result) ? result : defaultValue;
        }

        public static bool ToBool(this string input, bool defaultValue = false)
        {
            if (input.IsNullOrBlank()) return defaultValue;
            if (input == "1") return true;
            if (input == "0") return false;
            return bool.TryParse(input, out var result) ? result : defaultValue;
        }

        public static DateTime ToDateTime(this string input, DateTime? defaultValue = null)
        {
            return DateTime.TryParse(input, out var result) ? result : (defaultValue ?? DateTime.MinValue);
        }

        public static string StripHtml(this string input)
        {
            if (input == null) return string.Empty;
            return Regex.Replace(input, "<.*?>", string.Empty);
        }

        public static string Truncate(this string input, int maxLength, string ellipsis = "…")
        {
            if (input == null) return string.Empty;
            return input.Length <= maxLength ? input : input[..maxLength] + ellipsis;
        }

        public static string Capitalize(this string input)
        {
            if (string.IsNullOrWhiteSpace(input)) return input;
            return char.ToUpper(input[0]) + input.Substring(1);
        }

        // ---------- long 扩展（时间戳） ----------

        public static DateTime FromUnixSeconds(this long timestamp)
        {
            return DateTimeOffset.FromUnixTimeSeconds(timestamp).LocalDateTime;
        }

        public static DateTime FromUnixMilliseconds(this long timestamp)
        {
            return DateTimeOffset.FromUnixTimeMilliseconds(timestamp).LocalDateTime;
        }

        public static string ToReadableFileSize(this long bytes)
        {
            string[] sizes = { "B", "KB", "MB", "GB", "TB" };
            double len = bytes;
            int order = 0;
            while (len >= 1024 && order < sizes.Length - 1)
            {
                order++;
                len /= 1024;
            }
            return $"{len:0.##} {sizes[order]}";
        }

        // ---------- DateTime 扩展 ----------

        public static long ToUnixSeconds(this DateTime dt)
        {
            return new DateTimeOffset(dt).ToUnixTimeSeconds();
        }

        public static long ToUnixMilliseconds(this DateTime dt)
        {
            return new DateTimeOffset(dt).ToUnixTimeMilliseconds();
        }

        // ---------- 文件操作扩展 ----------

        public static string ReadFileText(this string filePath)
        {
            return File.Exists(filePath) ? File.ReadAllText(filePath, Encoding.UTF8) : string.Empty;
        }

        public static void WriteFileText(this string filePath, string content, bool append = false)
        {
            var dir = Path.GetDirectoryName(filePath);
            if (!string.IsNullOrEmpty(dir) && !Directory.Exists(dir))
            {
                Directory.CreateDirectory(dir);
            }

            if (append)
                File.AppendAllText(filePath, content, Encoding.UTF8);
            else
                File.WriteAllText(filePath, content, Encoding.UTF8);
        }
    }
}