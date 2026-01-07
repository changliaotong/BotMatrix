using BotWorker.Core.MetaDatas;

namespace BotWorker.Core.Data
{
    public class Cidian : MetaData<Cidian>
    {
        public override string DataBase => "baseinfo";
        public override string TableName => "ciba";
        public override string KeyField => "keyword";

        // 翻译功能 先从数据库读取单词翻译，不存在的再调用有道翻译
        public static string GetCiDianRes(string text)
        {
            string res = GetValue("DESCRIPTION", text);

            res = res.ReplaceInvalid();

            return res;
        }
        public static string GetCiba(string? text)
        {
            if (text == null) return "";
            string sWhere = $"keyword like " + $"{text}%".Quotes();
            string res = QueryWhere("top 20 KEYWORD,DESCRIPTION", sWhere, "KEYWORD", "{0} {1}<br />");
            if (res == "")
            {
                if (text.Trim().Contains(' '))
                    res = GetCiba(text.Trim().Split(['\u002C', ' ', '，', '、', '\n'], StringSplitOptions.RemoveEmptyEntries).Last());
            }
            return res;
        }
    }
}
