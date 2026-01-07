using System.Net;
using System.Text.RegularExpressions;

namespace BotWorker.Common.Exts
{
    public static class ExtAsHtml
    {
        // HTML高亮显示关键字
        public static string HighLightHtml(this string? text, string keyword)
        {
            if (string.IsNullOrWhiteSpace(text))
                return string.Empty;

            if (string.IsNullOrWhiteSpace(keyword))
                return text;

            // 多关键词处理
            if (keyword.Contains('%'))
            {
                var keys = keyword.Split('%', StringSplitOptions.RemoveEmptyEntries | StringSplitOptions.TrimEntries);
                string res = text;
                foreach (var key in keys)
                {
                    res = res.HighLightHtml(key); // 递归处理每个关键词
                }
                return res;
            }

            // 安全替换，仅替换标签外的内容
            return Regex.Replace(
                text,
                $@"(?<=>)([^<]*?)({Regex.Escape(keyword)})([^<]*?)(?=<|$)", // 只替换标签外的文本
                m => $"{m.Groups[1].Value}<span class='highlight'>{m.Groups[2].Value}</span>{m.Groups[3].Value}",
                RegexOptions.IgnoreCase | RegexOptions.Compiled
            );
        }


        public static string HighlightText(this string text, string highlight)
        {
            if (string.IsNullOrEmpty(highlight)) return text;

            // 转义高亮文本中的特殊字符
            var escapedHighlight = EscapeSpecialChars(highlight);

            // 创建正则表达式来匹配高亮文本
            var regex = new Regex(escapedHighlight, RegexOptions.IgnoreCase | RegexOptions.Compiled);

            // 对整个文本进行 HTML 实体转义，防止干扰
            string encodedText = HtmlEncode(text);

            // 替换匹配的文本并添加高亮
            string highlightedText = regex.Replace(encodedText, match => $"<span class='highlight'>{match.Value}</span>");

            // 恢复 HTML 实体为原始字符
            var res = WebUtility.HtmlDecode(highlightedText);
            return res;
        }

        // 对特殊字符进行转义
        private static string EscapeSpecialChars(string input)
        {
            if (string.IsNullOrEmpty(input)) return input;

            return Regex.Escape(input) // 使用 Regex.Escape 自动处理特殊字符
                .Replace("'", "\\'")    // 单独处理单引号
                .Replace("\"", "\\\""); // 单独处理双引号
        }

        // HTML 实体转义方法
        private static string HtmlEncode(string text)
        {
            if (string.IsNullOrEmpty(text)) return text;
            return WebUtility.HtmlEncode(text);
        }

        // 定义一个新的方法，将高亮的部分反转义
        public static string HighlightAndDecode(string inputText, string highlightText)
        {
            if (string.IsNullOrEmpty(highlightText)) return inputText;

            // 对高亮文本进行简单查找和替换，添加高亮标签
            string encodedHighlight = WebUtility.HtmlEncode(highlightText);

            // 反转义高亮内容，并使用 HTML <span> 标签包裹
            string decodedHighlight = $"<span class=\"highlight\">{WebUtility.HtmlDecode(encodedHighlight)}</span>";

            // 将高亮内容替换到原文本中
            return inputText.Replace(encodedHighlight, decodedHighlight);
        }

        public static string ParseMarkdownToHtml(string markdown)
        {
            // 处理标题并转换为粗体显示
            string html = markdown.RegexReplace(@"^#+\s*(.+)", match => { return $"<strong>{match.Groups[1].Value.Trim()}</strong>"; }, RegexOptions.Multiline);

            // 处理分割线
            html = html.RegexReplace(@"^\s*(---|\*\*\*|___)\s*$", "<hr />", RegexOptions.Multiline);

            // 处理粗体
            html = html.RegexReplace(@"\*\*(.+?)\*\*", "<strong>$1</strong>");

            html = html.RegexReplace(@"```(\w+)?\s*([\s\S]*?)```", m =>
            {
                var codeType = string.IsNullOrEmpty(m.Groups[1].Value) ? "code" : m.Groups[1].Value;
                return $"<pre class='code-block'><div class='code-type'>{codeType}</div><code>{m.Groups[2].Value}</code></pre>";
            }, RegexOptions.Singleline);

            // 处理行内代码
            html = html.RegexReplace(@"`([^`]+)`", "<code>$1</code>");

            // 处理引用
            html = html.RegexReplace(@"^\>\s*(.+)", "<blockquote>$1</blockquote>", RegexOptions.Multiline);

            // 处理链接
            html = html.RegexReplace(@"\[(.+?)\]\((.+?)\)", "<a href=\"$2\">$1</a>");

            return html;
        }

        // 改进后的正则表达式模式，确保 URL 不在 <a> 标签中
        public const string UrlPattern = @"(?<!<a\s[^>]*?href=[""'])((https?|ftp|file)://[-A-Za-z0-9+&@#/%?=~_|!:,.;\()]+[-A-Za-z0-9+&@#/%=~_|()])(?![""'][^>]*?>)";

        public static string UrlEvaluator(Match match)
        {
            string url = match.Value;

            return $"<a href=\"{url}\" target=\"_blank\">{(url.Length > 50 ? $"{url[..50]}..." : url)}</a>";
        }

        public static string HighLight(this string text, string hightLight) => HighlightText(text, hightLight);

        // 替换文本中的 换行、qq表情、url、点歌网址 以适合网页显示
        public static string AsHtml(this string? text, bool isAI = false, string hightLight = "")
        {
            if (text == null) return "";
            if (text.StartsWith("https://dalleprodaue.blob.core.windows.net"))
                text = $"<img src=\"{text}\" width=\"600\" alt=\"图片\">";
            else if (isAI)
            {
                // 先进行 HTML 编码
                text = WebUtility.HtmlEncode(text);

                if (!hightLight.IsNull())
                {
                    // 高亮处理，且对高亮部分进行反转义
                    text = HighlightAndDecode(text, hightLight);
                }

                // 然后将 Markdown 转换为 HTML
                text = ParseMarkdownToHtml(text);
            }
            else
            {
                text = text.RegexReplace(UrlPattern, UrlEvaluator);
                text = text.Replace("\r\n", "<br />").Replace("\n\r", "<br />").Replace("\n", "<br />").Replace("\r", "<br />");
                text = text.RegexReplace("\\[Face(\\d+).gif]", "<img src=\"/images/face/$1.gif\">");
                text = HighlightAndDecode(text, hightLight);
            }
            return text;
        }
    }
}
