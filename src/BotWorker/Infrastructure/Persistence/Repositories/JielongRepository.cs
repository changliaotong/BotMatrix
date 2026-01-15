using System;
using System.Collections.Generic;
using System.Data;
using System.Linq;
using System.Threading.Tasks;
using BotWorker.Common;
using BotWorker.Domain.Entities;
using BotWorker.Domain.Repositories;
using BotWorker.Modules.Games;
using Dapper;

namespace BotWorker.Infrastructure.Persistence.Repositories
{
    public class JielongRepository : BaseRepository<Jielong>, IJielongRepository
    {
        public JielongRepository(string? connectionString = null)
            : base("jielong_logs", connectionString ?? GlobalConfig.KnowledgeBaseConnection)
        {
        }

        public async Task<string?> GetRandomChengyuAsync()
        {
            const string sql = "SELECT chengyu FROM chengyu ORDER BY RANDOM() LIMIT 1";
            using var conn = CreateConnection();
            return await conn.ExecuteScalarAsync<string>(sql);
        }

        public async Task<string?> GetChengYuByPinyinAsync(string pinyin, long groupId)
        {
            const string sql = @"
                SELECT chengyu FROM chengyu 
                WHERE pinyin LIKE @pinyin 
                AND chengyu NOT IN (
                    SELECT chengyu FROM jielong_logs 
                    WHERE group_id = @groupId 
                    AND id > (
                        SELECT id FROM jielong_logs 
                        WHERE group_id = @groupId AND game_no = 1 
                        ORDER BY insert_date DESC LIMIT 1
                    )
                )
                ORDER BY RANDOM() LIMIT 1";
            using var conn = CreateConnection();
            return await conn.ExecuteScalarAsync<string>(sql, new { pinyin = pinyin + " %", groupId });
        }

        public async Task<bool> IsDupAsync(long groupId, long userId, string chengYu)
        {
            string sql;
            if (groupId == 0)
            {
                sql = @"
                    SELECT COUNT(1) FROM jielong_logs 
                    WHERE group_id = @groupId AND user_id = @userId AND chengyu = @chengYu 
                    AND id > (
                        SELECT id FROM jielong_logs 
                        WHERE group_id = @groupId AND user_id = @userId AND game_no = 1 
                        ORDER BY id DESC LIMIT 1
                    )";
            }
            else
            {
                sql = @"
                    SELECT COUNT(1) FROM jielong_logs 
                    WHERE group_id = @groupId AND chengyu = @chengYu 
                    AND id > (
                        SELECT id FROM jielong_logs 
                        WHERE group_id = @groupId AND game_no = 1 
                        ORDER BY id DESC LIMIT 1
                    )";
            }
            using var conn = CreateConnection();
            return await conn.ExecuteScalarAsync<int>(sql, new { groupId, userId, chengYu }) > 0;
        }

        public async Task<int> GetMaxIdAsync(long groupId)
        {
            const string sql = "SELECT COALESCE(MAX(id), 0) FROM jielong_logs WHERE group_id = @groupId AND game_no = 1";
            using var conn = CreateConnection();
            return await conn.ExecuteScalarAsync<int>(sql, new { groupId });
        }

        public async Task<int> GetCountAsync(long groupId, long userId)
        {
            int maxId = await GetMaxIdAsync(groupId);
            const string sql = "SELECT COUNT(id) FROM jielong_logs WHERE user_id = @userId AND id >= @maxId";
            using var conn = CreateConnection();
            return await conn.ExecuteScalarAsync<int>(sql, new { userId, maxId });
        }

        public async Task<long> GetCreditAddAsync(long userId)
        {
            const string sql = "SELECT COALESCE(SUM(credit), 0) FROM jielong_logs WHERE user_id = @userId AND credit > 0";
            using var conn = CreateConnection();
            return await conn.ExecuteScalarAsync<long>(sql, new { userId });
        }

        public async Task<bool> InGameAsync(long groupId, long userId)
        {
            // This logic is mostly handled in the entity/plugin, but here we can check the state
            // For now, return false as placeholder or implement if needed
            return false; 
        }

        public async Task<string> GetCurrentChengYuAsync(long groupId, long userId)
        {
            // Similar to InGameAsync, this might be better handled in the entity using UserInfo/GroupInfo
            return string.Empty;
        }

        public async Task<int> AppendAsync(long groupId, long userId, string userName, string chengYu, int gameNo)
        {
            const string sql = @"
                INSERT INTO jielong_logs (group_id, user_id, user_name, chengyu, game_no, insert_date)
                VALUES (@groupId, @userId, @userName, @chengYu, @gameNo, CURRENT_TIMESTAMP)";
            using var conn = CreateConnection();
            return await conn.ExecuteAsync(sql, new { groupId, userId, userName, chengYu, gameNo });
        }

        public async Task<int> EndGameAsync(long groupId, long userId)
        {
            // Implementation depends on how games are ended in the DB
            return 0;
        }
    }
}
