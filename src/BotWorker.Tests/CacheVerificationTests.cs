using Moq;
using Xunit;
using BotWorker.Infrastructure.Caching;
using BotWorker.Infrastructure.Persistence.ORM;
using System.Data;

namespace BotWorker.Tests
{
    public class TestEntity : MetaData<TestEntity>
    {
        public override string TableName => "test_table";
        public override string KeyField => "id";
        
        [Column]
        public int id { get; set; }
        
        [Column]
        public string Name { get; set; } = "";

        [Column]
        [HighFrequency]
        public int credit { get; set; }
    }

    public class CacheVerificationTests
    {
        [Fact]
        public async Task TestHighFrequencyField_DoesNotInvalidateRowCache()
        {
            var mockCache = new Mock<ICacheService>();
            MetaData.CacheService = mockCache.Object;

            // 模拟数据库执行成功 (捕获异常，因为我们没有真实的数据库连接)
            try
            {
                await TestEntity.PlusAsync("credit", 10, 1);
            }
            catch (InvalidOperationException)
            {
                // 忽略 ConnectionString 未初始化错误，我们只关心缓存调用
            }
            
            // 验证：行缓存 Remove 不应该被调用 (key 格式为 MetaData:[sz84_robot].[dbo].[test_table]:Id:1)
            mockCache.Verify(c => c.RemoveAsync(It.Is<string>(s => s.EndsWith("Id:1"))), Times.Never);
            // 验证：列缓存 Remove 应该被调用 (key 格式为 MetaData:[sz84_robot].[dbo].[test_table]:Id:credit_1)
            mockCache.Verify(c => c.RemoveAsync(It.Is<string>(s => s.Contains("Id:credit_1"))), Times.Once);

            // 2. 测试普通字段更新
            try
            {
                await TestEntity.SetValueAsync("Name", "NewName", 1);
            }
            catch (InvalidOperationException) { }

            mockCache.Verify(c => c.RemoveAsync(It.Is<string>(s => s.EndsWith("Id:1"))), Times.Once);
        }

        [Fact]
        public async Task TestCacheKeyGeneration()
        {
            var mockCache = new Mock<ICacheService>();
            MetaData.CacheService = mockCache.Object;

            // 验证行级缓存键生成
            // MetaData:{FullName}:Id:{keys}
            // 注意：TestEntity 的 FullName 取决于 DataBase 和 TableName
            // 默认 DataBase 是 "sz84_robot"，TableName 是 "test_table"
            // 所以 FullName 应该是 "[sz84_robot].[dbo].[test_table]"
            
            var entity = new TestEntity { id = 1, Name = "Test" };
            
            // 手动触发失效，验证键是否正确
            await TestEntity.InvalidateCacheAsync(1);
            mockCache.Verify(c => c.RemoveAsync("MetaData:[sz84_robot].[dbo].[test_table]:Id:1"), Times.Once);

            // 验证列级缓存键生成
            await TestEntity.InvalidateFieldCacheAsync("Name", 1);
            mockCache.Verify(c => c.RemoveAsync("MetaData:[sz84_robot].[dbo].[test_table]:Id:Name_1"), Times.Once);
        }

        [Fact]
        public async Task TestGetByKeyAsync_CachesResult()
        {
            var mockCache = new Mock<ICacheService>();
            MetaData.CacheService = mockCache.Object;

            // 模拟缓存中没有数据
            mockCache.Setup(c => c.GetAsync<TestEntity>(It.IsAny<string>()))
                     .ReturnsAsync((TestEntity?)null);

            // 注意：这里会尝试调用数据库，因为缓存未命中
            // 由于我们没有配置数据库，这里会报错，但我们可以捕获它
            // 我们主要验证它是否尝试从缓存读取了
            try
            {
                await TestEntity.GetByKeyAsync(1);
            }
            catch
            {
                // 忽略数据库连接错误
            }

            mockCache.Verify(c => c.GetAsync<TestEntity>("MetaData:[sz84_robot].[dbo].[test_table]:Id:1"), Times.Once);
        }

        [Fact]
        public async Task TestGetCached_CachesResult()
        {
            var mockCache = new Mock<ICacheService>();
            MetaData.CacheService = mockCache.Object;

            // GetCached 使用的是同步方法 GetOrAdd
            // 验证它是否调用了 GetOrAdd
            try
            {
                TestEntity.GetCached<string>("Name", 1);
            }
            catch
            {
                // 忽略数据库连接错误
            }

            mockCache.Verify(c => c.GetOrAdd(
                "MetaData:[sz84_robot].[dbo].[test_table]:Id:Name_1", 
                It.IsAny<Func<string>>(), 
                It.IsAny<TimeSpan?>()), 
                Times.Once);
        }
    }
}
