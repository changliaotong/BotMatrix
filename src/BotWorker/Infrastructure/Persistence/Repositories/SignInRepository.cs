using System;
using System.Threading.Tasks;
using BotWorker.Domain.Entities;
using BotWorker.Domain.Repositories;
using Dapper;

namespace BotWorker.Infrastructure.Persistence.Repositories
{
    public class SignInRepository : BaseRepository<GroupSignIn>, ISignInRepository
    {
        public SignInRepository(string? connectionString = null) 
            : base("group_signin", connectionString ?? GlobalConfig.BaseInfoConnection)
        {
        }

        public async Task<int> AddSignInAsync(long botUin, long groupId, long qq, string info, IDbTransaction? trans = null)
        {
            const string sql = @"
                INSERT INTO group_signin (
                    robot_qq, weibo_qq, weibo_info, weibo_type, group_id, insert_date
                ) VALUES (
                    @botUin, @qq, @info, 1, @groupId, @now
                ) RETURNING weibo_id";
            
            var conn = trans?.Connection ?? CreateConnection();
            return await conn.ExecuteScalarAsync<int>(sql, new { 
                botUin, qq, info, groupId, now = DateTime.Now 
            }, trans);
        }

        public async Task<long> GetTodaySignCountAsync(long groupId)
        {
            const string sql = @"
                SELECT COUNT(*) FROM group_signin 
                WHERE weibo_type = 1 AND group_id = @groupId 
                AND insert_date::date = CURRENT_DATE";
            
            using var conn = CreateConnection();
            return await conn.ExecuteScalarAsync<long>(sql, new { groupId });
        }

        public async Task<long> GetYesterdaySignCountAsync(long groupId)
        {
            const string sql = @"
                SELECT COUNT(*) FROM group_signin 
                WHERE weibo_type = 1 AND group_id = @groupId 
                AND insert_date::date = CURRENT_DATE - INTERVAL '1 day'";
            
            using var conn = CreateConnection();
            return await conn.ExecuteScalarAsync<long>(sql, new { groupId });
        }

        public async Task<long> GetUserMonthSignCountAsync(long groupId, long qq)
        {
            const string sql = @"
                SELECT COUNT(*) FROM group_signin 
                WHERE weibo_type = 1 AND group_id = @groupId AND weibo_qq = @qq 
                AND date_trunc('month', insert_date) = date_trunc('month', CURRENT_DATE)";
            
            using var conn = CreateConnection();
            return await conn.ExecuteScalarAsync<long>(sql, new { groupId, qq });
        }
    }
}
