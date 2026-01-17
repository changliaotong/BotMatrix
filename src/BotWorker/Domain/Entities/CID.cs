using System;
using System.Globalization;
using System.Linq;
using System.Threading.Tasks;
using BotWorker.Common;
using BotWorker.Domain.Models.BotMessages;
using BotWorker.Domain.Repositories;
using Microsoft.Extensions.DependencyInjection;

namespace BotWorker.Domain.Entities
{
    public class CID
    {
        public static string GetCidRes(BotMessage msg)
        {
            string cid = msg.Message;
            if (!CheckIDCard(cid)) return "身份证号码格式不正确";
            return $"身份证号码：{cid}\n查询结果：校验通过，格式合法。";
        }

        //验证身份证号码
        public static bool CheckIDCard(string Id)
        {
            return Id.Length switch
            {
                18 => CheckIDCard18(Id),
                15 => CheckIDCard15(Id),
                _ => false
            };
        }

        // 验证18位身份证号
        private static bool CheckIDCard18(string id, bool isCheckValid = true)
        {
            if (long.TryParse(id.Remove(17), out long n) == false || n < Math.Pow(10, 16) || long.TryParse(id.Replace('x', '0').Replace('X', '0'), out n) == false)
                return false; // 数字验证

            string[] provinces = { "11", "12", "13", "14", "15", "21", "22", "23", "31", "32", "33", "34", "35", "36", "37", "41", "42", "43", "44",
                "45", "46", "50", "51", "52", "53", "54", "61", "62", "63", "64", "65", "71", "81", "82", "91" };
            if (!provinces.Contains(id.Substring(0, 2)))
                return false; // 省份验证

            if (!DateTime.TryParseExact(id.Substring(6, 8), "yyyyMMdd", CultureInfo.InvariantCulture, DateTimeStyles.None, out _))
                return false; // 生日验证

            if (isCheckValid)
            {
                int[] factors = { 7, 9, 10, 5, 8, 4, 2, 1, 6, 3, 7, 9, 10, 5, 8, 4, 2 };
                int sum = factors.Select((factor, index) => factor * int.Parse(id[index].ToString())).Sum();
                int mod = sum % 11;
                string[] checkCode = { "1", "0", "x", "9", "8", "7", "6", "5", "4", "3", "2" };
                if (!string.Equals(checkCode[mod], id.Substring(17, 1), StringComparison.OrdinalIgnoreCase))
                    return false; // 校验码验证
            }

            return true; // 符合GB11643-1999标准
        }

        /// 验证15位身份证号
        private static bool CheckIDCard15(string id)
        {
            if (long.TryParse(id, out long n) == false || n < Math.Pow(10, 14))
                return false; // 数字验证

            string[] provinces = { "11", "12", "13", "14", "15", "21", "22", "23", "31", "32", "33", "34", "35", "36", "37", "41", "42", "43", "44", "45",
                "46", "50", "51", "52", "53", "54", "61", "62", "63", "64", "65", "71", "81", "82", "91" };
            if (!provinces.Contains(id.Substring(0, 2)))
                return false; // 省份验证

            if (!DateTime.TryParseExact(id.Substring(6, 6), "yyMMdd", CultureInfo.InvariantCulture, DateTimeStyles.None, out _))
                return false; // 生日验证

            return true; // 符合15位身份证标准
        }

        public static string GetCidRes(string text, IIDCRepository repository)
        {
            var id = text;
            if (id.Length != 18)
                return $"命令格式：身份证 + 18位号码\n例如：\n身份证 {GenerateRandomID(repository, id)}";
            string ymd = id[6..14];

            string result;
            if (ymd == "********")
            {
                if (!CheckIDCard18(id.Replace("********", "20111111"), false))
                    return "身份证号不正确";
                result = GuessId(id);
            }
            else
            {
                if (!CheckIDCard(id))
                    return "身份证号不正确";

                result = $"身份证号：{id}\n" +
                         $"地区：{GetAreaName(repository, id[..6])}\n" +
                         $"生日：{id[6..10]}年{id[10..12]}月{id[12..14]}日\n" +
                         $"性别：{(int.Parse(id[14..17]) % 2 == 0 ? "女" : "男")} 年龄：{DateTime.Now.Year - int.Parse(id[6..10])}";
            }
            return result;
        }

        public static string GetCidRes(BotMessage bm, bool isMinus = true)
        {
            var res = GetCidRes(bm.Message, bm.IDCRepository);
            if (isMinus)            
                res += bm.MinusCreditRes(10, "查身份证扣分");
            return res;
        }

        public static string GuessId(string id)
        {
            string res = string.Empty;
            for (int year = DateTime.Now.Year; year >= 1900; year--)
            {
                for (int month = 12; month >= 1; month--)
                {
                    int daysInMonth = DateTime.DaysInMonth(year, month);

                    for (int day = daysInMonth; day >= 1; day--)
                    {
                        string newid = id.Replace("********", $"{year}{month:00}{day:00}");
                        if (CheckIDCard18(newid))
                        {
                            res += $"{newid}\n";
                        }
                    }
                }
            }
            return res;
        }

        public static string GenerateRandomID(IIDCRepository repository, string dq = "")
        {
            string areaCode = repository.GetRandomBmAsync(dq).GetAwaiter().GetResult() ?? "110101";

            Random rnd = new Random();
            int year = rnd.Next(1920, DateTime.Now.Year);
            int month = rnd.Next(1, 13);
            int day = rnd.Next(1, DateTime.DaysInMonth(year, month) + 1);
            int order = rnd.Next(1, 1000);

            string id = $"{areaCode}{year}{month:D2}{day:D2}{order:D3}";

            int[] factors = { 7, 9, 10, 5, 8, 4, 2, 1, 6, 3, 7, 9, 10, 5, 8, 4, 2 };
            int sum = 0;
            for (int i = 0; i < 17; i++)
            {
                sum += factors[i] * int.Parse(id[i].ToString());
            }
            int mod = sum % 11;
            string[] checkCodes = { "1", "0", "X", "9", "8", "7", "6", "5", "4", "3", "2" };

            return $"{id}{checkCodes[mod]}";
        }

        // 身份证归属地
        public static string GetAreaName(IIDCRepository repository, string areaCode)
        {
            return repository.GetAreaNameAsync(areaCode).GetAwaiter().GetResult() ?? "未知";
        }
    }
}
