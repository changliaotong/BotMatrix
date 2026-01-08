using System.Diagnostics;
using System.Management;
using System.Runtime.InteropServices;
using System.Runtime.Versioning;

namespace BotWorker.Common
{
    [SupportedOSPlatform("windows")]
    public static class ConsoleHelper
    {
        const int ENABLE_QUICK_EDIT = 0x0040; // Quick Edit mode
        const int ENABLE_EXTENDED_FLAGS = 0x0080;
        const int STD_INPUT_HANDLE = -10;

        public static void ClearLine(int lineNumber)
        {
            if (lineNumber < 0 || lineNumber >= Console.WindowHeight)
                return;

            Console.SetCursorPosition(0, lineNumber);
            Console.Write(new string(' ', Console.WindowWidth));
            Console.SetCursorPosition(0, lineNumber); // 回到行首
        }

        //检测窗口是否被关闭
        public static void DetectConsoleClosing(Action onClosing)
        {
            SetConsoleCtrlHandler(eventType =>
                        {
                            if (eventType == CtrlTypes.CTRL_CLOSE_EVENT)
                            {
                                onClosing?.Invoke();
                            }
                            return false;
                        }, true);
        }

        public static void DisableQuickEditMode()
        {
            IntPtr consoleHandle = GetStdHandle(STD_INPUT_HANDLE);
            if (!GetConsoleMode(consoleHandle, out int mode))
            {
                Console.WriteLine("Unable to get console mode.");
                return;
            }

            mode &= ~ENABLE_QUICK_EDIT; // Remove the Quick Edit mode flag
            mode |= ENABLE_EXTENDED_FLAGS; // Add the extended flags to ensure mode change applies

            if (!SetConsoleMode(consoleHandle, mode))
            {
                Console.WriteLine("Unable to set console mode.");
            }
        }

        public static bool IsRunningInPowerShell()
        {
            string parentProcess = GetParentProcessName();
            return parentProcess.Contains("powershell", StringComparison.OrdinalIgnoreCase) ||
                   parentProcess.Contains("pwsh", StringComparison.OrdinalIgnoreCase);
        }

        public static string ReadPassword()
        {
            string password = string.Empty;
            ConsoleKeyInfo key;

            do
            {
                key = Console.ReadKey(intercept: true);

                if (key.Key == ConsoleKey.Backspace && password.Length > 0)
                {
                    password = password[0..^1];
                    Console.Write("\b \b");
                }
                else if (!char.IsControl(key.KeyChar))
                {
                    password += key.KeyChar;
                    Console.Write("*");
                }
            } while (key.Key != ConsoleKey.Enter);

            Console.WriteLine();
            return password;
        }

        public static void SetConsoleBufferSize(int width, int height)
        {
            if (Environment.OSVersion.Platform != PlatformID.Win32NT)
            {
                throw new NotSupportedException("This method is only supported on Windows platforms.");
            }
            Console.SetBufferSize(Math.Max(width, Console.WindowWidth),
                                  Math.Max(height, Console.WindowHeight));
        }

        [StructLayout(LayoutKind.Sequential)]
        public struct ConsoleMode
        {
            public uint Fullscreen;
            // Add other console mode flags as needed
        }

        public delegate bool ConsoleCtrlDelegate(CtrlTypes ctrlType);

        public enum CtrlTypes
        {
            CTRL_C_EVENT = 0,
            CTRL_BREAK_EVENT = 1,
            CTRL_CLOSE_EVENT = 2,
            CTRL_LOGOFF_EVENT = 5,
            CTRL_SHUTDOWN_EVENT = 6
        }


        public static void SetConsoleFullscreen()
        {
            IntPtr hConsole = GetStdHandle(-11); // STD_OUTPUT_HANDLE
            var consoleMode = new ConsoleMode
            {
                Fullscreen = 0x0001
            };
            SetConsoleDisplayMode(hConsole, consoleMode);
        }

        public static void SetConsoleTitle(string title)
        {
            Console.Title = title;
        }

        public static void SetConsoleWindowPosition(int x, int y)
        {
            IntPtr consoleWindow = GetConsoleWindow();
            if (consoleWindow != IntPtr.Zero)
            {
                SetWindowPos(consoleWindow, IntPtr.Zero, x, y, 0, 0, 0x0001 | 0x0004);
            }
        }

        public static void SetConsoleWindowSize(int width, int height)
        {
            if (OperatingSystem.IsWindows())
            {
                Console.SetWindowSize(Math.Min(width, Console.LargestWindowWidth),
                                      Math.Min(height, Console.LargestWindowHeight));
            }
        }

        public static void WriteColoredText(string text, ConsoleColor foreground, ConsoleColor background = ConsoleColor.Black)
        {
            var originalForeground = Console.ForegroundColor;
            var originalBackground = Console.BackgroundColor;

            Console.ForegroundColor = foreground;
            Console.BackgroundColor = background;

            Console.WriteLine(text);

            Console.ForegroundColor = originalForeground;
            Console.BackgroundColor = originalBackground;
        }

        [DllImport("kernel32.dll", SetLastError = true)]
        public static extern bool GetConsoleMode(IntPtr hConsoleHandle, out int lpMode);

        [DllImport("kernel32.dll", SetLastError = true)]
        public static extern IntPtr GetConsoleWindow();

        public static string GetParentProcessName()
        {
            if (Environment.OSVersion.Platform != PlatformID.Win32NT)
            {
                throw new NotSupportedException("This method is only supported on Windows platforms.");
            }

            using var process = Process.GetCurrentProcess();
            var query = $"SELECT ParentProcessId FROM Win32_Process WHERE ProcessId = {process.Id}";
            using var searcher = new ManagementObjectSearcher(query);
            var results = searcher.Get().Cast<ManagementObject>().FirstOrDefault();
            return results?["ParentProcessId"]?.ToString() ?? string.Empty;
        }

        [DllImport("kernel32.dll", SetLastError = true)]
        public static extern IntPtr GetStdHandle(int nStdHandle);

        [DllImport("kernel32.dll")]
        public static extern bool SetConsoleCtrlHandler(ConsoleCtrlDelegate handler, bool add);

        [DllImport("kernel32.dll")]
        public static extern bool SetConsoleDisplayMode(IntPtr hConsoleOutput, ConsoleMode dwFlags);

        [DllImport("kernel32.dll", SetLastError = true)]
        static extern bool SetConsoleMode(IntPtr hConsoleHandle, int dwMode);

        [DllImport("user32.dll")]
        public static extern bool SetWindowPos(IntPtr hWnd, IntPtr hWndInsertAfter, int x, int y, int cx, int cy, uint uFlags);
    }
}

