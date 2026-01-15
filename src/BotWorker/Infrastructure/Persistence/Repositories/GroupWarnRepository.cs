using System.Threading.Tasks;
using BotWorker.Domain.Entities;
using BotWorker.Domain.Repositories;
using Dapper.Contrib.Extensions;

namespace BotWorker.Infrastructure.Persistence.Repositories
{
    public class GroupWarnRepository : BaseRepository<GroupWarn>, IGroupWarnRepository
    {
        public GroupWarnRepository(string? connectionString = null) : base("Warn", connectionString)
        {
        }

        public async Task<long> CountByGroupAndUserAsync(long groupId, long userId)
        {
            using var conn = CreateConnection();
            string sql = $"SELECT COUNT(*) FROM {_tableName} WHERE \"GroupId\" = @groupId AND \"UserId\" = @userId";
            return await conn.ExecuteScalarAsync<long>(sql, new { groupId, userId });
        }

        public async Task<int> DeleteByGroupAndUserAsync(long groupId, long userId)
        {
            using var conn = CreateConnection();
            string sql = $"DELETE FROM {_tableName} WHERE \"GroupId\" = @groupId AND \"UserId\" = @userId";
            return await conn.ExecuteAsync(sql, new { groupId, userId });
        }
    }
}
