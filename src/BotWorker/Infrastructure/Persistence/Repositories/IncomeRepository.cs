using System;
using System.Collections.Generic;
using System.Threading.Tasks;
using BotWorker.Domain.Entities;
using BotWorker.Domain.Repositories;
using BotWorker.Infrastructure.Persistence.Repositories;
using BotWorker.Modules.Office;
using Dapper;
using System.Data;

namespace BotWorker.Infrastructure.Persistence.Repositories
{
    public class IncomeRepository : BaseRepository<Income>, IIncomeRepository
    {
        public IncomeRepository(string? connectionString = null)
            : base("income", connectionString ?? GlobalConfig.BaseInfoConnection)
        {
        }

        public async Task<long> AddAsync(Income income, IDbTransaction? trans = null)
        {
            return await InsertAsync(income, trans);
        }

        public async Task<float> GetTotalAsync(long userId)
        {
            string sql = $"SELECT COALESCE(SUM(income_money), 0) FROM {_tableName} WHERE user_id = @userId";
            using var conn = CreateConnection();
            return await conn.ExecuteScalarAsync<float>(sql, new { userId });
        }

        public async Task<float> GetTotalLastYearAsync(long userId)
        {
            // Postgres compatible date logic
            string sql = $@"
                SELECT COALESCE(SUM(income_money), 0) 
                FROM {_tableName} 
                WHERE user_id = @userId 
                AND ABS(EXTRACT(YEAR FROM CURRENT_DATE) - EXTRACT(YEAR FROM income_date)) <= 1";
            using var conn = CreateConnection();
            return await conn.ExecuteScalarAsync<float>(sql, new { userId });
        }

        public async Task<bool> IsVipOnceAsync(long groupId)
        {
            return await CountAsync("WHERE group_id = @groupId", new { groupId }) > 0;
        }

        public async Task<int> GetClientLevelAsync(long userId)
        {
            string sql = $"SELECT get_client_level(COALESCE(SUM(income_money), 0)) FROM {_tableName} WHERE user_id = @userId";
            using var conn = CreateConnection();
            return await conn.ExecuteScalarAsync<int>(sql, new { userId });
        }

        public async Task<string> GetLevelListAsync(long groupId)
        {
            string sql = $@"
                SELECT user_id, COALESCE(SUM(income_money), 0) as SIncome, get_client_level(COALESCE(SUM(income_money), 0)) as client_level 
                FROM {_tableName} 
                WHERE user_id IN (SELECT user_id FROM credit_log WHERE group_id = @groupId) 
                GROUP BY user_id 
                ORDER BY SIncome DESC 
                LIMIT 3";
            
            using var conn = CreateConnection();
            var results = await conn.QueryAsync<(long UserId, decimal SIncome, int client_level)>(sql, new { groupId });
            
            var sb = new System.Text.StringBuilder();
            int i = 1;
            foreach (var res in results)
            {
                sb.Append($"【第{i++}名】：[@:{res.UserId}]   荣誉等级：LV{res.client_level}\n");
            }
            return sb.ToString();
        }

        public async Task<string> GetLeverOrderAsync(long groupId, long userId)
        {
            string sql = $@"
                SELECT COUNT(user_id) + 1 
                FROM (
                    SELECT user_id 
                    FROM {_tableName} 
                    WHERE user_id IN (SELECT user_id FROM credit_log WHERE group_id = @groupId) 
                    GROUP BY user_id 
                    HAVING SUM(income_money) > (SELECT SUM(income_money) FROM {_tableName} WHERE user_id = @userId)
                ) a";
            using var conn = CreateConnection();
            return (await conn.ExecuteScalarAsync<long>(sql, new { groupId, userId })).ToString();
        }

        public async Task<string> GetStatAsync(string range)
        {
            string whereClause = range switch
            {
                "today" => "EXTRACT(DAY FROM CURRENT_TIMESTAMP - income_date) < 1",
                "yesterday" => "EXTRACT(DAY FROM CURRENT_TIMESTAMP - income_date) = 1",
                "month" => "EXTRACT(DAY FROM CURRENT_TIMESTAMP - income_date) <= 30",
                "year" => "EXTRACT(DAY FROM CURRENT_TIMESTAMP - income_date) <= 365",
                "all" => "1=1",
                _ => "1=0"
            };

            string sql = $"SELECT COALESCE(SUM(income_money), 0) FROM {_tableName} WHERE {whereClause}";
            using var conn = CreateConnection();
            var val = await conn.ExecuteScalarAsync<decimal>(sql);
            return val.ToString("C");
        }
    }
}
