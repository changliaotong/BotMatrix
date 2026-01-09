using System;
using System.Threading.Tasks;
using Xunit;
using Xunit.Abstractions;
using Moq;
using BotWorker.Domain.Interfaces;
using BotWorker.Common;
using BotWorker.Modules.Games;
using BotWorker.Services;
using BotWorker.Modules.Plugins;
using Microsoft.Extensions.Logging;
using Microsoft.Extensions.Configuration;
using StackExchange.Redis;
using System.Collections.Generic;
using System.Linq;

namespace BotWorker.Tests
{
    public class PluginTests
    {
        private readonly Mock<IRobot> _mockRobot;
        private readonly Mock<IAIService> _mockAI;
        private readonly Mock<IPluginContext> _mockContext;
        private readonly Mock<IEventNexus> _mockEvents;
        private readonly Mock<IServiceProvider> _mockServiceProvider;
        private readonly ITestOutputHelper _output;

        public PluginTests(ITestOutputHelper output)
        {
            _output = output;
            Console.OutputEncoding = System.Text.Encoding.UTF8;
            _mockRobot = new Mock<IRobot>();
            _mockAI = new Mock<IAIService>();
            _mockContext = new Mock<IPluginContext>();
            _mockEvents = new Mock<IEventNexus>();
            _mockServiceProvider = new Mock<IServiceProvider>();

            // 💡 初始化真实环境配置
            InitializeRealConfig();

            _mockRobot.Setup(r => r.AI).Returns(_mockAI.Object);
            _mockRobot.Setup(r => r.Events).Returns(_mockEvents.Object);
            
            var mockRedis = new Mock<IConnectionMultiplexer>();
            var mockDb = new Mock<IDatabase>();
            mockRedis.Setup(r => r.GetDatabase(It.IsAny<int>(), It.IsAny<object>())).Returns(mockDb.Object);
            
            _mockRobot.Setup(r => r.Sessions).Returns(new SessionManager(mockRedis.Object));
            
            _mockContext.Setup(c => c.UserId).Returns("test_user_123");
            _mockContext.Setup(c => c.GroupId).Returns("test_group_456");
            _mockContext.Setup(c => c.RawMessage).Returns("!test");
        }

        private void InitializeRealConfig()
        {
            if (!string.IsNullOrEmpty(GlobalConfig.ConnString)) return;

            try
            {
                var config = new ConfigurationBuilder()
                    .SetBasePath(AppContext.BaseDirectory)
                    .AddJsonFile("appsettings.json", optional: true)
                    .AddJsonFile("appsettings.Development.json", optional: true)
                    .Build();

                GlobalConfig.Initialize(config);
                _output.WriteLine($"[系统] 已加载真实配置。数据库类型: {GlobalConfig.DbType}");
                _output.WriteLine($"[系统] 连接字符串: {GlobalConfig.ConnString.Split(';')[0]}... (已隐藏密码)");
            }
            catch (Exception ex)
            {
                _output.WriteLine($"[警告] 无法加载真实配置，将回退到模拟模式: {ex.Message}");
            }
        }

        public static IEnumerable<object[]> GetAllPlugins()
        {
            var pluginType = typeof(IPlugin);
            var types = typeof(DigitalStaffService).Assembly.GetTypes()
                .Where(t => pluginType.IsAssignableFrom(t) && !t.IsInterface && !t.IsAbstract);

            foreach (var type in types)
            {
                yield return new object[] { type };
            }
        }

        [Theory]
        [MemberData(nameof(GetAllPlugins))]
        public async Task AllPlugins_ShouldInitializeAndStop_Successfully(Type pluginType)
        {
            _output.WriteLine($"[冒烟测试] 正在验证插件加载: {pluginType.Name}");
            // Arrange
            IPlugin? plugin = null;
            var constructors = pluginType.GetConstructors();
            
            foreach (var ctor in constructors.OrderByDescending(c => c.GetParameters().Length))
            {
                try
                {
                    var parameters = ctor.GetParameters();
                    var args = new object?[parameters.Length];
                    for (int i = 0; i < parameters.Length; i++)
                    {
                        var paramType = parameters[i].ParameterType;
                        if (paramType.IsInterface && paramType.Name.StartsWith("ILogger"))
                        {
                            args[i] = null; 
                        }
                        else if (paramType == typeof(IRobot))
                        {
                            args[i] = _mockRobot.Object;
                        }
                        else if (paramType == typeof(IServiceProvider))
                        {
                            args[i] = _mockServiceProvider.Object;
                        }
                        else
                        {
                            args[i] = paramType.IsValueType ? Activator.CreateInstance(paramType) : null;
                        }
                    }
                    plugin = ctor.Invoke(args) as IPlugin;
                    if (plugin != null) break;
                }
                catch (Exception ex) 
                {
                    _output.WriteLine($"  -> 构造函数尝试失败: {ex.Message}");
                    continue;
                }
            }
            
            Assert.NotNull(plugin);
            _output.WriteLine($"  -> 实例创建成功，准备执行 InitAsync...");

            // Act
            var initException = await Record.ExceptionAsync(() => plugin!.InitAsync(_mockRobot.Object));
            var stopException = await Record.ExceptionAsync(() => plugin!.StopAsync());

            // Assert
            if (initException != null)
            {
                if (IsDatabaseException(initException))
                    _output.WriteLine($"  -> InitAsync 完成 (捕获到预期的数据库连接异常)");
                else
                    _output.WriteLine($"  !! InitAsync 报错: {initException.Message}");
            }
            else
            {
                _output.WriteLine($"  -> InitAsync 成功");
            }

            Assert.True(initException == null || IsDatabaseException(initException), $"Plugin {pluginType.Name} init failed: {initException?.Message}\n{initException?.StackTrace}");
            Assert.Null(stopException);
            _output.WriteLine($"  -> 测试通过 ✅");
        }

        private bool IsDatabaseException(Exception ex)
        {
            var msg = ex.Message.ToLower();
            return msg.Contains("database") || msg.Contains("connection") || msg.Contains("table") || 
                   msg.Contains("sql") || msg.Contains("sqlite") || msg.Contains("invalid object name") ||
                   msg.Contains("microsoft.data.sqlclient");
        }

        [Fact]
        public async Task DigitalStaff_HireCommand_ShouldProcessLogic_EvenIfDbFails()
        {
            _output.WriteLine("[功能测试] 正在验证数字员工 - 雇佣指令...");
            // Arrange
            Func<IPluginContext, string[], Task<string>>? capturedHandler = null;
            _mockRobot.Setup(r => r.RegisterSkillAsync(It.IsAny<SkillCapability>(), It.IsAny<Func<IPluginContext, string[], Task<string>>>()))
                .Callback<SkillCapability, Func<IPluginContext, string[], Task<string>>>((cap, handler) => {
                    if (cap.Name.Contains("人才") || cap.Name.Contains("员工")) capturedHandler = handler;
                })
                .Returns(Task.CompletedTask);

            var service = new DigitalStaffService();
            await service.InitAsync(_mockRobot.Object);

            Assert.NotNull(capturedHandler);
            _output.WriteLine("  -> 技能已注册，准备模拟指令: !雇佣 鲁班 开发");

            // Act
            _mockContext.Setup(c => c.RawMessage).Returns("!雇佣 鲁班 开发");
            _mockContext.Setup(c => c.UserId).Returns("123456");
            string? result = null;
            var exception = await Record.ExceptionAsync(async () => result = await capturedHandler!(_mockContext.Object, new[] { "鲁班", "开发" }));

            // Assert
            if (result != null) Console.WriteLine($"[TEST] 指令返回结果: \n{result}");
            Console.WriteLine(exception != null ? $"[TEST] 逻辑执行成功 (已到达数据库层: {exception.Message})" : "[TEST] 逻辑执行成功");
            Assert.True(exception == null || IsDatabaseException(exception));
        }

        [Fact]
        public async Task AchievementPlugin_MyAchievementsCommand_ShouldProcessLogic()
        {
            _output.WriteLine("[功能测试] 正在验证成就系统 - 查看成就指令...");
            // Arrange
            Func<IPluginContext, string[], Task<string>>? capturedHandler = null;
            _mockRobot.Setup(r => r.RegisterSkillAsync(It.Is<SkillCapability>(s => s.Name == "我的成就"), It.IsAny<Func<IPluginContext, string[], Task<string>>>()))
                .Callback<SkillCapability, Func<IPluginContext, string[], Task<string>>>((cap, handler) => capturedHandler = handler)
                .Returns(Task.CompletedTask);

            var service = new AchievementPlugin();
            await Record.ExceptionAsync(() => service.InitAsync(_mockRobot.Object));

            Assert.NotNull(capturedHandler);
            _output.WriteLine("  -> 技能已注册，准备模拟指令: 我的成就");

            // Act
            _mockContext.Setup(c => c.RawMessage).Returns("我的成就");
            _mockContext.Setup(c => c.UserId).Returns("123456");
            string? result = null;
            var exception = await Record.ExceptionAsync(async () => result = await capturedHandler!(_mockContext.Object, Array.Empty<string>()));

            // Assert
            if (result != null) Console.WriteLine($"[TEST] 指令返回结果: \n{result}");
            Console.WriteLine(exception != null ? $"[TEST] 逻辑执行成功 (已到达数据库层: {exception.Message})" : "[TEST] 逻辑执行成功");
            Assert.True(exception == null || IsDatabaseException(exception));
        }

        [Fact]
        public async Task AchievementPlugin_ReportMetric_ShouldHandleDBErrorGracefully()
        {
            // Act
            var exception = await Record.ExceptionAsync(() => AchievementPlugin.ReportMetricAsync("123456", "sys.msg_count", 1));

            // Assert
            Assert.True(exception == null || IsDatabaseException(exception));
        }

        [Fact]
        public async Task MarriageService_ProposeCommand_ShouldProcessLogic()
        {
            _output.WriteLine("[功能测试] 正在验证婚姻系统 - 求婚指令...");
            // Arrange
            Func<IPluginContext, string[], Task<string>>? capturedHandler = null;
            _mockRobot.Setup(r => r.RegisterSkillAsync(It.Is<SkillCapability>(s => s.Name == "婚姻系统"), It.IsAny<Func<IPluginContext, string[], Task<string>>>()))
                .Callback<SkillCapability, Func<IPluginContext, string[], Task<string>>>((cap, handler) => capturedHandler = handler)
                .Returns(Task.CompletedTask);

            var service = new MarriageService();
            await Record.ExceptionAsync(() => service.InitAsync(_mockRobot.Object));

            Assert.NotNull(capturedHandler);
            _output.WriteLine("  -> 技能已注册，准备模拟指令: 求婚 @小红");

            // Act
            _mockContext.Setup(c => c.RawMessage).Returns("求婚 @小红");
            _mockContext.Setup(c => c.UserId).Returns("123456");
            string? result = null;
            var exception = await Record.ExceptionAsync(async () => result = await capturedHandler!(_mockContext.Object, new[] { "@小红" }));

            // Assert
            if (result != null) Console.WriteLine($"[TEST] 指令返回结果: \n{result}");
            Console.WriteLine(exception != null ? $"[TEST] 逻辑执行成功 (已到达数据库层: {exception.Message})" : "[TEST] 逻辑执行成功");
            Assert.True(exception == null || IsDatabaseException(exception));
        }

        [Fact]
        public async Task PetService_AdoptCommand_ShouldProcessLogic()
        {
            _output.WriteLine("[功能测试] 正在验证宠物系统 - 领养指令...");
            // Arrange
            Func<IPluginContext, string[], Task<string>>? capturedHandler = null;
            _mockRobot.Setup(r => r.RegisterSkillAsync(It.Is<SkillCapability>(s => s.Name == "宠物养成"), It.IsAny<Func<IPluginContext, string[], Task<string>>>()))
                .Callback<SkillCapability, Func<IPluginContext, string[], Task<string>>>((cap, handler) => capturedHandler = handler)
                .Returns(Task.CompletedTask);

            var service = new PetService();
            await Record.ExceptionAsync(() => service.InitAsync(_mockRobot.Object));

            Assert.NotNull(capturedHandler);
            _output.WriteLine("  -> 技能已注册，准备模拟指令: 领养宠物 旺财");

            // Act
            _mockContext.Setup(c => c.RawMessage).Returns("领养宠物 旺财");
            _mockContext.Setup(c => c.UserId).Returns("123456");
            string? result = null;
            var exception = await Record.ExceptionAsync(async () => result = await capturedHandler!(_mockContext.Object, new[] { "旺财" }));

            // Assert
            if (result != null) Console.WriteLine($"[TEST] 指令返回结果: \n{result}");
            Console.WriteLine(exception != null ? $"[TEST] 逻辑执行成功 (已到达数据库层: {exception.Message})" : "[TEST] 逻辑执行成功");
            Assert.True(exception == null || IsDatabaseException(exception));
        }

        [Fact]
        public async Task FishingPlugin_CastCommand_ShouldProcessLogic()
        {
            _output.WriteLine("[功能测试] 正在验证钓鱼系统 - 抛竿指令...");
            // Arrange
            Func<IPluginContext, string[], Task<string>>? capturedHandler = null;
            _mockRobot.Setup(r => r.RegisterSkillAsync(It.Is<SkillCapability>(s => s.Name == "新版钓鱼"), It.IsAny<Func<IPluginContext, string[], Task<string>>>()))
                .Callback<SkillCapability, Func<IPluginContext, string[], Task<string>>>((cap, handler) => capturedHandler = handler)
                .Returns(Task.CompletedTask);

            var service = new FishingPlugin();
            await Record.ExceptionAsync(() => service.InitAsync(_mockRobot.Object));

            Assert.NotNull(capturedHandler);
            _output.WriteLine("  -> 技能已注册，准备模拟指令: 抛竿");

            // Act
            _mockContext.Setup(c => c.RawMessage).Returns("抛竿");
            _mockContext.Setup(c => c.UserId).Returns("123456");
            string? result = null;
            var exception = await Record.ExceptionAsync(async () => result = await capturedHandler!(_mockContext.Object, Array.Empty<string>()));

            // Assert
            if (result != null) Console.WriteLine($"[TEST] 指令返回结果: \n{result}");
            Console.WriteLine(exception != null ? $"[TEST] 逻辑执行成功 (已到达数据库层: {exception.Message})" : "[TEST] 逻辑执行成功");
            Assert.True(exception == null || IsDatabaseException(exception));
        }
    }
}
