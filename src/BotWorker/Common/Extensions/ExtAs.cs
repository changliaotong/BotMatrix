using Newtonsoft.Json;

namespace BotWorker.Common.Extensions
{
    public static class ExtAs
    {

        public static string AsChineseUnits(this long value)
        {
            if (value >= 100_000_000) // 亿
                return (value / 100_000_000.0).ToString("0.##") + " 亿";
            else if (value >= 10_000) // 万
                return (value / 10_000.0).ToString("0.##") + " 万";
            else
                return value.ToString("N0"); // 千位分隔
        }

        public static string AsTime(this DateTime dateTime)
        {
            var today = DateTime.Today;
            var diff = today - dateTime.Date;

            if (diff.Days == 0)
            {
                return dateTime.ToString("HH:mm");
            }
            else if (diff.Days == 1)
            {
                return "昨天 " + dateTime.ToString("HH:mm");
            }
            else if (diff.Days == 2)
            {
                return "前天 " + dateTime.ToString("HH:mm");
            }
            else if (diff.Days < 7)
            {
                return dateTime.ToString("dddd HH:mm"); // 过去一周内的星期几
            }
            else
            {
                return dateTime.ToString("yyyy-MM-dd HH:mm"); // 更久远的日期
            }
        }

        public static string AsJianti(this string? text)
        {
            return text ?? "";
            // return ChineseConverter.Convert(text, ChineseConversionDirection.TraditionalToSimplified);
        }

        public static string AsFanti(this string? text)
        {
            return text ?? "";
            // return ChineseConverter.Convert(text, ChineseConversionDirection.SimplifiedToTraditional);
        }

        public static string AsWide(this string input)
        {
            // 半角转全角：
            char[] c = input.ToCharArray();
            for (int i = 0; i < c.Length; i++)
            {
                if (c[i] == 32)
                {
                    c[i] = (char)12288;
                    continue;
                }
                if (c[i] < 127)
                    c[i] = (char)(c[i] + 65248);
            }
            return new string(c);
        }

        public static string AsNarrow(this string input)
        {
            char[] c = input.ToCharArray();
            for (int i = 0; i < c.Length; i++)
            {
                if (c[i] == 12288)
                {
                    c[i] = (char)32;
                    continue;
                }
                if (c[i] > 65280 && c[i] < 65375)
                    c[i] = (char)(c[i] - 65248);
            }
            return new string(c);
        }

        // 10w+
        public static string As10WPlus(this long num)
        {
            if (num >= 10000 && num < 99999999)
                return $"{num / 10000}W+";

            if (num >= 100000000)
                return $"{num / 100000000}Y+";

            return num.AsString();
        }



        /// <summary>
        /// 将 JSON 字符串转换为指定类型的对象。
        /// </summary>
        /// <typeparam name="T">目标对象类型。</typeparam>
        /// <param name="json">JSON 字符串。</param>
        /// <param name="def">如果转换失败，返回的默认值。</param>
        /// <returns>转换后的对象，或者默认值。</returns>
        public static T? AsObject<T>(this string json, T? def = default)
        {
            if (string.IsNullOrWhiteSpace(json))
            {
                return def;
            }

            try
            {
                var settings = new JsonSerializerSettings
                {
                    NullValueHandling = NullValueHandling.Ignore // 忽略 null 值
                };
                //json = JsonConvert.SerializeObject(json);
                return JsonConvert.DeserializeObject<T>(json);
            }
            catch (Exception ex)
            {
                // 获取 T 的类型名
                var typeName = typeof(T).FullName ?? typeof(T).Name;

                // 将 T 的类型信息和默认值 def 记录下来
                ErrorMessage($"Ext:AsObject {ex.Message}\nType: {typeName}\nJson:{json}");

                return def;
            }
        }

        public static string AsString(this DateTime dt, string format = "yyyy-MM-dd hh:mm:ss")
        {
            return dt.ToString(format);
        }

        public static string AsString(this object? obj, string def = "")
        {
            return obj == null ? def : obj.ToString() ?? def;
        }

        public static long AsLong(this object? obj, long def = 0)
        {
            return obj.AsString().AsLong(def);
        }

        public static int AsInt(this object? obj, int def = 0)
        {
            return obj.AsString().AsInt(def);
        }


        public static string AsDateTimeFormat(this string? text, string format = "yyyy-MM-dd HH:ss:mm")
        {
            DateTime dt = text.IsNull()
                ? DateTime.Now
                : DateTime.TryParse(text?.Trim(), out DateTime res) ? res : DateTime.Now;
            return dt.ToString(format);
        }


        public static DateTime AsDateTime(this string? text)
        {
            DateTime dt = text.IsNull()
                ? DateTime.Now
                : DateTime.TryParse(text?.Trim(), out DateTime res) ? res : DateTime.Now;
            return dt;
        }


        public static long AsLong(this string? text, long def = 0)
        {
            return text.IsNull()
                ? def
                : long.TryParse(text?.Trim(), out long res) ? res : def;
        }

        public static ulong AsULong(this string? text, ulong def = 0)
        {
            return text.IsNull()
                ? def
                : ulong.TryParse(text?.Trim(), out ulong res) ? res : def;
        }


        public static int AsInt(this bool isTrue)
        {
            return isTrue ? 1 : 0;
        }

        public static string AsCurrency(this string? balance)
        {
            return string.Format("{0:N2}", balance.AsFloat());
        }

        public static int AsInt(this string? text, int def = 0)
        {
            return text.IsNull() ? def : int.TryParse(text?.Trim(), out int res) ? res : def;
        }

        public static float AsFloat(this string? text, float def = 0)
        {
            return text.IsNull() ? def : float.TryParse(text?.Trim(), out float res) ? res : def;
        }

        public static double AsDouble(this string? text, double def = 0)
        {
            return text.IsNull() ? def : double.TryParse(text?.Trim(), out double res) ? res : def;
        }

        public static decimal AsDecimal(this string? text, decimal def = 0)
        {
            return text.IsNull() ? def : decimal.TryParse(text?.Trim(), out decimal res) ? res : def;
        }

        public static bool AsBool(this string text, bool def = false)
        {
            return text.IsNull() ? def : !text.ToLower().In("", "0", "false");
        }

        public static bool AsBool(this object? obj, bool def = false)
        {
            return obj == null ? def : obj.AsString().AsBool();
        }

        public static T? GetValue<T>(this List<Cov> result, string fieldName)
        {
            var columnValue = result.FirstOrDefault(cv => cv.Name == fieldName);
            return columnValue != null ? (T)columnValue.Value! : default;
        }
    }
}
