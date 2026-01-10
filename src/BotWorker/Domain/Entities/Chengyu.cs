using BotWorker.Domain.Models.Messages.BotMessages;
using BotWorker.Domain.Entities;

using BotWorker.Infrastructure.Extensions;
using BotWorker.Infrastructure.Persistence.ORM;

namespace BotWorker.Domain.Entities
{
    public class Chengyu : MetaData<Chengyu>
    {
        public override string DataBase => "baseinfo";
        public override string TableName => "chengyu";
        public override string KeyField => "oid";

        public static async Task<long> GetOidAsync(string text)
        {
            return await QueryScalarAsync<long>($"select {SqlTop(1)} {Key} from {FullName} where replace(chengyu, '，', '') = '{text.RemoveBiaodian()}'{SqlLimit(1)}");
        }

        public static async Task<bool> ExistsAsync(string text)
        {
            return await GetOidAsync(text) != 0;
        }

        public static async Task<string> PinYinAsync(string text)
        {
            return await GetValueAsync("pingyin", await GetOidAsync(text));
        }

        /// 拼音ASCII
        public static async Task<string> PinYinAsciiAsync(string text)
        {
            return await GetValueAsync("pinyin", await GetOidAsync(text));
        }

        /// 成语解释
        public static async Task<string> GetCyInfoAsync(string text, long oid = 0)
        {
            if (oid == 0)
                oid = await GetOidAsync(text);
            string prefix = IsPostgreSql ? "" : "N";
            string sSelect = $"chengyu, pingyin, {SqlIsNull(prefix + "'\n💡【释义】' + diangu", "''")}, {SqlIsNull(prefix + "'\n📜【出处】' + chuchu", "''")}, {SqlIsNull(prefix + "'\n📝【例子】' + lizi", "''")}";
            string sWhere = $"oid = {oid}";
            string sOrderby = "";
            string format = "📚【成语】{0}\n🔤【拼音】{1}{2}{3}{4}";
            return await QueryWhereAsync(sSelect, sWhere, sOrderby, format);
        }

        //一次获得多个成语的解释网页版
        public static async Task<Dictionary<string, string>> GetCyInfoAsync(IEnumerable<string> cys)
        {
            Dictionary<string, string> res = [];
            foreach (var cy in cys)
            {
                string cyInfo = await GetCyInfoAsync(cy);
                res.TryAdd(cy, cyInfo);
            }
            return res;
        }

        //成语解释网页版 拼音部分更详细
        public static async Task<string> GetInfoHtmlAsync(string text, long oid = 0)
        {
            if (oid == 0)
                oid = await GetOidAsync(text);
            string sSelect = $"chengyu, pingyin +' <span>|</span> ' + pinyin + ' <span>|</span> ' + spinyin, {SqlIsNull("'\n【释义】' + diangu", "''")}, {SqlIsNull("'\n【出处】' + chuchu", "''")}, {SqlIsNull("'\n【例子】' + lizi", "''")}";
            string sWhere = $"oid = {oid}";
            string sOrderby = "";
            string format = "📚【成语】{0}\n🔤【拼音】{1}{2}{3}{4}";
            return await QueryWhereAsync(sSelect, sWhere, sOrderby, format);
        }

        //一次获得多个成语的解释网页版
        public static async Task<Dictionary<string, string>> GetInfoHtmlAsync(IEnumerable<string> cys)
        {
            Dictionary<string, string> res = [];
            foreach (var cy in cys)
            {
                string cyInfo = await GetInfoHtmlAsync(cy);
                res.TryAdd(cy, cyInfo);
            }
            return res;
        }

        //首字拼音
        public static async Task<string> PinYinFirstAsync(string textCy)
        {
            var pinyin = await PinYinAsciiAsync(textCy);
            return pinyin[..pinyin.IndexOf(' ')];
        }

        //尾字拼音
        public static async Task<string> PinYinLastAsync(string text)
        {
            var pinyin = await PinYinAsciiAsync(text);
            return pinyin.Substring(pinyin.LastIndexOf(' ') + 1, pinyin.Length - pinyin.LastIndexOf(" ") - 1);
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
            var i = await CountWhereAsync(sWhere);
            if (i == 0)
                return "没有找到相关成语";
            string res = i == 1
                ? await GetCyInfoAsync("", (await GetWhereAsync("oid", sWhere)).AsLong())
                : "📚" + await QueryWhereAsync("top 50 chengyu", sWhere, "newid()", "【{0}】", "共{c}条");
            return res + await bm.MinusCreditResAsync(10, "成语扣分");
        }

        // 反查 根据释义反查成语
        public static string GetFanChaRes(BotMessage bm)
            => GetFanChaResAsync(bm).GetAwaiter().GetResult();

        public static async Task<string> GetFanChaResAsync(BotMessage bm)
        {
            if (bm.CmdPara.IsNullOrWhiteSpace())
                return "📚 格式：反查 + 关键字\n例如：反查 坚强 ";
            string sWhere = $"diangu like {bm.CmdPara.QuotesLike()}";
            var i = await CountWhereAsync(sWhere);
            if (i == 0)
                return "没有找到相关成语";
            string res = i == 1
                ? await GetCyInfoAsync("", await GetWhereAsync<long>("oid", sWhere))
                : await QueryWhereAsync("top 50 chengyu", sWhere, "newid()", "【{0}】", "共{c}条");
            res += await bm.MinusCreditResAsync(10, "成语扣分");
            return res;
        }
    }
}

