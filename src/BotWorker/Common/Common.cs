using System;
using System.ComponentModel;
using System.Diagnostics;
using System.Net;
using System.Runtime.InteropServices;
using System.Text;
using System.Text.RegularExpressions;

namespace BotWorker.Common
{
    public class Common
    {
        public const string RsaPrivateKey = $"<RSAKeyValue><Modulus>v3YlM8/BZ/nC+Ix3W0CtWUoOhkkN2XTbrXTwYlppxbTHVtWkZ9Mm+E4ZIbaaxja18LxmcOjvo0rHRZbD++/XK98fwTtfJPIKMKSaJR8WsrDyntQUB2rdfCRNmx3O17ds6PGVnjefHWUc4Yichdl/E//ITyJ6XXUqPLO8IWCT86E=</Modulus><Exponent>AQAB</Exponent><P>zGGIv8GMnX8zYL/Xw3dnewWi+rsCDcK66bgaF8pC/CLiNwhKlCS6FAnz6C/Y2cye9uMl+KVTRz5a5A4N+cdGRw==</P><Q>79FJyNQT0z+bhUQDtSx1Z67Fu9pB5yGEubBETgbI88X5KnsyyXw5pnmt3b7QI3sazXblFV/To7ok3qNOHsmi1w==</Q><DP>hrZfCW2Mvp8CAWpR0D/a0Ea11yf+QY2x361+XWHu5vwjOPzZE25lzCGHR+qJt31c5gRwmcR28MWT6S+uXI3Rrw==</DP><DQ>QnAMrOJ0C5YXk7ff/xUuAWddyEkS8OFMT9URVzxx93blLGutCjysC/6xuDjgmLPGHR3PITjG/RjYlgVP4x+hSQ==</DQ><InverseQ>BBs/6OvSwSzroCpW+7HLwNBiSbwoPKM8Jb56oXUMB2F2tZUjvXq9Z5FvfKTMXUczyH0SYA4pDJeBNTH9bNjwkQ==</InverseQ><D>tV+ytnZlfZ45eUN3/lYy4ZcqU0P5frsZMCTLZCDKeqRbAoO5DzIUhL1XSXy2+nbxvHB9ixDfkw1P4TiFyLDYX+5eDXIJPSFOy/XEYVjxCGIsxQ8D+x5TUWG5x9906grDqiWFRuG7yLT4IFLIaaWkOjJ6rrvCnjwkCh0+Ws1xNBE=</D></RSAKeyValue>";
        public const string url = "https://sz84.com";
        public const string apiKey = $"AFCDE195E9EE00DCFCB5E0ED44D129EB";
        public const string C2CMessageGroupOpenid = "02DF48FCE01B41D532ECB28B63898DE7";
        public const long C2CMessageGroupId = 990000000003;
        public const long SystemPromptGroup = 320;
        public const string RetryMsg = "操作失败，请稍后重试";
        public const string RetryMsgTooFast = $"速度太快了，请稍后再试";
        public const string OwnerOnlyMsg = $"此命令仅机器人主人可用";
        public const string YearOnlyMsg = "非年费版不能使用此功能";
        public const string BlackListMsg = "该号码已被官方拉黑";
        public const string CreditSystemClosed = "积分系统已关闭";
        public const string NoAnswer = $"这个问题我不会，输入【教学】了解如何教我说话";
        public const string AnswerExists = $"这个我已经学过了，再教我点别的吧~";        

        public static long[] OfficalBots { get; set; } = [3889418604, 3889420782, 3889411042, 3889610970, 3889535978, 3889494926, 3889699720, 3889699721, 3889699722, 3889699723];

        private static readonly Random _random = new();

        public static bool RandomBool()
        {
            return _random.Next(2) == 0;
        }

        public static void RestartApplication()
        {
            string? currentExePath = Process.GetCurrentProcess().MainModule?.FileName;

            if (!string.IsNullOrEmpty(currentExePath))
            {
                // 启动一个新进程
                Process.Start(new ProcessStartInfo
                {
                    FileName = currentExePath,
                    UseShellExecute = true // 必须设置为 true 才能启动新的进程
                });
            }

            // 结束当前进程
            Environment.Exit(0);
        }

        // windows 登录用户名
        [DllImport("secur32.dll", EntryPoint = "GetUserNameEx", CharSet = CharSet.Unicode, SetLastError = true)]
        [return: MarshalAs(UnmanagedType.Bool)]
        private static extern bool GetUserNameEx(int nameDisplay, StringBuilder lpNameBuffer, ref int nSize);

        // windows 登录用户名
        [DllImport("advapi32.dll", CharSet = CharSet.Unicode, SetLastError = true)]
        [return: MarshalAs(UnmanagedType.Bool)]
        private static extern bool GetUserName(StringBuilder lpBuffer, ref int nSize);

        // 用户名
        public static string GetUserName()
        {
            StringBuilder username = new(256);
            int usernameSize = username.Capacity;
            if (!GetUserName(username, ref usernameSize))
            {
                int errorCode = Marshal.GetLastWin32Error();
                throw new Win32Exception(errorCode);
            }
            return username.ToString();
        }

        public static string GetHostName()
        {
            return Dns.GetHostName();
        }

        public static long RandomInt64(long max)
        {
            return new Random().NextInt64(max + 1);
        }

        public static long RandomInt64(long min, long max)
        {
            return new Random().NextInt64(min, max + 1);
        }

        public static int RandomInt(int max)
        {
            return new Random().Next(max + 1);
        }

        public static int RandomInt(int min, int max)
        {
            return new Random().Next(min, max + 1);
        }

        public static void ShowMessage(string text, ConsoleColor color = ConsoleColor.Cyan, bool displayTime = true)
        {
            if (displayTime)
            {
                Console.Write(DateTime.Now.ToLongTimeString() + " ", ConsoleColor.White);
            }

            var colorMap = new Dictionary<string, ConsoleColor>
                    {
                        {"red", ConsoleColor.Red},
                        {"green", ConsoleColor.Green},
                        {"blue", ConsoleColor.Blue},
                        {"yellow", ConsoleColor.Yellow},
                        {"cyan", ConsoleColor.Cyan},
                        {"white", ConsoleColor.Gray}
                    };

            var pattern = @"\<(.*?)\>(.*?)\<(.*?)\>";
            var matches = Regex.Matches(text, pattern);

            var textParts = Regex.Split(text, pattern);

            for (int i = 0; i < textParts.Length; i++)
            {
                if (i < matches.Count)
                {
                    var colorText = matches[i].Groups[1].Value.ToLower();
                    var textPart = matches[i].Groups[2].Value;

                    ConsoleColor consoleColor;
                    if (colorMap.TryGetValue(colorText, out consoleColor))
                    {
                        Console.ForegroundColor = consoleColor;
                    }

                    Console.Write(textPart);
                }
                else
                {
                    Console.ForegroundColor = color;
                    Console.Write(textParts[i]);
                }
            }

            Console.ResetColor();
            Console.WriteLine();
        }

        // 不带时间
        public static void InfoMessage(string text, ConsoleColor color = ConsoleColor.Cyan)
        {
            ShowMessage(text, color, false);
        }

        // 错误信息 红字显示
        public static void ErrorMessage(string text, bool displayTime = true, ConsoleColor color = ConsoleColor.Red)
        {
            ShowMessage(text, color, displayTime);
        }
    }
}
