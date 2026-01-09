using BotWorker.Common;
using BotWorker.Modules.AI.Plugins;
using BotWorker.Modules.AI.Interfaces;
using BotWorker.Infrastructure.Caching;
using BotWorker.Infrastructure.Persistence.ORM;
using Microsoft.Extensions.DependencyInjection;
using Microsoft.Extensions.Configuration;
using StackExchange.Redis;

namespace BotWorker
{
    public static class TestConsole
    {
        private static IServiceProvider? _serviceProvider;

        public static async Task RunAsync(IConfiguration configuration)
        {
            // 初始化全局配置
            GlobalConfig.Initialize(configuration);

            // 初始化 AI 提供者
            await BotMessage.LLMApp.InitializeAsync();

            // 确保菜单指令存在
            await BotWorker.Domain.Entities.BotCmd.EnsureCommandExistsAsync("菜单", "菜单");

            // 设置服务提供者
            var services = new ServiceCollection();
            services.AddHttpClient();

            // 注册 Redis 和 缓存服务
            var redisHost = configuration["redis:host"] ?? "localhost";
            var redisPort = configuration["redis:port"] ?? "6379";
            var redisConn = $"{redisHost}:{redisPort},abortConnect=false,connectTimeout=5000,syncTimeout=5000";
            try 
            {
                var redis = ConnectionMultiplexer.Connect(redisConn);
                services.AddSingleton<IConnectionMultiplexer>(redis);
                services.AddSingleton<ICacheService, RedisCacheService>();
            }
            catch (Exception ex)
            {
                Console.WriteLine($"Redis 连接失败: {ex.Message}");
                // 如果 Redis 失败，不注册 ICacheService，MetaData.CacheService 将保持为 null，自动降级为数据库查询
            }

            services.AddSingleton<IKnowledgeBaseService, KnowledgeBaseService>(sp => 
            {
                var httpClient = sp.GetRequiredService<IHttpClientFactory>().CreateClient();
                httpClient.BaseAddress = new Uri(configuration["BotWorker:KbApiUrl"] ?? "http://localhost:5000");
                return new KnowledgeBaseService(httpClient);
            });
            _serviceProvider = services.BuildServiceProvider();

            // 初始化 MetaData 缓存
            MetaData.CacheService = _serviceProvider.GetService<ICacheService>();

            Console.WriteLine("========================================");
            Console.WriteLine("   BotWorker 机器人回复功能测试控制台   ");
            Console.WriteLine("========================================");
            Console.WriteLine("输入 'exit' 或 'quit' 退出测试。");
            Console.WriteLine();

            // 初始化模拟数据
            var selfInfo = new BotInfo { BotUin = 51437810, BotName = "测试机器人", BotType = Platforms.BotType(Platforms.NapCat) };
            var groupInfo = new GroupInfo { Id = 86433316, GroupName = "测试群", IsPowerOn = true, IsUseKnowledgebase = true, IsAI = true, IsCloudAnswer = 5 };
            var userInfo = new UserInfo { Id = 1653346663, Name = "测试用户", Credit = 1000, IsAI = true };

            // 自动化测试一个命令
            Console.WriteLine(">>> 自动运行测试命令: 菜单");
            await ProcessInput("菜单", selfInfo, groupInfo, userInfo);

            Console.WriteLine(">>> 自动运行 RAG 测试: 什么是知识库？");
            await ProcessInput("什么是知识库？", selfInfo, groupInfo, userInfo);

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
                    KbService = _serviceProvider?.GetService<IKnowledgeBaseService>() as KnowledgeBaseService
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

                if (botMsg.KbService != null && botMsg.NewQuestionId > 0)
                {
                    Console.ForegroundColor = ConsoleColor.Yellow;
                    Console.WriteLine("--- RAG 检索详情 ---");
                    Console.WriteLine($"匹配问题: {botMsg.NewQuestion}");
                    Console.WriteLine($"相似度: {botMsg.Similarity:P2}");
                    Console.WriteLine($"问题 ID: {botMsg.NewQuestionId}");
                    Console.ResetColor();
                }

                if (!string.IsNullOrEmpty(botMsg.Reason))
                {
                    Console.ForegroundColor = ConsoleColor.DarkGray;
                    Console.WriteLine($"[原因/标签: {botMsg.Reason}]");
                    Console.ResetColor();
                }
                Console.WriteLine();
            }
            catch (Exception ex)
            {
                Console.ForegroundColor = ConsoleColor.Red;
                Console.WriteLine($"处理出错: {ex.Message}");
                Console.WriteLine(ex.StackTrace);
                Console.ResetColor();
            }
        }
    }
}
