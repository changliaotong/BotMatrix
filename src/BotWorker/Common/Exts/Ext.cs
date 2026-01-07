using System.Data;
using System.Diagnostics.CodeAnalysis;
using System.Reflection;
using System.Text;
using System.Text.RegularExpressions;

namespace BotWorker.Common.Exts
{
    public class Cov(string Name, object? value)
    {
        public string Name { get; set; } = Name;
        public object? Value { get; set; } = value;
    }

    public static class Ext
    {
        public static string MaskIdiom(this string idiom, string maskEmoji = "🔒")
        {
            if (string.IsNullOrWhiteSpace(idiom) || idiom.Length < 2)
                return idiom;

            int middleLength = idiom.Length - 2;
            string middle = string.Concat(Enumerable.Repeat(maskEmoji, middleLength));
            return idiom[0] + middle + idiom[^1];
        }

        public static string ReplaceSensitive(this string input, string regex)
        {
            return input.RegexReplace(regex, MaskWithChar());
        }

        public static MatchEvaluator MaskWithChar(string maskChar = "🔒")
        {
            return match => maskChar.Times(match.Length);
        }

        public static Dictionary<string, object?> ToFields(this object obj)
        {
            return obj.GetType()
                      .GetProperties(BindingFlags.Public | BindingFlags.Instance)
                      .Where(p => p.CanRead)
                      .Select(p => (p.Name, p.GetValue(obj)))
                      .ToDictionary();
        }

        private static readonly Random Random = new();  

        public static string GetRandom(this string[] strs)
        {
            return strs[Random.Next(strs.Length)];
        }

        public static string EnsureStartsWith(this string input, string prefix, StringComparison comparison = StringComparison.OrdinalIgnoreCase)
        {
            if (string.IsNullOrWhiteSpace(input))
            {
                return string.Empty;
            }

            // Normalize input to remove extra leading spaces
            string normalizedInput = input.TrimStart();

            // Create a regular expression pattern that allows for any whitespace between the words in the prefix
            // 更可靠的构造正则匹配前缀
            string pattern = "^" + string.Join(@"\s+", prefix.Split(' ').Select(Regex.Escape));

            // Determine the regex options based on StringComparison
            RegexOptions regexOptions = comparison == StringComparison.OrdinalIgnoreCase || comparison == StringComparison.InvariantCultureIgnoreCase
                ? RegexOptions.IgnoreCase
                : RegexOptions.None;

            // Check if the normalized input matches the pattern (using regex options for case sensitivity)
            if (normalizedInput.IsMatch(pattern, regexOptions))
            {
                return input;
            }

            // Add the prefix with a single space between ORDER and BY, and a space before the input
            return $"{prefix} {input}";
        }


        #region 基础操作

        public static string RestDays(this DateTime dt, DateTime dt2, out int years, out int months, out int days, out int hours, out int minutes, out int seconds)
        {
            var ts = dt > dt2 ? dt - dt2 : dt2 - dt;
            years = ts.Days / 365;
            months = ts.Days % 365 / 30;
            days = ts.Days % 365 % 30;
            hours = ts.Hours;
            minutes = ts.Minutes;
            seconds = ts.Seconds;
            return $"{years}年{months}月{days}天";
        }

        public static string RestDays(this DateTime dt, out int years, out int months, out int days, out int hours, out int minutes, out int seconds)
        {
            return dt.RestDays(DateTime.Now, out years, out months, out days, out hours, out minutes, out seconds);
        }

        public static string RestDays(this DateTime dt, out int years, out int months, out int days)
        {
            return dt.RestDays(out years, out months, out days, out _, out _, out _);
        }

        public static string RestDays(this DateTime dt)
        {
            return dt.RestDays(out _, out _, out _, out _, out _, out _);
        }

        public static long Max(params long[] nums)
        {
            return nums.Max(x => x);
        }

        public static long Min(params long[] nums)
        {
            return nums.Min(x => x);
        }

        public static string RemoveDup(this string text, string dups = " ")
        {
            if (text.IsNull())
                return string.Empty;
            else
            {
                string res = text.Trim();
                if (res.Contains(dups.Times(2)))
                {
                    res = res.Replace(dups.Times(2), dups);
                    res = res.RemoveDup(dups);
                }
                return res;
            }
        }

        private static readonly Dictionary<int, string> FaceEmojiMap = new Dictionary<int, string>
        {
            {0, "😄"}, {1, "😃"}, {2, "😀"}, {3, "😊"}, {4, "☺️"}, {5, "😉"},
            {6, "😍"}, {7, "😘"}, {8, "😚"}, {9, "😗"}, {10, "😙"}, {11, "😜"},
            {12, "😝"}, {13, "😛"}, {14, "😳"}, {15, "😁"}, {16, "😔"}, {17, "😌"},
            {18, "😒"}, {19, "😞"}, {20, "😣"}, {21, "😢"}, {22, "😂"}, {23, "😭"},
            {24, "😪"}, {25, "😥"}, {26, "😰"}, {27, "😅"}, {28, "😓"}, {29, "😩"},
            {30, "😫"}, {31, "😨"}, {32, "😱"}, {33, "😠"}, {34, "😡"}, {35, "😤"},
            {36, "😖"}, {37, "😆"}, {38, "😋"}, {39, "😷"}, {40, "😎"}, {41, "😴"},
            {42, "😵"}, {43, "😲"}, {44, "😟"}, {45, "😦"}, {46, "😧"}, {47, "😈"},
            {48, "👿"}, {49, "😮"}, {50, "😬"}, {51, "😐"}, {52, "😕"}, {53, "😯"},
            {54, "😶"}, {55, "😇"}, {56, "😏"}, {57, "😑"}, {58, "👲"}, {59, "👳‍♀️"},
            {60, "👮"}, {61, "👷"}, {62, "💂"}, {63, "👶"}, {64, "👦"}, {65, "👧"},
            {66, "👨"}, {67, "👩"}, {68, "👴"}, {69, "👵"}, {70, "👱"}, {71, "👼"},
            {72, "🎅"}, {73, "👸"}, {74, "👰"}, {75, "🙍"}, {76, "🙇"}, {77, "💁"},
            {78, "🙅"}, {79, "🙆"}, {80, "🙋"}, {81, "🙎"}, {82, "🙍‍♂️"}, {83, "💆"},
            {84, "💇"}, {85, "🚶"}, {86, "🏃"}, {87, "💃"}, {88, "⛷"}, {89, "🏂"},
            {90, "🏌"}, {91, "🏄"}, {92, "🚣"}, {93, "🏊"}, {94, "⛹"}, {95, "🏋"},
            {96, "🚴"}, {97, "🚵"}, {98, "🏎"}, {99, "🚓"}, {100, "🚑"}, {101, "🚒"},
            {102, "🚐"}, {103, "🚚"}, {104, "🚲"}
        };

        public static string ReplaceFaceWithEmoji(this string input)
        {
            if (string.IsNullOrEmpty(input)) return input;

            return input.RegexReplace(@"\[Face(\d+)\.gif\]", match =>
            {
                int index = int.Parse(match.Groups[1].Value);
                if (FaceEmojiMap.TryGetValue(index, out var emoji))
                    return emoji;
                else
                    return "🖼️"; // 默认替代，105-212
            });
        }

        /// <summary>
        /// 将频道 Emoji 转换为表情
        /// </summary>
        /// <param name="emoji"></param>
        /// <returns></returns>
        public static string ConvertEmojiToFace(this string emoji)
        {
            // Define the format for the face ID
            var faceFormat = "[Face{0}.gif]";

            // Use regular expressions to find and replace emoji IDs
            return RegexReplace(emoji, @"<emoji:(.*?)>", match =>
            {
                var emojiId = match.Groups[1].Value;
                return string.Format(faceFormat, emojiId);
            });
        }

        /// <summary>
        /// 将表情转换为频道 Emoji
        /// </summary>
        /// <param name="face"></param>
        /// <returns></returns>
        public static string ConvertFaceToEmoji(this string face)
        {
            // Define the format for the emoji ID
            var emojiFormat = "<emoji:{0}>";

            // Use regular expressions to find and replace face IDs
            return RegexReplace(face, @"\[Face(.*?)\.gif\]", match =>
            {
                var faceId = match.Groups[1].Value;
                return string.Format(emojiFormat, faceId);
            });
        }

        // 达到类似 python 中字符串*2的效果
        public static string Times(this string origin, int times = 2)
        {
            if (string.IsNullOrEmpty(origin) || times <= 0)
                return string.Empty;

            StringBuilder sb = new(origin.Length * times);
            for (int i = 0; i < times; i++)
            {
                sb.Append(origin);
            }
            return sb.ToString();
        }

        // 缓存 Regex 实例
        private static readonly Dictionary<string, Regex> _regexCache = [];

        public static List<string> RegexGetValues(this string? text, string regexText, string key)
        {
            if (string.IsNullOrWhiteSpace(text))
            {
                return []; // 返回空列表而不是null
            }

            // 获取或创建Regex实例
            Regex rx = _regexCache.GetValueOrDefault(regexText) ??
                       new Regex(regexText, RegexOptions.Compiled | RegexOptions.IgnoreCase);

            if (!_regexCache.ContainsKey(regexText))
            {
                _regexCache[regexText] = rx; // 缓存实例
            }

            // 获取匹配项
            MatchCollection matches = rx.Matches(text);
            return matches.Cast<Match>()
                          .Select(match => match.Groups[key].Value)
                          .ToList(); // 返回所有匹配项的列表
        }

        /// <summary>
        /// 根据正则表达式获取文本中匹配组的值。
        /// </summary>
        /// <param name="text">待匹配的文本</param>
        /// <param name="regexText">正则表达式</param>
        /// <param name="key">匹配组的名称</param>
        /// <returns>匹配组的最后一个值，如果没有匹配则返回null</returns>
        public static string RegexGetValue(this string? text, string regexText, string key)
        {
            if (string.IsNullOrWhiteSpace(text))
            {
                return string.Empty; // 返回null而不是空字符串，明确无匹配
            }

            // 获取或创建Regex实例
            Regex rx = _regexCache.GetValueOrDefault(regexText) ??
                       new Regex(regexText, RegexOptions.Compiled | RegexOptions.IgnoreCase);

            if (!_regexCache.ContainsKey(regexText))
            {
                _regexCache[regexText] = rx; // 缓存实例
            }

            // 获取匹配结果
            var matches = rx.Matches(text);
            if (matches.Count == 0)
            {
                return string.Empty; // 明确无匹配
            }

            // 获取最后一个匹配
            Match lastMatch = matches[matches.Count - 1];

            // 检查组是否存在且成功
            return lastMatch.Groups[key].Success ? lastMatch.Groups[key].Value : string.Empty;
        }


        // 得到@的QQ号码
        public static long GetAtUserId(this string? text)
        {
            return text.RegexGetValue(Regexs.User, "UserId").AsLong();
        }

        // 是否包含 Url
        public static bool ContainsURL(this string msg)
        {
            return msg.IsMatch(Regexs.Url) || msg.IsMatch(Regexs.Url2);
        }

        // 是否包含QQ号码
        public static bool HaveUserId(this string msg)
        {
            return msg.IsMatch(Regexs.HaveUserId);
        }

        // 是否数字(整数)
        public static bool IsNum(this string? text)
        {
            return long.TryParse(text, out _);
        }


        public static bool IsLong(this string? text)
        {
            return long.TryParse(text, out _);
        }


        public static bool IsDouble(this string? text)
        {
            return double.TryParse(text, out _);
        }

        public static bool IsDecimal(this string? origin)
        {
            return decimal.TryParse(origin, out _);
        }

        public static bool IsMatchQQ(this string? text)
        {
            return !text.IsNull() && text.IsMatch(Regexs.User);
        }

        public static bool IsHaveUserId(this string? text)
        {
            return !text.IsNull() && text.IsMatch(Regexs.Users);
        }

        public static string GetUrlData(this string? url)
        {
            HttpClient client = new();
            try
            {
                return client.GetAsync(url).Result.Content.ReadAsStringAsync().Result;
            }
            catch (Exception e)
            {
                Common.ErrorMessage($"\nException Caught!\nMessage :{e.Message}");
                return "";
            }
        }

        public static List<string> GetValueList(this string Text, string RegexText)
        {
            Regex rx = new(RegexText, RegexOptions.Compiled | RegexOptions.IgnoreCase);
            MatchCollection matches = rx.Matches(Text);

            List<string> res = [];
            foreach (Match match in matches.Cast<Match>())
            {
                res.Add(match.Value);
            }

            return res;
        }

        public static string RegexReplace(this string text, string pattern, MatchEvaluator evaluator, RegexOptions options = RegexOptions.Compiled | RegexOptions.IgnoreCase)
        {
            return Regex.Replace(text, pattern, evaluator, options);
        }

        public static string RegexReplace(this string text, string regexText, string replaceText, RegexOptions options = RegexOptions.Compiled | RegexOptions.IgnoreCase)
        {
             return Regex.Replace(text, regexText, replaceText, options);
        }

        public static string GetRegex(this string regex)
        {
            return regex.RegexReplace(@"(?<!\\)([+*])", "\\$1");
        }

        public static string ReplaceRegex(this string regexText)
        {
            return regexText.Replace("+", "\\+").Replace("*", "\\*").Replace(".", "\\.").Replace("?", "\\?").Replace("$", "\\$").Replace("(", "\\(").Replace(")", "\\)").Replace("[", "\\[").Replace("^", "\\^").Replace("{", "\\{");
        }

        public static string RemoveWhiteSpace(this string text)
        {
            return text.Replace(" ", "").Replace("　", "");
        }

        public static string RemoveBiaodian(this string text)
        {
            return Regex.Replace(text, Regexs.BiaoDian, "");
        }

        public static Match RegexMatch(this string text, string regex, RegexOptions options = RegexOptions.Compiled | RegexOptions.IgnoreCase | RegexOptions.Singleline)
        {
            return Regex.Match(text, regex, options);
        }

        public static bool IsMatch(this string text, string regex, RegexOptions options = RegexOptions.Compiled | RegexOptions.IgnoreCase | RegexOptions.Singleline)
        {
            return Regex.IsMatch(text, regex, options);
        }

        public static MatchCollection Matches(this string text, string regex, RegexOptions options = RegexOptions.Compiled | RegexOptions.IgnoreCase | RegexOptions.Singleline)
        {
            return new Regex(regex, options).Matches(text);
        }

        public static bool NotMatch(this string text, string regex, RegexOptions options = RegexOptions.Compiled | RegexOptions.IgnoreCase | RegexOptions.Singleline)
        {
            return !text.IsMatch(regex, options);
        }

        public static string MaskNo(this string? text, string mask = "*")
        {
            if (text == null)
                return "";
            if (text.Length > 6)
                return text[..3] + mask.Times(3) + text[^3..];
            else
                return text;
        }

        public static string WrapWord(this string text, int width)
        {
            return ImageGen.WordWrap(text, width);
        }

        public static bool IsNullOrEmpty(this string? text)
        {
            return string.IsNullOrEmpty(text);
        }

        public static bool IsNull([NotNullWhen(returnValue: false)] this string? text)
        {
            return string.IsNullOrEmpty(text) || string.IsNullOrWhiteSpace(text);
        }

        [SuppressMessage("Performance", "SYSLIB1045:转换为“GeneratedRegexAttribute”。", Justification = "<挂起>")]
        public static bool IsValidEmail(this string text)
        {
            return Regex.IsMatch(text, @"^([\w\.\-]+)@([\w\-]+)((\.(\w){2,3})+)$");
        }

        #endregion 基础操作

        public static int GetTokensCount(this string text)
        {
            // 每个汉字单独算一个 token
            string chinesePattern = @"[\u4e00-\u9fa5]";
            // 每个字母单独算一个 token
            string languagePattern = @"[\p{L}]";
            // 标点和 emoji
            string punctuationPattern = @"[^\w\s]|[\u201c\u201d\u2018\u2019]|[\uD800-\uDBFF][\uDC00-\uDFFF]";

            int tokenCount = 0;

            tokenCount += CountMatches(text.Matches(chinesePattern));
            tokenCount += CountMatches(text.Matches(languagePattern));
            tokenCount += CountMatches(text.Matches(punctuationPattern));

            return tokenCount;
        }

        private static int CountMatches(MatchCollection matches)
        {
            return matches.Count; // 返回匹配的 tokens 数量
        }

        // 假设你已有的两个 URL 正则
        // public static class Regexs { public const string Url = "..."; public const string Url2 = "..."; }
        // 我们这里用它们合并成一个 regex，避免重复替换同一段文本
        private static readonly Regex CombinedUrlRegex =
            new Regex($"({Regexs.Url})|({Regexs.Url2})", RegexOptions.Compiled | RegexOptions.IgnoreCase);

        // 核心：只替换非白名单 URL
        public static string BlockUrl(this string text, Func<string, bool> isWhiteListed)
        {
            if (string.IsNullOrEmpty(text)) return text;
            if (isWhiteListed == null) throw new ArgumentNullException(nameof(isWhiteListed));

            // 用一个回调逐个 Match 处理
            string evaluator(Match m)
            {
                var u = m.Value;

                // 有时候 Match 会包含前后括号或中文标点，去掉前后常见包裹字符再判断
                var trimmed = u.Trim(' ', '\u200B', '\r', '\n', '\t', '\"', '\'', '(', ')', '，', '。', '；', ',');

                try
                {
                    return isWhiteListed(trimmed) ? u : "[网址屏蔽]";
                }
                catch
                {
                    // 白名单判断出错时，保守处理为屏蔽
                    return "[网址屏蔽]";
                }
            }

            // 一次性通过合并正则替换，避免先后两个正则重复命中造成混乱
            return CombinedUrlRegex.Replace(text, new MatchEvaluator(evaluator));
        }


        // 敏感词过滤替换为方框
        public static string ReplaceInvalidKeyword(string msg)
        {
            string res = msg;
            if (!res.Contains(".sz84.com") && !res.Contains("i.y.qq.com/n2/m/playsong")
                && !res.Contains("music.163.com/#/song") && !res.Contains(".qq.com/51437810")
                && !res.Contains(".qq.com/1653346663") && !res.Contains("mp.weixin.qq.com/"))
            {
                MatchEvaluator myEvaluator = new(ReplaceEvaluator);
                res = res.RegexReplace(Regexs.DirtyWords, myEvaluator);
                res = res.RegexReplace(Regexs.AdWords, myEvaluator);
                res = res.RegexReplace(Regexs.BlackWords, myEvaluator);
                res = res.RegexReplace(Regexs.ReplaceWords, myEvaluator);
            }
            return res;
        }

        public static string ReplaceEvaluator(Match match)
        {
            string res = "";
            for (int i = 0; i < match.Length; i++)
                res += "□";
            return res;
        }

        public static string RemoveUserIds(this string text)
        {
            return text.RegexReplace(Regexs.Users, "");
        }

        #region 机器人相关操作
        // 替换非法字符为□
        public static string ReplaceInvalid(this string text)
        {
            return ReplaceInvalidKeyword(text);
        }

        public static string RemoveEmojis(this string input)
        {
            // Regular expression to match emojis and symbols
            string emojiPattern = @"[\uD800-\uDBFF][\uDC00-\uDFFF]";

            // Replace emojis with an empty string
            return Regex.Replace(input, emojiPattern, string.Empty);
        }

        // 去除文本中指定的QQ号码
        public static string RemoveUserId(this string text, long qq)
        {
            return text.Replace($"[@:{qq}]", "").Replace($"@{qq}", "").Replace($"{qq}", "").Trim();
        }

        /// <summary>
        /// 去除文本中的所有 At 消息，如 [@:12345678] 或 @昵称
        /// </summary>
        public static string RemoveAt(this string text)
        {
            if (string.IsNullOrWhiteSpace(text)) return text;

            // 移除 [@:数字] 格式的 At
            text = text.RegexReplace(@"\[@:\d+\]", "");

            return text.Trim();
        }

        //去掉空格与换行符
        public static string RemoveWhiteSpaces(this string text)
        {
            return text.Replace(" ", "").Replace("\n", "").Replace("\r", "").Replace("'", "").Replace("　", "");
        }

        // 去掉关键字中的广告信息 去掉手机QQ消息后缀 表情[FaceXXX.gif] 图片 [ImageXXXX....jpg] 等无用信息
        public static string RemoveQqAds(this string text)
        {
            text = text.RemoveQqTail();
            text = text.RemoveQqFace();
            text = text.RemoveQqImage();
            return text;
        }

        // 去除QQ尾巴
        public static string RemoveQqTail(this string text)
        {
            return text.RegexReplace(@"[\(|【|（][\S\s]*?(qq|QQ)[\S\s]*?[\)|】|）]", "");
        }

        // 去除QQ表情 微信表情 公众号表情 emoji
        public static string RemoveQqFace(this string text)
        {
            text = text.RegexReplace(@"\[Face\d.*\]", "");
            text = text.RegexReplace(Regexs.NewFace, "");
            text = text.RegexReplace(Regexs.PublicFace, "");
            text = text.RemoveEmojis();
            return text;
        }

        // 去除消息中的图片
        public static string RemoveQqImage(this string text)
        {
            return text.RegexReplace(@"\[Image[\s\S]*?\]", "");
        }

        /// 替换文本中的'为''，并在首位添加'用于生成SQL语句
        public static string Quotes(this string? text)
        {
            if (text == null)
                return "null";
            else
                return $"'{(text.IsNull() ? "" : text.Trim().Replace("'", "''"))}'";
        }

        /// 替换单引号用于sql
        public static string DoubleQuotes(this string? text)
        {
            return text.IsNull() ? "" : text.Trim().Replace("'", "''");
        }

        /// 添加 '%{text}%' 并替换 text 中的'用于SQL
        public static string QuotesLike(this string? text)
        {
            if (text == null)
                return "null";
            else
                return $"%{text}%".Quotes();
        }

        #endregion 机器人相关操作

    }
}
