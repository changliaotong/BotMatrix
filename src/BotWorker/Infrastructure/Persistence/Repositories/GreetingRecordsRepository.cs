using System;
using System.Threading.Tasks;
using BotWorker.Domain.Entities;
using BotWorker.Domain.Repositories;
using Dapper;

namespace BotWorker.Infrastructure.Persistence.Repositories
{
    public class GreetingRecordsRepository : BaseRepository<GreetingRecords>, IGreetingRecordsRepository
    {
        public GreetingRecordsRepository() : base("GreetingRecords")
        {
        }

        public async Task<int> AppendAsync(long botQQ, long groupId, string groupName, long qq, string name, int greetingType = 0)
        {
            var hour = greetingType == 0 ? -3 : -5;
            var logicalDateSql = $"SELECT (CURRENT_TIMESTAMP + INTERVAL '{hour} hour')::date";
            
            using var conn = CreateConnection();
            var logicalDate = await conn.ExecuteScalarAsync<DateTime>(logicalDateSql);

            var record = new GreetingRecords
            {
                BotQQ = botQQ,
                GroupId = groupId,
                GroupName = groupName,
                QQ = qq,
                Name = name,
                GreetingType = greetingType,
                LogicalDate = logicalDate
            };

            await InsertAsync(record);
            return 1;
        }

        public async Task<bool> ExistsAsync(long groupId, long qq, int greetingType = 0)
        {
            var hour = greetingType == 0 ? -3 : -5;
            var sql = $"SELECT EXISTS(SELECT 1 FROM {_tableName} WHERE GroupId = @groupId AND QQ = @qq AND GreetingType = @greetingType AND LogicalDate = (CURRENT_TIMESTAMP + INTERVAL '{hour} hour')::date)";
            
            using var conn = CreateConnection();
            return await conn.ExecuteScalarAsync<bool>(sql, new { groupId, qq, greetingType });
        }

        public async Task<int> GetCountAsync(int greetingType = 0)
        {
            var hour = greetingType == 0 ? -3 : -5;
            var sql = $"SELECT COUNT(Id) + 1 FROM {_tableName} WHERE GreetingType = @greetingType AND LogicalDate = (CURRENT_TIMESTAMP + INTERVAL '{hour} hour')::date";
            
            using var conn = CreateConnection();
            return await conn.ExecuteScalarAsync<int>(sql, new { greetingType });
        }

        public async Task<int> GetCountAsync(long groupId, int greetingType = 0)
        {
            var hour = greetingType == 0 ? -3 : -5;
            var sql = $"SELECT COUNT(Id) + 1 FROM {_tableName} WHERE GroupId = @groupId AND GreetingType = @greetingType AND LogicalDate = (CURRENT_TIMESTAMP + INTERVAL '{hour} hour')::date";
            
            using var conn = CreateConnection();
            return await conn.ExecuteScalarAsync<int>(sql, new { groupId, greetingType });
        }
    }
}
