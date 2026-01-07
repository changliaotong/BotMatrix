using sz84.Bots.BotMessages;
using sz84.Bots.Entries;
using sz84.Core.MetaDatas;

namespace sz84.Core.Data
{
    public class Chengyu : MetaData<Chengyu>
    {
        public override string DataBase => "baseinfo";
        public override string TableName => "chengyu";
        public override string KeyField => "oid";

        public static long GetOid(string text)
        {
            return Query($"select top 1 {Key} from {FullName} where replace(chengyu, '，', '') = '{text.RemoveBiaodian()}'").AsLong();
        }

        public static bool Exists(string text)
        {
            return GetOid(text) != 0;
        }

        public static string PinYin(string text)
        {
            return GetValue("pingyin", GetOid(text));
        }

        /// 拼音ASCII
        public static string PinYinAscii(string text)
        {
            return GetValue("pinyin", GetOid(text));
        }

        /// 成语解释
        public static string GetCyInfo(string text, long oid = 0)
        {
            if (oid == 0)
                oid = GetOid(text);
            string sSelect = $"chengyu, pingyin, isnull(N'\n💡【释义】' + diangu,''), isnull(N'\n📜【出处】' + chuchu,''), isnull(N'\n📝【例子】' + lizi,'')";
            string sWhere = $"oid = {oid}";
            string sOrderby = "";
            string format = "📚【成语】{0}\n🔤【拼音】{1}{2}{3}{4})";
            return QueryWhere(sSelect, sWhere, sOrderby, format);
        }

        //一次获得多个成语的解释网页版
        public static Dictionary<string, string> GetCyInfo(IEnumerable<string> cys)
        {
            Dictionary<string, string> res = [];
            foreach (var cy in cys)
            {
                string cyInfo = GetCyInfo(cy);
                res.TryAdd(cy, cyInfo);
            }
            return res;
        }

        //成语解释网页版 拼音部分更详细
        public static string GetInfoHtml(string text, long oid = 0)
        {
            if (oid == 0)
                oid = GetOid(text);
            string sSelect = $"chengyu, pingyin +' <span>|</span> ' + pinyin + ' <span>|</span> ' + spinyin, isnull('\n【释义】' + diangu,''), isnull('\n【出处】' + chuchu,''), isnull('\n【例子】' + lizi,'')";
            string sWhere = $"oid = {oid}";
            string sOrderby = "";
            string format = "📚【成语】{0}\n🔤【拼音】{1}{2}{3}{4})";
            return QueryWhere(sSelect, sWhere, sOrderby, format);
        }

        //一次获得多个成语的解释网页版
        public static Dictionary<string, string> GetInfoHtml(IEnumerable<string> cys)
        {
            Dictionary<string, string> res = [];
            foreach (var cy in cys)
            {
                string cyInfo = GetInfoHtml(cy);
                res.TryAdd(cy, cyInfo);
            }
            return res;
        }

        //首字拼音
        public static string PinYinFirst(string textCy)
        {
            return PinYinAscii(textCy)[..PinYinAscii(textCy).IndexOf(' ')];
        }

        //尾字拼音
        public static string PinYinLast(string text)
        {
            return PinYinAscii(text).Substring(PinYinAscii(text).LastIndexOf(' ') + 1, PinYinAscii(text).Length - PinYinAscii(text).LastIndexOf(" ") - 1);
        }


        //成语解释
        public static async Task<string> GetCyResAsync(BotMessage bm)
        {
            if (bm.CmdPara.Contains("接龙"))
            {
                if (BotCmd.IsClosedCmd(bm.GroupId, "接龙"))
                    return "接龙功能已关闭";
                else
                {
                    bm.Answer = bm.Answer.Replace("接龙", "");
                    return await bm.GetJielongRes();
                }
            }

            if (bm.CmdPara.IsNull())
                return "📚 格式：成语 + 关键字\n📌 例如：成语 德高望重";
            string sWhere = $"chengyu like {bm.CmdPara.QuotesLike()} or replace(pinyin, ' ', '') like {bm.CmdPara.Replace(" ", "").QuotesLike()} or spinyin like {bm.CmdPara.QuotesLike()}";
            var i = CountWhere(sWhere);
            if (i == 0)
                return "没有找到相关成语";
            string res = i == 1
                ? GetCyInfo("", GetWhere("oid", sWhere).AsLong())
                : "📚" + QueryWhere("top 50 chengyu", sWhere, "newid()", "【{0}】", "共{c}条");
            return res + bm.MinusCreditRes(10, "成语扣分");
        }

        // 反查 根据释义反查成语
        public static string GetFanChaRes(BotMessage bm)
        {
            if (bm.CmdPara.IsNullOrWhiteSpace())
                return "📚 格式：反查 + 关键字\n例如：反查 坚强 ";
            string sWhere = $"diangu like {bm.CmdPara.QuotesLike()}";
            var i = CountWhere(sWhere);
            if (i == 0)
                return "没有找到相关成语";
            string res = i == 1
                ? GetCyInfo("", GetWhere("oid", sWhere).AsLong())
                : QueryWhere("top 50 chengyu", sWhere, "newid()", "【{0}】", "共{c}条");
            res += bm.MinusCredit(10, "成语扣分");
            return res;
        }

    }
}
