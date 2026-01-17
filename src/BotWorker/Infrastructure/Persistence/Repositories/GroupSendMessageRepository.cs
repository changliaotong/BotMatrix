using System;
using System.Collections.Generic;
using System.Threading.Tasks;
using BotWorker.Domain.Entities;
using BotWorker.Domain.Repositories;
using BotWorker.Modules.AI.Models;
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

        public async Task<IEnumerable<ChatHistoryItem>> GetChatHistoryAsync(long groupId, long userId, bool isMultAI, int context)
        {
            var query = $"SELECT {SqlTop(context)} Question, CASE WHEN IsAI = 1 THEN AnswerAi ELSE Message END AS Answer, UserName FROM {_tableName} " +
                        $"WHERE (AnswerId <> 0 or IsAI = 1) AND GroupId = @groupId {(isMultAI ? "" : "AND UserId = @userId")} " +
                        $"AND ABS({SqlDateDiff("HOUR", SqlDateTime, "InsertDate")}) <= 24 ORDER BY Id DESC {SqlLimit(context)}";

            using var conn = CreateConnection();
            return await conn.QueryAsync<ChatHistoryItem>(query, new { groupId, userId });
        }
    }
}
