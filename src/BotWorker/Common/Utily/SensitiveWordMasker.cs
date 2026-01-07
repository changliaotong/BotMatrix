using System.Text.RegularExpressions;

namespace BotWorker.Common.Utily
{
    public static class SensitiveWordMasker
    {


        public enum MaskStrategy
        {
            Full,       // 全部遮挡（全部替换成遮挡符）
            KeepFirst,  // 保留首字，其他遮挡
            KeepLast,   // 保留尾字，其他遮挡
            KeepFirstLast, // 保留首尾字，中间遮挡
            KeepNone    // 不遮挡，原文返回（预留）
        }

        /// <summary>
        /// 进行敏感词遮挡
        /// </summary>
        /// <param name="input">待处理文本</param>
        /// <param name="pattern">敏感词正则（建议用敏感词列表拼接转义）</param>
        /// <param name="strategy">遮挡策略</param>
        /// <param name="maskChar">遮挡字符，默认‘口’</param>
        /// <returns>遮挡后文本</returns>
        public static string Mask(string input, string pattern, MaskStrategy strategy = MaskStrategy.Full, char maskChar = '口')
        {
            if (string.IsNullOrEmpty(input) || string.IsNullOrEmpty(pattern))
                return input;

            return Regex.Replace(input, pattern, match =>
            {
                var word = match.Value;
                int len = word.Length;

                switch (strategy)
                {
                    case MaskStrategy.Full:
                        return new string(maskChar, len);

                    case MaskStrategy.KeepFirst:
                        if (len == 1) return new string(maskChar, 1);
                        return word[0] + new string(maskChar, len - 1);

                    case MaskStrategy.KeepLast:
                        if (len == 1) return new string(maskChar, 1);
                        return new string(maskChar, len - 1) + word[len - 1];

                    case MaskStrategy.KeepFirstLast:
                        if (len <= 2) return new string(maskChar, len);
                        return word[0] + new string(maskChar, len - 2) + word[len - 1];

                    case MaskStrategy.KeepNone:
                        return word;

                    default:
                        return new string(maskChar, len);
                }
            });
        }
    }
}