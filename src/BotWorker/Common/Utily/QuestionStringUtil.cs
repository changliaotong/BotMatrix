using System.Collections.Generic;
using BotWorker.Common.Extensions;
using BotWorker.Infrastructure.Utils;

namespace BotWorker.Common.Utily
{
    public static class QuestionStringUtil
    {
        public static string GetNew(string text)
        {
            if (string.IsNullOrWhiteSpace(text))
                return text;

            text = text.RemoveWhiteSpaces();

            // 精准匹配表情
            var faceRegex = @"\[Face\d{1,3}\.gif\]";
            var matches = text.Matches(faceRegex);

            // 替换成不含任何标点的占位符
            int faceIndex = 0;
            var faceDict = new Dictionary<string, string>(); // 占位符 -> 表情
            string temp = text.RegexReplace(faceRegex, m =>
            {
                string key = $"QQFACE{faceIndex}PLACEHOLDER";
                faceDict[key] = m.Value;
                faceIndex++;
                return key;
            });

            // 去标点（你已有的）
            temp = temp.RegexReplace(Regexs.BiaoDian, "");

            // 还原表情
            foreach (var kv in faceDict)
            {
                temp = temp.Replace(kv.Key, kv.Value);
            }

            // 判断是否全为表情或标点
            bool isAllRemoved = temp.RemoveQqFace().RemoveBiaodian().IsNull();
            return isAllRemoved ? text : temp;
        }
    }
}
