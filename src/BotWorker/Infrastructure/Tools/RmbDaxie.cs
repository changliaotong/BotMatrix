namespace BotWorker.Infrastructure.Tools
{
    public class RmbDaxie
    {

        public static string GetRmbDaxie(decimal number)
        {            
            var s = number.ToString("#L#E#D#C#K#E#D#C#J#E#D#C#I#E#D#C#H#E#D#C#G#E#D#C#F#E#D#C#.0B0A");
            var d = s.RegexReplace(@"((?<=-|^)[^1-9]*)|((?'z'0)[0A-E]*((?=[1-9])|(?'-z'(?=[F-L\.]|$))))|((?'b'[F-L])(?'z'0)[0A-L]*((?=[1-9])|(?'-z'(?=[\.]|$))))", "${b}${z}");
            var r = d.RegexReplace(".", m => "负元空零壹贰叁肆伍陆柒捌玖空空空空空空空分角拾佰仟万亿兆京垓秭穰"[m.Value[0] - '-'].ToString());
            return r;
        }

        // 人民币大写
        public static string GetDaxieRes(string text)
        {
            string res;
            if (text.IsDecimal())
            {
                res = GetRmbDaxie(text.AsDecimal());
                res = text.Contains('.') ? res : $"{res}整";
            }
            else
            {
                return text == ""
                    ? "格式：大写 + 金额\n例如：大写 52013.14\n伍萬贰仟零壹拾叁元壹角肆分"
                    : text.ToUpper();
            }
            return res;
        }

        public static string GetXiaoxieRes(string text)
        {
            return text.ToLower();
        }
    }
}
