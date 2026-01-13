using System.Text.RegularExpressions;

namespace BotWorker.Common.Utily
{
    public static class MarkdownCleaner
    {
        public static string StripMarkdown(this string text)
        {
            if (string.IsNullOrEmpty(text)) return text;

            // 去除标题标记 #
            text = text.RegexReplace(@"^#+\s*", "", RegexOptions.Multiline);

            // 粗体和斜体 **text** 或 *text* 或 __text__ 或 _text_
            text = text.RegexReplace(@"(\*\*|__)(.*?)\1", "$2");
            text = text.RegexReplace(@"(\*|_)(.*?)\1", "$2");

            // 行内代码 `code`
            text = text.RegexReplace(@"`([^`]*)`", "$1");

            // 多行代码块 ```code```
            text = text.RegexReplace(@"```[\s\S]*?```", "", RegexOptions.Multiline);

            // 图片 ![alt](url)
            text = text.RegexReplace(@"!\[(.*?)\]\((.*?)\)", "");

            // 链接 [text](url) → text
            text = text.RegexReplace(@"\[(.*?)\]\((.*?)\)", "$1");

            // 引用 > 
            text = text.RegexReplace(@"^\s*>+\s*", "", RegexOptions.Multiline);

            // 列表符号 - * + 或数字开头
            text = text.RegexReplace(@"^\s*([-*+]|\d+\.)\s+", "", RegexOptions.Multiline);

            // 去除多余的 markdown 特殊符号
            text = text.RegexReplace(@"[*_`#>\[\]!\(\)]", "");

            return text;
        }
    }
}
