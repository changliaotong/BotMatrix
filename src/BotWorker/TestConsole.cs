using System;
using System.Threading.Tasks;
using BotWorker.Bots.BotMessages;
using BotWorker.Bots.Entries;
using BotWorker.Bots.Users;
using BotWorker.Core.Configurations;
using BotWorker.Common;
using Microsoft.Extensions.Configuration;
using BotWorker.Bots.Platform;
using Internal;

namespace BotWorker
{
    public static class TestConsole
    {
        public static async Task RunAsync(IConfiguration configuration)
        {
            // 初始化全局配置，确保数据库连接字符串等被加载
            GlobalConfig.Initialize(configuration);

            Console.WriteLine("========================================");
            Console.WriteLine("   BotWorker 机器人回复功能测试控制台   ");
            Console.WriteLine("========================================");
            Console.WriteLine("输入 'exit' 或 'quit' 退出测试。");
            Console.WriteLine();

            // 初始化模拟数据
            var selfInfo = new BotInfo { BotUin = 51437810, BotName = "测试机器人", BotType = Platforms.BotType(Platforms.NapCat) };
            var groupInfo = new GroupInfo { Id = 86433316, GroupName = "测试群", IsPowerOn = true };
            var userInfo = new UserInfo { Id = 1653346663, Name = "测试用户", Credit = 1000 };

            while (true)
            {
                Console.ForegroundColor = ConsoleColor.Green;
                Console.WriteLine("用户发送 (输入 'exit' 退出): ");
                Console.ResetColor();
                
                var input = Console.ReadLine();
                if (string.IsNullOrEmpty(input)) continue;
                if (input.ToLower() is "exit" or "quit") break;
                Console.WriteLine($"{input}");
                await ProcessInput(input, selfInfo, groupInfo, userInfo);
            }
        }

        private static async Task ProcessInput(string input, BotInfo selfInfo, GroupInfo groupInfo, UserInfo userInfo)
        {
            try
            {
                // 创建模拟的 BotMessage
                var botMsg = new BotMessage
                {
                    SelfInfo = selfInfo,
                    Group = groupInfo,
                    User = userInfo,
                    Message = input,
                    CurrentMessage = input,
                    EventType = "GroupMessageEvent", 
                    IsAtMe = input.Contains("@测试机器人") || input.Contains("@me"),
                };

                // 处理回复逻辑
                await botMsg.HandleEventAsync();

                Console.ForegroundColor = ConsoleColor.Cyan;
                Console.Write("机器人回复: ");
                Console.ResetColor();

                if (string.IsNullOrEmpty(botMsg.Answer))
                {
                    Console.WriteLine("(无回复内容 / 可能是命中静默逻辑或数据库查询失败)");
                }
                else
                {
                    Console.WriteLine(botMsg.Answer);
                }

                if (!string.IsNullOrEmpty(botMsg.Reason))
                {
                    Console.ForegroundColor = ConsoleColor.DarkGray;
                    Console.WriteLine($"[原因/标签: {botMsg.Reason}]");
                    Console.ResetColor();
                }
            }
            catch (Exception ex)
            {
                Console.ForegroundColor = ConsoleColor.Red;
                Console.WriteLine($"处理出错: {ex.Message}");
                if (ex.InnerException != null)
                {
                    Console.WriteLine($"内部错误: {ex.InnerException.Message}");
                }
                Console.ResetColor();
            }
            Console.WriteLine();
        }
    }
}
