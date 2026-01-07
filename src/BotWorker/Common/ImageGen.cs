using System.Drawing;
using System.Drawing.Imaging;
using System.Text;

namespace BotWorker.Common
{
    public class ImageGen
    {
        static readonly char[] splitChars = [' ', '-', '\t'];

        // 自动换行
        public static string WordWrap(string str, int width)
        {
            string[] words = Explode(str, splitChars);

            int curLineLength = 0;
            StringBuilder strBuilder = new();
            for (int i = 0; i < words.Length; i += 1)
            {
                string word = words[i];
                // If adding the new word to the current line would be too long,
                // then put it on a new line (and split it up if it's too long).
                if (curLineLength + word.Length > width)
                {
                    // Only move down to a new line if we have text on the current line.
                    // Avoids situation where wrapped whitespace causes emptylines in text.
                    if (curLineLength > 0)
                    {
                        strBuilder.Append(Environment.NewLine);
                        curLineLength = 0;
                    }

                    // If the current word is too long to fit on a line even on it's own then
                    // split the word up.
                    while (word.Length > width)
                    {
                        strBuilder.Append(word.AsSpan(0, width - 1));
                        word = word[(width - 1)..];

                        strBuilder.Append(Environment.NewLine);
                    }

                    // Remove leading whitespace from the word so the new line starts flush to the left.
                    word = word.TrimStart();
                }
                strBuilder.Append(word);
                curLineLength += word.Length;
            }

            return strBuilder.ToString();
        }

        private static string[] Explode(string str, char[] splitChars)
        {
            List<string> parts = [];
            int startIndex = 0;
            while (true)
            {
                int index = str.IndexOfAny(splitChars, startIndex);

                if (index == -1)
                {
                    parts.Add(str[startIndex..]);
                    return [.. parts];
                }

                string word = str[startIndex..index];
                char nextChar = str.Substring(index, 1)[0];
                // Dashes and the likes should stick to the word occuring before it. Whitespace doesn't have to.
                if (char.IsWhiteSpace(nextChar))
                {
                    parts.Add(word);
                    parts.Add(nextChar.ToString());
                }
                else
                {
                    parts.Add(word + nextChar);
                }

                startIndex = index + 1;
            }
        }

        // 利用反射获取颜色名称列表
        public static List<string> GetBrushes()
        {
            List<string> res = [];
            foreach (var item in typeof(Brushes).GetProperties())
            {
                res.Add($"{item.Name}");
            }
            return res;
        }

        // 开发平台 对接
        public enum PlatForm
        {
            java,
            dotnet
        }

        // 默认值
        public static string ImageUrl(string text, PlatForm platform = PlatForm.dotnet)
        {
            if (OperatingSystem.IsWindows())
            {
#pragma warning disable CA1416 // 验证平台兼容性
                using Font font = new("宋体", 9);
#pragma warning restore CA1416 // 验证平台兼容性
                return ImageUrl(text, font, Color.White, Color.Black, platform);
            }
            else return "";
        }

        // 文字转化为图片形式输出 base64
        public static string ImageUrl(string text, Font font, Color bgColor, Color fontColor, PlatForm platform = PlatForm.dotnet)
        {
            string str = text;
            if (OperatingSystem.IsWindows())
            {
#pragma warning disable CA1416 // 验证平台兼容性
                Graphics graphics = Graphics.FromImage(new Bitmap(1, 1));
                SizeF sizeF = graphics.MeasureString(str, font); //测量出字体的高度和宽度
                PointF pointF = new(0, 0);
                Bitmap img = new(Convert.ToInt32(sizeF.Width), Convert.ToInt32(sizeF.Height));
                graphics = Graphics.FromImage(img);
                Brush brush = new SolidBrush(bgColor);
                graphics.FillRectangle(brush, new RectangleF(pointF, sizeF));
                brush = new SolidBrush(fontColor);
                graphics.DrawString(str, font, brush, pointF);
                //输出图片
                MemoryStream ms = new();
                img.Save(ms, ImageFormat.Gif);
                string res = Convert.ToBase64String(ms.ToArray());
                return platform switch
                {
                    PlatForm.java => res,
                    PlatForm.dotnet => $"data:image/gif;base64,{res}",
                    _ => res
                };
#pragma warning disable CA1416 // 验证平台兼容性
            }
            else
                return string.Empty;
        }
    }
}
