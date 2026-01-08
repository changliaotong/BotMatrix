using BotWorker.Bots.BotMessages;
using BotWorker.Bots.Entries;
using BotWorker.Core.MetaDatas;

namespace BotWorker.Core.Data
{
    public class Chengyu : MetaData<Chengyu>
    {
        public override string DataBase => "baseinfo";
        public override string TableName => "chengyu";
        public override string KeyField => "oid";

        public static long GetOid(string text)
        {
            return Query($"select top 1 {Key} from {FullName} where replace(chengyu, '��', '') = '{text.RemoveBiaodian()}'").AsLong();
        }

        public static bool Exists(string text)
        {
            return GetOid(text) != 0;
        }

        public static string PinYin(string text)
        {
            return GetValue("pingyin", GetOid(text));
        }

        /// ƴ��ASCII
        public static string PinYinAscii(string text)
        {
            return GetValue("pinyin", GetOid(text));
        }

        /// �������
        public static string GetCyInfo(string text, long oid = 0)
        {
            if (oid == 0)
                oid = GetOid(text);
            string sSelect = $"chengyu, pingyin, isnull(N'\n??�����塿' + diangu,''), isnull(N'\n??��������' + chuchu,''), isnull(N'\n??�����ӡ�' + lizi,'')";
            string sWhere = $"oid = {oid}";
            string sOrderby = "";
            string format = "??�����{0}\n??��ƴ���{1}{2}{3}{4})";
            return QueryWhere(sSelect, sWhere, sOrderby, format);
        }

        //һ�λ�ö������Ľ�����ҳ��
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

        //���������ҳ�� ƴ����ָ���ϸ
        public static string GetInfoHtml(string text, long oid = 0)
        {
            if (oid == 0)
                oid = GetOid(text);
            string sSelect = $"chengyu, pingyin +' <span>|</span> ' + pinyin + ' <span>|</span> ' + spinyin, isnull('\n�����塿' + diangu,''), isnull('\n��������' + chuchu,''), isnull('\n�����ӡ�' + lizi,'')";
            string sWhere = $"oid = {oid}";
            string sOrderby = "";
            string format = "??�����{0}\n??��ƴ���{1}{2}{3}{4})";
            return QueryWhere(sSelect, sWhere, sOrderby, format);
        }

        //һ�λ�ö������Ľ�����ҳ��
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

        //����ƴ��
        public static string PinYinFirst(string textCy)
        {
            return PinYinAscii(textCy)[..PinYinAscii(textCy).IndexOf(' ')];
        }

        //β��ƴ��
        public static string PinYinLast(string text)
        {
            return PinYinAscii(text).Substring(PinYinAscii(text).LastIndexOf(' ') + 1, PinYinAscii(text).Length - PinYinAscii(text).LastIndexOf(" ") - 1);
        }


        //�������
        public static async Task<string> GetCyResAsync(BotMessage bm)
        {
            if (bm.CmdPara.Contains("����"))
            {
                if (BotCmd.IsClosedCmd(bm.GroupId, "����"))
                    return "���������ѹر�";
                else
                {
                    bm.Answer = bm.Answer.Replace("����", "");
                    return await bm.GetJielongRes();
                }
            }

            if (bm.CmdPara.IsNull())
                return "?? ��ʽ������ + �ؼ���\n?? ���磺���� �¸�����";
            string sWhere = $"chengyu like {bm.CmdPara.QuotesLike()} or replace(pinyin, ' ', '') like {bm.CmdPara.Replace(" ", "").QuotesLike()} or spinyin like {bm.CmdPara.QuotesLike()}";
            var i = CountWhere(sWhere);
            if (i == 0)
                return "û���ҵ���س���";
            string res = i == 1
                ? GetCyInfo("", GetWhere("oid", sWhere).AsLong())
                : "??" + QueryWhere("top 50 chengyu", sWhere, "newid()", "��{0}��", "��{c}��");
            return res + bm.MinusCreditRes(10, "����۷�");
        }

        // ���� �������巴�����
        public static string GetFanChaRes(BotMessage bm)
        {
            if (bm.CmdPara.IsNullOrWhiteSpace())
                return "?? ��ʽ������ + �ؼ���\n���磺���� ��ǿ ";
            string sWhere = $"diangu like {bm.CmdPara.QuotesLike()}";
            var i = CountWhere(sWhere);
            if (i == 0)
                return "û���ҵ���س���";
            string res = i == 1
                ? GetCyInfo("", GetWhere("oid", sWhere).AsLong())
                : QueryWhere("top 50 chengyu", sWhere, "newid()", "��{0}��", "��{c}��");
            res += bm.MinusCredit(10, "����۷�");
            return res;
        }

    }
}


