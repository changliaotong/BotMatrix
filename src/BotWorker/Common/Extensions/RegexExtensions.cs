using System.Text.RegularExpressions;

namespace BotWorker.BotWorker.BotWorker.Common.Exts
{
    public static class RegexExtensions
    {
        public static string MatchGroup(this string input, string pattern, string groupName)
        {
            var match = Regex.Match(input, pattern);
            return match.Success && match.Groups[groupName].Success ? match.Groups[groupName].Value : "";
        }

        public static IEnumerable<string> FindAllUrls(this string text)
        {
            return Regex.Matches(text, @"https?://[^\s]+").Select(m => m.Value);
        }
    }
}


