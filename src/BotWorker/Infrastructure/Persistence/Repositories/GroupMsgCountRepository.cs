using System;
using System.Collections.Generic;
using System.Data;
using System.Linq;
using System.Threading.Tasks;
using BotWorker.Domain.Entities;
using BotWorker.Domain.Repositories;
using Dapper;

namespace BotWorker.Infrastructure.Persistence.Repositories
{
    public class GroupMsgCountRepository : BaseRepository<GroupMsgCount>, IGroupMsgCountRepository
    {
        public GroupMsgCountRepository() : base("msgcount")
        {
        }

        public async Task<bool> ExistTodayAsync(long groupId, long userId)
        {
            const string sql = "SELECT EXISTS(SELECT 1 FROM msgcount WHERE group_id = @groupId AND user_id = @userId AND c_date = CURRENT_DATE)";
            using var conn = CreateConnection();
            return await conn.ExecuteScalarAsync<bool>(sql, new { groupId, userId });
        }

        public async Task<int> AppendAsync(long botUin, long groupId, string groupName, long userId, string name)
        {
            const string sql = @"
                INSERT INTO msgcount (bot_uin, group_id, group_name, user_id, user_name, c_date, msg_date, c_msg) 
                VALUES (@botUin, @groupId, @groupName, @userId, @name, CURRENT_DATE, CURRENT_TIMESTAMP, 1)";
            using var conn = CreateConnection();
            return await conn.ExecuteAsync(sql, new { botUin, groupId, groupName, userId, name });
        }

        public async Task<int> UpdateAsync(long botUin, long groupId, string groupName, long userId, string name)
        {
            if (!await ExistTodayAsync(groupId, userId))
            {
                return await AppendAsync(botUin, groupId, groupName, userId, name);
            }

            const string sql = @"
                UPDATE msgcount 
                SET msg_date = CURRENT_TIMESTAMP, c_msg = c_msg + 1 
                WHERE group_id = @groupId AND user_id = @userId AND c_date = CURRENT_DATE";
            using var conn = CreateConnection();
            return await conn.ExecuteAsync(sql, new { groupId, userId });
        }

        public async Task<int> GetMsgCountAsync(long groupId, long userId, bool yesterday = false)
        {
            string dateExpr = yesterday ? "CURRENT_DATE - INTERVAL '1 day'" : "CURRENT_DATE";
            string sql = $"SELECT c_msg FROM msgcount WHERE group_id = @groupId AND user_id = @userId AND c_date = {dateExpr} LIMIT 1";
            using var conn = CreateConnection();
            return await conn.QueryFirstOrDefaultAsync<int>(sql, new { groupId, userId });
        }

        public async Task<int> GetCountOrderAsync(long groupId, long userId, bool yesterday = false)
        {
            string dateExpr = yesterday ? "CURRENT_DATE - INTERVAL '1 day'" : "CURRENT_DATE";
            string sql = $@"
                SELECT COUNT(id) + 1 
                FROM msgcount 
                WHERE group_id = @groupId AND c_date = {dateExpr} 
                AND c_msg > (SELECT c_msg FROM msgcount WHERE group_id = @groupId AND user_id = @userId AND c_date = {dateExpr} LIMIT 1)";
            using var conn = CreateConnection();
            return await conn.ExecuteScalarAsync<int>(sql, new { groupId, userId });
        }

        public async Task<string> GetCountListAsync(long groupId, bool yesterday = false, long top = 10)
        {
            string dateExpr = yesterday ? "CURRENT_DATE - INTERVAL '1 day'" : "CURRENT_DATE";
            string sql = $@"
                SELECT user_id, c_msg 
                FROM msgcount 
                WHERE group_id = @groupId AND c_date = {dateExpr} 
                ORDER BY c_msg DESC 
                LIMIT @top";
            
            using var conn = CreateConnection();
            var results = await conn.QueryAsync<(long UserId, int CMsg)>(sql, new { groupId, top });
            
            if (!results.Any()) return "";

            var list = results.ToList();
            var res = "";
            for (int i = 0; i < list.Count; i++)
            {
                res += $"【第{i + 1}名】 [@:{list[i].UserId}] 发言：{list[i].CMsg}\n";
            }
            return res;
        }
    }
}
