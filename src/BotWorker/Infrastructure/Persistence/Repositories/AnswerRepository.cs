using System;
using System.Threading.Tasks;
using System.Data;
using BotWorker.Domain.Entities;
using BotWorker.Domain.Repositories;
using Dapper;

namespace BotWorker.Infrastructure.Persistence.Repositories
{
    public class AnswerRepository : BaseRepository<AnswerInfo>, IAnswerRepository
    {
        public AnswerRepository(string? connectionString = null) 
            : base("Answer", connectionString ?? GlobalConfig.BaseInfoConnection)
        {
        }

        public async Task<long> AppendAsync(long botUin, long groupId, long qq, long robotId, long questionId, string textQuestion, string textAnswer, int audit, long credit, int audit2, string audit2Info, IDbTransaction? trans = null)
        {
            var answer = new AnswerInfo
            {
                BotUin = botUin,
                GroupId = groupId,
                UserId = qq,
                RobotId = robotId,
                QuestionId = questionId,
                Question = textQuestion,
                Answer = textAnswer,
                Audit = audit,
                Credit = (int)credit,
                Audit2 = audit2,
                Audit2Info = audit2Info,
                InsertDate = DateTime.Now,
                UpdateDate = DateTime.Now
            };
            return await InsertAsync(answer, trans);
        }

        public async Task<bool> ExistsAsync(long questionId, string textAnswer, long groupId)
        {
            string func = "remove_biaodian"; // Assuming Postgres function exists as per original code
            string sql = $"SELECT COUNT(1) FROM {_tableName} WHERE QuestionId = @questionId AND RobotId = @groupId AND {func}(Answer) = {func}(@textAnswer)";
            using var conn = CreateConnection();
            return await conn.ExecuteScalarAsync<int>(sql, new { questionId, groupId, textAnswer }) > 0;
        }

        public async Task<bool> ExistsAsync(long qqRobot, long questionId, string answer)
        {
            string func = "remove_biaodian";
            string sql = $"SELECT COUNT(1) FROM {_tableName} WHERE QuestionId = @questionId AND {func}(Answer) = {func}(@answer)";
            using var conn = CreateConnection();
            return await conn.ExecuteScalarAsync<int>(sql, new { questionId, answer }) > 0;
        }

        public async Task<long> CountAnswerAsync(long questionId)
        {
            return await CountAsync("WHERE QuestionId = @questionId", new { questionId });
        }

        public async Task<int> IncrementUsedTimesAsync(long answerId)
        {
            string sql = $"UPDATE {_tableName} SET UsedTimes = UsedTimes + 1 WHERE Id = @answerId";
            using var conn = CreateConnection();
            return await conn.ExecuteAsync(sql, new { answerId });
        }

        public async Task<int> AuditAsync(long answerId, int audit, long qq)
        {
            string sql = $"UPDATE {_tableName} SET Audit = @audit, AuditBy = @qq, AuditDate = CURRENT_TIMESTAMP WHERE Id = @answerId";
            using var conn = CreateConnection();
            return await conn.ExecuteAsync(sql, new { audit, qq, answerId });
        }
    }
}
