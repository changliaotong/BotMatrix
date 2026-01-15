using System;
using System.Threading.Tasks;
using BotWorker.Domain.Entities;
using BotWorker.Domain.Repositories;
using Dapper;

namespace BotWorker.Infrastructure.Persistence.Repositories
{
    public class GroupSendMessageRepository : BaseRepository<GroupSendMessage>, IGroupSendMessageRepository
    {
        public async Task<int> UserCountAsync(long groupId)
        {
            string sql = $"SELECT COUNT(DISTINCT UserId) FROM {_tableName} WHERE DATEDIFF(SECOND, InsertDate, GETDATE()) < 60 AND GroupId = @groupId";
            using var conn = CreateConnection();
            return await conn.ExecuteScalarAsync<int>(sql, new { groupId });
        }

        public async Task<int> AppendAsync(GroupSendMessage entity)
        {
            return await InsertAsync(entity);
        }
    }
}
