using System;
using System.Threading.Tasks;
using BotWorker.Common;
using BotWorker.Domain.Entities;
using BotWorker.Domain.Repositories;
using Dapper;

namespace BotWorker.Infrastructure.Persistence.Repositories
{
    public class QuestionInfoRepository : BaseRepository<QuestionInfo>, IQuestionInfoRepository
    {
        public QuestionInfoRepository() : base("Question", GlobalConfig.BaseInfoConnection)
        {
        }

        public async Task<bool> ExistsByQuestionAsync(string question)
        {
            var cleanQuestion = QuestionInfo.GetNew(question);
            return await CountAsync("WHERE question = @question", new { question = cleanQuestion }) > 0;
        }

        public async Task<long> AddQuestionAsync(long botUin, long groupId, long userId, string question)
        {
            var cleanQuestion = QuestionInfo.GetNew(question);
            if (string.IsNullOrEmpty(cleanQuestion) || cleanQuestion.Length > 200)
                return 0;

            var existingId = await GetIdByQuestionAsync(cleanQuestion);
            if (existingId > 0) return existingId;

            var entity = new QuestionInfo
            {
                BotUin = botUin,
                GroupId = groupId,
                UserId = userId,
                Question = cleanQuestion,
                InsertDate = DateTime.Now
            };

            await InsertAsync(entity);
            return await GetIdByQuestionAsync(cleanQuestion);
        }

        public async Task<int> IncrementUsedTimesAsync(long questionId)
        {
            return await IncrementValueAsync("CUsed", 1, "WHERE Id = @questionId", new { questionId });
        }

        public async Task<bool> IsSystemAsync(long questionId)
        {
            return await GetValueAsync<bool>("IsSystem", "WHERE Id = @questionId", new { questionId });
        }

        public async Task<int> AuditAsync(long questionId, int audit2, int isSystem)
        {
            var sql = $"UPDATE {_tableName} SET Audit2 = @audit2, Audit2Date = @now, Audit2By = @auditBy, IsSystem = @isSystem WHERE Id = @questionId";
            using var conn = CreateConnection();
            return await conn.ExecuteAsync(sql, new 
            { 
                audit2, 
                now = DateTime.Now, 
                auditBy = BotInfo.SystemUid, 
                isSystem, 
                questionId 
            });
        }

        public async Task<long> GetIdByQuestionAsync(string question)
        {
            if (question.Length > 200) return 0;
            return await GetValueAsync<long>("Id", "WHERE question = @question", new { question });
        }
    }
}
