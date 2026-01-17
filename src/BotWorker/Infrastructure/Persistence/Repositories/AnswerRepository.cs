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

        public async Task<long> GetGroupAnswerIdAsync(long groupId, long questionId, int length = 0)
        {
            var sql = $"SELECT Id FROM {_tableName} WHERE QuestionId = @questionId AND ABS(audit) = 1 AND ((RobotId = @groupId AND audit2 <> -4) OR audit2 = 3)";
            if (length > 0) sql += " AND LENGTH(answer) >= @length";
            sql += " ORDER BY RANDOM() LIMIT 1";
            
            using var conn = CreateConnection();
            return await conn.ExecuteScalarAsync<long>(sql, new { questionId, groupId, length });
        }

        public async Task<long> GetDefaultAnswerIdAsync(long questionId, long robotId, int length = 0)
        {
            var sql = $"SELECT Id FROM {_tableName} WHERE QuestionId = @questionId AND ABS(audit) = 1 AND RobotId = @robotId AND audit2 >= 0";
            if (length > 0) sql += " AND LENGTH(answer) >= @length";
            sql += " ORDER BY RANDOM() LIMIT 1";
            
            using var conn = CreateConnection();
            return await conn.ExecuteScalarAsync<long>(sql, new { questionId, robotId, length });
        }

        public async Task<long> GetDefaultAnswerAtIdAsync(long questionId, int length = 0)
        {
            var sql = $@"SELECT Id FROM {_tableName} 
                        WHERE QuestionId = @questionId AND ABS(audit) = 1
                        AND (
                            Id IN (SELECT Id FROM {_tableName} WHERE QuestionId = @questionId AND Audit2 >= 1 ORDER BY ((COALESCE(GoonTimes, 0) + 1)/(COALESCE(UsedTimes, 0) + 1)) DESC LIMIT 20)
                            OR 
                            Id IN (SELECT Id FROM {_tableName} WHERE QuestionId = @questionId AND Audit2 >= 1 AND UsedTimes < 100 ORDER BY UsedTimes DESC LIMIT 10)
                        )
                        ORDER BY RANDOM() LIMIT 1";
            
            using var conn = CreateConnection();
            return await conn.ExecuteScalarAsync<long>(sql, new { questionId, length });
        }

        public async Task<long> GetAllAnswerAuditIdAsync(long questionId, int length = 0)
        {
            var sql = $"SELECT Id FROM {_tableName} WHERE QuestionId = @questionId AND ABS(audit) = 1 AND audit2 = 0";
            if (length > 0) sql += " AND LENGTH(answer) >= @length";
            sql += " ORDER BY RANDOM() LIMIT 1";
            
            using var conn = CreateConnection();
            return await conn.ExecuteScalarAsync<long>(sql, new { questionId, length });
        }

        public async Task<long> GetAllAnswerNotAuditIdAsync(long questionId, int length = 0)
        {
            var sql = $"SELECT Id FROM {_tableName} WHERE QuestionId = @questionId AND ABS(audit) = 0";
            if (length > 0) sql += " AND LENGTH(answer) >= @length";
            sql += " ORDER BY RANDOM() LIMIT 1";
            
            using var conn = CreateConnection();
            return await conn.ExecuteScalarAsync<long>(sql, new { questionId, length });
        }

        public async Task<long> GetAllAnswerIdAsync(long questionId, int length = 0)
        {
            var sql = $"SELECT Id FROM {_tableName} WHERE QuestionId = @questionId";
            if (length > 0) sql += " AND LENGTH(answer) >= @length";
            sql += " ORDER BY RANDOM() LIMIT 1";
            
            using var conn = CreateConnection();
            return await conn.ExecuteScalarAsync<long>(sql, new { questionId, length });
        }

        public async Task<long> GetStoryIdAsync()
        {
            var sql = $"SELECT Id FROM {_tableName} WHERE QuestionId IN (50701, 545) AND LENGTH(answer) > 40 AND ABS(audit) = 1 AND audit2 >= 0 ORDER BY RANDOM() LIMIT 1";
            using var conn = CreateConnection();
            return await conn.ExecuteScalarAsync<long>(sql);
        }

        public async Task<long> GetGhostStoryIdAsync()
        {
            var sql = $"SELECT Id FROM {_tableName} WHERE QuestionId IN (SELECT Id FROM Question WHERE question like '鬼故事%') AND LENGTH(answer) > 40 AND ABS(audit) = 1 AND audit2 > -3 ORDER BY RANDOM() LIMIT 1";
            using var conn = CreateConnection();
            return await conn.ExecuteScalarAsync<long>(sql);
        }

        public async Task<long> GetCoupletsIdAsync()
        {
            var sql = $"SELECT Id FROM {_tableName} WHERE QuestionId IN (SELECT Id FROM Question WHERE question LIKE '%对联%') AND LENGTH(answer) > 12 AND ABS(audit) = 1 AND audit2 > -3 ORDER BY RANDOM() LIMIT 1";
            using var conn = CreateConnection();
            return await conn.ExecuteScalarAsync<long>(sql);
        }

        public async Task<long> GetChouqianIdAsync()
        {
            var sql = $"SELECT Id FROM {_tableName} WHERE RobotId = 286946883 and QuestionId = 225781 AND AUDIT2 > 0 ORDER BY RANDOM() LIMIT 1";
            using var conn = CreateConnection();
            return await conn.ExecuteScalarAsync<long>(sql);
        }

        public async Task<long> GetJieqianAnswerIdAsync(long groupId, long userId)
        {
            var sql = $@"SELECT AnswerId FROM SendMessage 
                        WHERE GroupId = @groupId AND UserId = @userId 
                        AND AnswerId IN (SELECT Id FROM {_tableName} WHERE RobotId = 286946883 and QuestionId = 225781) 
                        ORDER BY Id DESC LIMIT 1";
            using var conn = CreateConnection();
            return await conn.ExecuteScalarAsync<long>(sql, new { groupId, userId });
        }

        public async Task<long> GetAnswerIdByParentAsync(long parentId)
        {
            var sql = $"SELECT Id FROM {_tableName} WHERE parentanswer = @parentId";
            using var conn = CreateConnection();
            return await conn.ExecuteScalarAsync<long>(sql, new { parentId });
        }

        public async Task<long> GetDatiIdAsync(string keyword)
        {
            var sql = $"SELECT Id FROM {_tableName} WHERE RobotId = 453174086 AND question LIKE @keywordPattern AND question NOT LIKE '%答案%' AND ABS(audit) = 1 AND audit2 <> -4 ORDER BY RANDOM() LIMIT 1";
            using var conn = CreateConnection();
            return await conn.ExecuteScalarAsync<long>(sql, new { keywordPattern = $"%{keyword}%" });
        }
    }
}
