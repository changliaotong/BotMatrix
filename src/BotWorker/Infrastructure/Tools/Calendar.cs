using System.Drawing.Imaging;
using System.Drawing;
using BotWorker.Common;

namespace BotWorker.Infrastructure.Tools
{
    public class Calendar : MetaData<Calendar>
    {
        public override string TableName => throw new NotImplementedException();

        public override string KeyField => throw new NotImplementedException();

        // 得到月历
        public static string GetMonthRes(DateTime dt, bool isYinli = false, int spaceCount = 3, int spaceCount2 = 1)
        {
            //1号
            DateTime FirstDay = dt.AddDays(-dt.Day + 1);
            DateTime LastDay = FirstDay.AddMonths(1).AddDays(-1);
            int dayOfWeek = (int)FirstDay.DayOfWeek;
            dayOfWeek = dayOfWeek == 0 ? 7 : dayOfWeek;
            //年月
            string res = $"\n\n{" ".Times((int)(isYinli ? 8 + Ext.Max(spaceCount*2, spaceCount2*3) : 4 +spaceCount2*3))}{dt.Year}年{dt.Month}月\n\n{(isYinli ? " " : "  ")}";
            //星期
            foreach (var dow in Yinli.dayOfWeeks2)
                res += isYinli ? $" {dow}{" ".Times(spaceCount2+1)}" : $"{dow}{" ".Times(spaceCount-2)}";
            //阳历
            string res1 = "\n" + " ".Times((dayOfWeek - 1) * (isYinli ? spaceCount + 2 : spaceCount) + 2);
            //阴历
            string res2 = " ".Times((dayOfWeek - 1) * (spaceCount2 + 4));
            int j = 0;
            for (int i = 0; i < LastDay.Day; i++)
            {
                DateTime today = FirstDay.AddDays(i);
                res1 += $"{(today.Day < 10 ? $"0{today.Day}" : $"{today.Day}")}{" ".Times(isYinli ? spaceCount : spaceCount-2)}";
                if (isYinli)
                {
                    if (isYinli && (dt > Yinli.dateMax || dt < Yinli.dateMin))
                        return $"农历仅支持{Yinli.dateMin}至{Yinli.dateMax}";
                    try
                    {
                        Yinli yldt = new(today);
                        res2 += (yldt.Day == 1 ? $"{yldt.MonthName}{(yldt.MonthName?.Length > 1 ? "" : "月")}" : yldt.DayName) + " ".Times(spaceCount2);
                    }
                    catch (Exception ex)
                    {
                        SQLConn.DbDebug(ex.Message, "Calendar 日历");
                        return $"农历仅支持{Yinli.dateMin}至{Yinli.dateMax}";
                    }                   
                    
                }
                if (today.DayOfWeek == DayOfWeek.Sunday || today.Month == LastDay.Month && today.Day == LastDay.Day)
                {
                    res += $"  {res1}\n";
                    if (isYinli)
                        res += $" {res2}\n";
                    res1 = "";
                    res2 = "";
                    j++;
                }
            }            
            return res + "\n".Times(6-j);
        }

        // 多个月份
        public static string GetMultMonth(DateTime dt, int month = 2, int spaceCount = 3, int spaceCount2 = 1, bool yinli = true)
        {
            string res = "";
            for (int i=0; i < month; i++)
            {
                string thisMonth = GetMonthRes(dt.AddMonths(i), yinli, spaceCount, spaceCount2);
                if (i % 3 == 0)
                    res += thisMonth;
                else
                    res = MergeCalendar(res, thisMonth);
            }
            return res;    
        }

        // 合并月历
        public static string MergeCalendar(string textMonth, string textNextMonth)
        {
            string res = "";
            string[] thisMonths = textMonth.Split("\n");
            string[] nextMonths = textNextMonth.Split("\n");
            for (int i = 0; i < Ext.Max(thisMonths.Length, nextMonths.Length); i++)
                res += $"{thisMonths[i]}     {nextMonths[i]}\n";
            return res;
        }

        // 文字转化为图片形式输出 base64
        public static string ImageCalendar(DateTime today, int month, Font font, Color bgColor, Color bgFontColor, Color fontColor, bool isYinli, int spaceCount, int spaceCount2, int months = 3, ImageGen.PlatForm platform = ImageGen.PlatForm.dotnet)
        {
            if (OperatingSystem.IsWindows())
            {
#pragma warning disable CA1416 // 验证平台兼容性
                Graphics graphics = Graphics.FromImage(new Bitmap(1, 1));
                SizeF sizeF = new(0, 0);
                PointF pointF = new(0, 0);
                float maxHeight = 0;
                float maxWidth = 0;
                List<string> monthList = new();
                for (int i = 0; i < month; i++)
                {
                    string textMonth = GetMonthRes(today.AddMonths(i), isYinli, spaceCount, spaceCount2);
                    monthList.Add(textMonth);
                    SizeF sizeF1 = graphics.MeasureString(textMonth, font);//测量出字体的高度和宽度
                    if (sizeF1.Height > maxHeight)
                        maxHeight = sizeF1.Height;
                    if (sizeF1.Width > maxWidth)
                        maxWidth = sizeF1.Width;
                }
                maxHeight += 50;
                maxWidth += 50;
                int x = Convert.ToInt32(maxWidth) * (month > months ? months : month) + 50;
                int y = Convert.ToInt32(maxHeight) * (month % months == 0 ? month / months : month / months + 1) + 50;
                Bitmap img = new(x, y);
                graphics = Graphics.FromImage(img);
                Brush brush = new SolidBrush(bgColor);
                graphics.FillRectangle(brush, new RectangleF(pointF, new SizeF(x, y)));
                for (int i = 0; i < month; i++)
                {
                    pointF.X = maxWidth * (i % months) + 50;
                    pointF.Y = maxHeight * (i / months) + 50;
                    sizeF.Height = maxHeight - 40;
                    sizeF.Width = maxWidth - 40;
                    brush = new SolidBrush(bgFontColor);
                    graphics.FillRectangle(brush, new RectangleF(pointF, sizeF));
                    Brush brush2 = new SolidBrush(fontColor);
                    graphics.DrawString(monthList[i], font, brush2, pointF);
                }
                //输出图片
                MemoryStream ms = new();
                img.Save(ms, ImageFormat.Gif);
                string res = Convert.ToBase64String(ms.ToArray());
                return platform switch
                {
                    ImageGen.PlatForm.java => res,
                    ImageGen.PlatForm.dotnet => $"data:image/gif;base64,{res}",
                    _ => res
                };
            }
            else
                return string.Empty;
        }
    }
}
