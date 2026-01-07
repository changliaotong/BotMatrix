namespace BotWorker.Common.Exts
{
    public static class StringExtensions
    {
        public static bool ContainsAny(this string input, params string[] keywords)
        => keywords.Any(k => input?.Contains(k, StringComparison.OrdinalIgnoreCase) == true);

        public static string RemoveEmoji(this string text)
            => text.RegexReplace(@"[\uD800-\uDFFF]", string.Empty);

        public static string? NormalizeMessage(this string text)
            => text?.Trim().Replace("\r", "").Replace("\n", " ").ToLowerInvariant();

        public static string[] SplitWords(this string text)
            => text?.Split([' ', '\t'], StringSplitOptions.RemoveEmptyEntries) ?? [];

        public static bool IsNullOrWhiteSpace(this string? str) =>
            string.IsNullOrWhiteSpace(str);

        public static string ToSlug(this string text)
            => string.Join("-", text.ToLower().Split(' ', StringSplitOptions.RemoveEmptyEntries));

        public static string Truncate(this string str, int length, string suffix = "...") =>
            str.Length <= length ? str : str[..length] + suffix;

        // 移除字符串中所有空白字符（包括空格、换行等）
        public static string RemoveWhitespace(this string str)
        {
            if (str == null) return string.Empty;
            return new string(str.Where(c => !char.IsWhiteSpace(c)).ToArray());
        }

        // 驼峰转下划线
        public static string ToSnakeCase(this string input) =>
            string.Concat(input.Select((c, i) =>
                i > 0 && char.IsUpper(c) ? "_" + char.ToLower(c) : char.ToLower(c).ToString()));

        public static string ToBase64(this string str)
        {
            if (str == null) return string.Empty;
            var bytes = System.Text.Encoding.UTF8.GetBytes(str);
            return Convert.ToBase64String(bytes);
        }

        public static string FromBase64(this string base64)
        {
            if (string.IsNullOrEmpty(base64)) return string.Empty;
            try
            {
                var bytes = Convert.FromBase64String(base64);
                return System.Text.Encoding.UTF8.GetString(bytes);
            }
            catch
            {
                return string.Empty;
            }
        }
    }

}
