using System;
using System.Threading.Tasks;
using Xunit;
using Xunit.Abstractions;
using Moq;
using BotWorker.Domain.Interfaces;
using BotWorker.Infrastructure.Persistence.Database;
using BotWorker.Common;
using BotWorker.Modules.Games;
using BotWorker.Domain.Entities;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using BotWorker.Infrastructure.Persistence.ORM;
using Microsoft.Extensions.Configuration;

namespace BotWorker.Tests
{
    [Collection("DatabaseTests")]
    public class GiftSystemTests : IAsyncLifetime
    {
        private readonly Mock<IRobot> _mockRobot;
        private readonly Mock<IPluginContext> _mockContext;
        private readonly ITestOutputHelper _output;
        private readonly GiftService _service;

        public GiftSystemTests(ITestOutputHelper output)
        {
            _output = output;
            _mockRobot = new Mock<IRobot>();
            _mockContext = new Mock<IPluginContext>();
            _service = new GiftService();

            // 初始化配置以支持数据库操作 (如果是集成测试)
            InitializeRealConfig();
        }

    public async Task InitializeAsync()
    {
        _output.WriteLine("开始初始化测试数据库...");

        try 
        {
            // 初始化数据库表和默认数据
            await _service.InitAsync(_mockRobot.Object);
            _output.WriteLine("InitAsync 完成。");
        }
        catch (Exception ex)
        {
            _output.WriteLine($"InitAsync 失败: {ex.Message}\n{ex.StackTrace}");
        }

        // 确保测试数据存在
        var gift = await GiftStoreItem.GetByNameAsync("鲜花");
        _output.WriteLine($"检查 '鲜花': {(gift != null ? "已存在" : "不存在")}");
        
        if (gift == null)
        {
            _output.WriteLine("尝试手动插入缺少的礼物数据...");
            await new GiftStoreItem { GiftName = "鲜花", GiftCredit = 50, GiftType = 1, IsValid = true }.InsertAsync();
            await new GiftStoreItem { GiftName = "跑车", GiftCredit = 10000, GiftType = 1, IsValid = true }.InsertAsync();
        }

        var count = await GiftStoreItem.CountAsync();
        _output.WriteLine($"当前有效礼物总数: {count}");
        if (count > 0)
        {
            var all = await GiftStoreItem.QueryWhere("IsValid = 1");
            foreach (var item in all)
            {
                _output.WriteLine($"- {item.GiftName} ({item.GiftCredit} 积分)");
            }
        }
    }

        public Task DisposeAsync() => Task.CompletedTask;

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
                _output.WriteLine($"[DEBUG] 配置初始化完成。数据库类型: {GlobalConfig.DbType}");
                _output.WriteLine($"[DEBUG] 连接字符串长度: {GlobalConfig.ConnString.Length}");
                
                // 测试连接
                try {
                    using var conn = BotWorker.Infrastructure.Persistence.Database.DbProviderFactory.CreateConnection();
                    conn.Open();
                    _output.WriteLine("[DEBUG] 数据库连接成功！");
                } catch (Exception ex) {
                    _output.WriteLine($"[ERROR] 数据库连接失败: {ex.Message}");
                }
            }
            catch (Exception ex)
            {
                _output.WriteLine($"[警告] 无法加载真实配置: {ex.Message}");
            }
        }

        [Fact]
        public async Task GiftStore_ShouldReturnList()
        {
            // Arrange
            _mockContext.Setup(c => c.RawMessage).Returns("礼物商店");

            // Act
            var result = await _service.HandleCommandAsync(_mockContext.Object, Array.Empty<string>());

            // Assert
            Assert.Contains("【礼物商店】", result);
            _output.WriteLine($"商店列表结果:\n{result}");
        }

        [Fact]
        public async Task BuyGift_WithInsufficientCredit_ShouldFail()
        {
            // Arrange
            string userId = "test_user_fail_" + Guid.NewGuid().ToString().Substring(0, 8);
            _mockContext.Setup(c => c.UserId).Returns(userId);
            _mockContext.Setup(c => c.RawMessage).Returns("购买礼物 跑车 1");
            _mockContext.Setup(c => c.BotId).Returns("123456");

            // 创建一个低积分的用户
            var user = new UserInfo { UserOpenId = userId, Credit = 10, BotUin = 123456 };
            await user.InsertAsync();

            // Act
            var result = await _service.HandleCommandAsync(_mockContext.Object, new[] { "跑车", "1" });

            // Assert
            Assert.Contains("积分不足", result);
            _output.WriteLine($"购买失败结果: {result}");

            // Cleanup
            await user.DeleteAsync();
        }

    [Fact]
    public async Task BuyGift_WithSufficientCredit_ShouldSucceed()
    { 
        // Arrange
        string userId = (100000 + new Random().Next(900000)).ToString();
        _mockContext.Setup(c => c.UserId).Returns(userId);
        _mockContext.Setup(c => c.RawMessage).Returns("购买礼物 鲜花 2");
        _mockContext.Setup(c => c.BotId).Returns("123456");

        // 使用 AddCreditAsync 来初始化用户积分，这更符合系统逻辑
        long numericUserId = long.Parse(userId);
        await UserInfo.AddCreditAsync(123456, 0, "测试群", numericUserId, "测试用户", 1000, "测试初始化");

        // Act
        var result = await _service.HandleCommandAsync(_mockContext.Object, new[] { "鲜花", "2" });

        // Assert
        Assert.Contains("购买成功", result);
        _output.WriteLine($"购买成功结果: {result}");

        // 检查背包
        _mockContext.Setup(c => c.RawMessage).Returns("我的背包");
        var backpackResult = await _service.HandleCommandAsync(_mockContext.Object, Array.Empty<string>());
        Assert.Contains("鲜花 x2", backpackResult);

        // Cleanup
        await SQLConn.ExecAsync($"DELETE FROM [sz84_robot].[dbo].[User] WHERE Id = {numericUserId}");
        var items = await GiftBackpack.QueryWhere($"UserId = '{userId}'", (System.Data.IDbTransaction?)null);
        foreach (var item in items) await item.DeleteAsync();
    }

    [Fact]
    public async Task SendGift_ToUser_ShouldSucceed()
    { 
        // Arrange
        string fromUserId = (100000 + new Random().Next(400000)).ToString();
        string toUserId = (500000 + new Random().Next(400000)).ToString();
        
        _mockContext.Setup(c => c.UserId).Returns(fromUserId);
        _mockContext.Setup(c => c.BotId).Returns("123456");
        _mockContext.Setup(c => c.MentionedUsers).Returns(new List<MentionedUser> { new MentionedUser { UserId = toUserId, Name = "接收者" } });
        _mockContext.Setup(c => c.RawMessage).Returns($"送礼物 [CQ:at,qq={toUserId}] 鲜花 1");

        // 初始化用户和背包
        long fromId = long.Parse(fromUserId);
        long toId = long.Parse(toUserId);
        
        // 确保用户存在且积分为0
        await UserInfo.AddCreditAsync(123456, 0, "测试群", fromId, "赠送者", 100, "测试初始化");
        await UserInfo.AddCreditAsync(123456, 0, "测试群", toId, "接收者", 0, "测试初始化");
        
        // 获取初始积分
        long initialToCredit = await UserInfo.GetCreditAsync(123456, 0, toId);
        _output.WriteLine($"接收者初始积分: {initialToCredit}");

        // 获取鲜花的 ID
        var gift = await GiftStoreItem.GetByNameAsync("鲜花");
        Assert.NotNull(gift);

        var backpack = new GiftBackpack { UserId = fromUserId, GiftId = gift.Id, ItemCount = 5 };
        await backpack.InsertAsync();

        // Act
        var result = await _service.HandleCommandAsync(_mockContext.Object, new[] { $"[CQ:at,qq={toUserId}]", "鲜花", "1" });

        // Assert
        Assert.Contains("赠送成功", result);
        _output.WriteLine($"赠送结果: {result}");

        // 检查接收者积分 (回馈 50% * 50 = 25)
        long finalToCredit = await UserInfo.GetCreditAsync(123456, 0, toId);
        _output.WriteLine($"接收者最终积分: {finalToCredit}");
        
        Assert.Equal(initialToCredit + 25, finalToCredit);

        // Cleanup
        await SQLConn.ExecAsync($"DELETE FROM [sz84_robot].[dbo].[User] WHERE Id IN ({fromId}, {toId})");
        await backpack.DeleteAsync();
        var logs = await GiftRecord.QueryWhere($"UserId = '{fromUserId}'", (System.Data.IDbTransaction?)null);
        foreach (var log in logs) await log.DeleteAsync();
    }
    }
}
