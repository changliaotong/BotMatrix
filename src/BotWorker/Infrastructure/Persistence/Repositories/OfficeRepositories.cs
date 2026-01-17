using System;
using System.Collections.Generic;
using System.Data;
using System.Linq;
using System.Threading.Tasks;
using BotWorker.Domain.Entities;
using BotWorker.Domain.Repositories;
using BotWorker.Infrastructure.Persistence.Database;
using BotWorker.Modules.Office;
using Dapper;

namespace BotWorker.Infrastructure.Persistence.Repositories
{
    public class PartnerRepository : BaseRepository<Partner>, IPartnerRepository
    {
        private readonly IUserRepository _userRepository;
        private readonly IIncomeRepository _incomeRepository;

        public PartnerRepository(IUserRepository userRepository, IIncomeRepository incomeRepository, string? connectionString = null)
            : base("Partner", connectionString ?? GlobalConfig.BaseInfoConnection)
        {
            _userRepository = userRepository;
            _incomeRepository = incomeRepository;
        }

        public async Task<bool> IsPartnerAsync(long userId)
        {
            if (userId == 0) return false;
            const string sql = "SELECT COUNT(1) FROM Partner WHERE UserID = @userId AND IsValid = 1";
            using var conn = CreateConnection();
            return await conn.ExecuteScalarAsync<int>(sql, new { userId }) > 0;
        }

        public async Task<bool> IsNotPartnerAsync(long userId)
        {
            return !await IsPartnerAsync(userId);
        }

        public async Task<string> BecomePartnerAsync(long userId)
        {
            var incomeTotal = await _incomeRepository.GetTotalAsync(userId);

            if (await IsPartnerAsync(userId))
                return "您已经是我们尊贵的合伙人";

            if (incomeTotal < 1000)
                return $"您的总消费金额{incomeTotal}不足1000元";

            int i = await AppendAsync(userId);
            if (i == -1)
                return "服务器繁忙，请稍后再试";

            return "恭喜你已经成为我司尊贵的合伙人。";
        }

        public async Task<int> AppendAsync(long userId, long refUserId = 0)
        {
            const string sql = "INSERT INTO Partner (UserId, refUserId) VALUES (@userId, @refUserId)";
            using var conn = CreateConnection();
            return await conn.ExecuteAsync(sql, new { userId, refUserId });
        }

        public async Task<string> GetCreditTodayAsync(long qq)
        {
            // Note: This SQL is complex and uses some SQL Server specific functions like SqlTop and SqlDateDiff.
            // In a real refactor, these should be converted to PostgreSQL syntax.
            // For now, I'll use a placeholder or try to adapt it.
            
            // Simplified adaptation for PostgreSQL
            string sql = $@"
                SELECT a.UserId, SUM(abs(CreditAdd)) * 6 / 1000 as partner_credit 
                FROM credit_log a 
                INNER JOIN user_info b ON a.UserId = b.UserId 
                WHERE (CURRENT_DATE - a.InsertDate::date) = 0 
                AND b.credit_freeze = 1 
                AND partner_qq = @qq 
                AND a.InsertDate > b.BindDateHome 
                AND (CreditInfo LIKE '%猜大小%' OR CreditInfo LIKE '%三公%' OR CreditInfo LIKE '%抽奖%' OR CreditInfo LIKE '%猜拳%' OR CreditInfo LIKE '%猜数字%')
                GROUP BY a.UserId 
                ORDER BY partner_credit DESC 
                LIMIT 10";

            using var conn = CreateConnection();
            var results = await conn.QueryAsync(sql, new { qq });
            string res = string.Join("\n", results.Select((r, i) => $"{i + 1} {r.UserId} {r.partner_credit}"));

            string sqlTotal = $@"
                SELECT SUM(abs(CreditAdd)) * 6 / 1000 as partner_credit 
                FROM credit_log a 
                INNER JOIN user_info b ON a.UserId = b.UserId 
                WHERE (CURRENT_DATE - a.InsertDate::date) = 0 
                AND b.IsSuper = 1 
                AND partner_qq = @qq 
                AND a.InsertDate > b.BinDate 
                AND (CreditInfo LIKE '%猜大小%' OR CreditInfo LIKE '%三公%' OR CreditInfo LIKE '%抽奖%' OR CreditInfo LIKE '%猜拳%' OR CreditInfo LIKE '%猜数字%')";
            
            var total = await conn.ExecuteScalarAsync<decimal?>(sqlTotal, new { qq });
            res += $"\n今日合计：{total ?? 0}";
            return res;
        }

        public async Task<long> GetUnsettledCreditAsync(long userId)
        {
            const string sqlRes = "SELECT SUM(partner_credit) FROM robot_credit_day WHERE partner_qq = @userId AND is_settle = 0";
            using var conn = CreateConnection();
            return await conn.ExecuteScalarAsync<long?>(sqlRes, new { userId }) ?? 0;
        }

        public async Task<int> SettleAsync(long userId, IDbTransaction? trans = null)
        {
            const string sqlUpdate = "UPDATE robot_credit_day SET settle_date = CURRENT_TIMESTAMP, is_settle = 1, settle_id = @settle_id WHERE is_settle = 0 AND partner_qq = @userId";
            
            if (trans != null)
            {
                return await trans.Connection.ExecuteAsync(sqlUpdate, new { userId, settle_id = userId }, trans);
            }
            else
            {
                using var conn = CreateConnection();
                return await conn.ExecuteAsync(sqlUpdate, new { userId, settle_id = userId });
            }
        }

        public async Task<string> GetCreditListAsync(long qq)
        {
            if (!await IsPartnerAsync(qq))
                return "此功能仅合伙人可用";

            const string sql = @"
                SELECT (EXTRACT(MONTH FROM credit_day) * 100 + EXTRACT(DAY FROM credit_day)) as c_day, 
                       COUNT(UserId) as c_client, 
                       SUM(partner_credit) as partner_credit 
                FROM robot_credit_day 
                WHERE partner_qq = @qq AND is_settle = 0 
                GROUP BY credit_day 
                ORDER BY credit_day DESC 
                LIMIT 7";

            using var conn = CreateConnection();
            var list = await conn.QueryAsync(sql, new { qq });
            string res = string.Join("\n", list.Select(r => $"{r.c_day} {r.c_client}人 {r.partner_credit}分"));

            const string sqlTotal = @"
                SELECT COUNT(DISTINCT UserId) as c_client, SUM(partner_credit) as partner_credit 
                FROM robot_credit_day 
                WHERE partner_qq = @qq AND is_settle = 0";
            
            var total = await conn.QueryFirstOrDefaultAsync(sqlTotal, new { qq });
            if (total != null)
            {
                res += $"\n合计 {total.c_client}人 {total.partner_credit}分";
            }
            return res;
        }
    }

    public class PriceRepository : BaseRepository<Price>, IPriceRepository
    {
        public PriceRepository(string? connectionString = null)
            : base("Price", connectionString ?? GlobalConfig.BaseInfoConnection)
        {
        }

        public async Task<decimal> GetRobotPriceAsync(long month)
        {
            if (month > 60) month = 60;
            const string sql = "SELECT price FROM Price WHERE month = @month";
            using var conn = CreateConnection();
            return await conn.ExecuteScalarAsync<decimal>(sql, new { month });
        }
    }
}
