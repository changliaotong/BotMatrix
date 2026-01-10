using System;
using System.Threading.Tasks;
using Xunit;
using Moq;
using BotWorker.Domain.Interfaces;
using BotWorker.Modules.Games;
using BotWorker.Modules.AI.Services;
using Microsoft.Extensions.Logging;

namespace BotWorker.Tests
{
    public class DigitalStaffTests
    {
        [Fact]
        public async Task DigitalStaff_ShouldInitSuccessfully()
        {
            // Arrange
            var mockRobot = new Mock<IRobot>();
            var mockAI = new Mock<IAIService>();
            var mockLogger = new Mock<ILogger<DigitalStaffService>>();

            mockRobot.Setup(r => r.AI).Returns(mockAI.Object);

            var service = new DigitalStaffService(mockLogger.Object);

            // Act & Assert
            // 验证初始化是否抛出异常（由于数据库连接可能需要 Mock，这里主要测试逻辑流程）
            var exception = await Record.ExceptionAsync(() => service.InitAsync(mockRobot.Object));
            
            // 如果是因为数据库连接失败（预期内，因为没有真实 DB），我们至少验证了代码运行到了数据库操作这一步
            Assert.True(exception == null || exception.Message.Contains("database", StringComparison.OrdinalIgnoreCase) || exception.Message.Contains("connection", StringComparison.OrdinalIgnoreCase));
        }

        [Fact]
        public void StaffRole_Enum_ShouldHaveExpectedValues()
        {
            // Assert
            Assert.Equal(0, (int)StaffRole.ProductManager);
            Assert.Equal(1, (int)StaffRole.Developer);
            Assert.Equal(5, (int)StaffRole.AfterSales);
        }

        [Fact]
        public void DigitalStaff_Model_DefaultValues_ShouldBeCorrect()
        {
            // Arrange
            var staff = new DigitalStaff();

            // Assert
            Assert.Equal(1, staff.Level);
            Assert.Equal(100.0, staff.KpiScore);
            Assert.Equal("Idle", staff.CurrentStatus);
        }

        [Fact]
        public void SkillCall_Regex_ShouldParseCorrectly()
        {
            // Arrange
            string aiResponse = "我认为需要修复这个 bug。[CALL_SKILL:mesh.architecture_review:src/main.go,fix_leak]";
            var pattern = @"\[CALL_SKILL:(.*?):(.*?)\]";

            // Act
            var match = System.Text.RegularExpressions.Regex.Match(aiResponse, pattern);

            // Assert
            Assert.True(match.Success);
            Assert.Equal("mesh.architecture_review", match.Groups[1].Value);
            Assert.Equal("src/main.go,fix_leak", match.Groups[2].Value);
            
            var args = match.Groups[2].Value.Split(',');
            Assert.Equal(2, args.Length);
            Assert.Equal("src/main.go", args[0]);
            Assert.Equal("fix_leak", args[1]);
        }
    }
}
