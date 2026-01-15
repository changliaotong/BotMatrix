using System;
using System.Collections.Generic;
using System.Threading.Tasks;
using BotWorker.Infrastructure.Persistence.Repositories;
using BotWorker.Domain.Repositories;
using BotWorker.Domain.Entities;
using Dapper;
using Dapper.Contrib.Extensions;

namespace BotWorker.Infrastructure.Persistence.Repositories
{
    public class UserRepository : BaseRepository<UserInfo>, IUserRepository
    {
        public UserRepository(string? connectionString = null) 
            : base("user_info", connectionString ?? GlobalConfig.BaseInfoConnection)
        {
        }

        public async Task<UserInfo?> GetByOpenIdAsync(string openId, long botUin)
        {
            string sql = $"SELECT * FROM {_tableName} WHERE user_open_id = @openId AND bot_uin = @botUin";
            using var conn = CreateConnection();
            return await conn.QueryFirstOrDefaultAsync<UserInfo>(sql, new { openId, botUin });
        }

        public async Task<UserInfo?> GetBySz84UidAsync(int sz84Uid)
        {
            string sql = $"SELECT * FROM {_tableName} WHERE sz84_uid = @sz84Uid";
            using var conn = CreateConnection();
            return await conn.QueryFirstOrDefaultAsync<UserInfo>(sql, new { sz84Uid });
        }

        public async Task<long> AddAsync(UserInfo user)
        {
            return await InsertAsync(user);
        }

        public async Task<bool> UpdateAsync(UserInfo user)
        {
            user.UpdatedAt = DateTime.Now;
            return await UpdateEntityAsync(user);
        }

        public async Task<long> GetCreditAsync(long botUin, long groupId, long qq, IDbTransaction? trans = null)
        {
            return await GetValueAsync<long>("credit", qq, trans);
        }

        public async Task<bool> AddCreditAsync(long botUin, long groupId, long qq, long amount, string reason, IDbTransaction? trans = null)
        {
            return await IncrementValueAsync("credit", amount, qq, trans) > 0;
        }

        public async Task<long> GetCreditForUpdateAsync(long botUin, long groupId, long qq, IDbTransaction? trans = null)
        {
            string sql = $"SELECT credit FROM {_tableName} WHERE id = @qq FOR UPDATE";
            var conn = trans?.Connection ?? CreateConnection();
            return await conn.ExecuteScalarAsync<long>(sql, new { qq }, trans);
        }

        public async Task<long> GetTokensAsync(long qq)
        {
            return await GetValueAsync<long>("tokens", qq);
        }

        public async Task<long> GetTokensForUpdateAsync(long qq, IDbTransaction? trans = null)
        {
            string sql = $"SELECT tokens FROM {_tableName} WHERE id = @qq FOR UPDATE";
            var conn = trans?.Connection ?? CreateConnection();
            return await conn.ExecuteScalarAsync<long>(sql, new { qq }, trans);
        }

        public async Task<bool> AddTokensAsync(long qq, long amount, IDbTransaction? trans = null)
        {
            return await IncrementValueAsync("tokens", amount, qq, trans) > 0;
        }

        private static ITokensLogRepository TokensLogRepository => 
            BotMessage.ServiceProvider?.GetRequiredService<ITokensLogRepository>() 
            ?? throw new InvalidOperationException("ITokensLogRepository not registered");

        public async Task<long> GetDayTokensGroupAsync(long groupId, long userId)
        {
            return await TokensLogRepository.GetDayTokensGroupAsync(groupId, userId);
        }

        public async Task<long> GetDayTokensAsync(long userId)
        {
            return await TokensLogRepository.GetDayTokensAsync(userId);
        }

        public async Task<string> GetTokensListAsync(long groupId, int top)
        {
            string sql = $@"
                SELECT u.id, u.tokens 
                FROM {_tableName} u
                JOIN group_member gm ON u.id = gm.user_id
                WHERE gm.group_id = @groupId
                ORDER BY u.tokens DESC
                LIMIT @top";
            
            using var conn = CreateConnection();
            var results = await conn.QueryAsync<(long UserId, long Tokens)>(sql, new { groupId, top });
            
            var sb = new System.Text.StringBuilder();
            int i = 1;
            foreach (var res in results)
            {
                sb.Append($"【第{i++}名】 [@:{res.UserId}] 算力：{res.Tokens}\n");
            }
            return sb.ToString();
        }

        public async Task<long> GetTokensRankingAsync(long groupId, long qq)
        {
            string sql = $@"
                SELECT count(*) + 1
                FROM {_tableName} u
                JOIN group_member gm ON u.id = gm.user_id
                WHERE gm.group_id = @groupId
                AND u.tokens > (SELECT tokens FROM {_tableName} WHERE id = @qq)";
            
            using var conn = CreateConnection();
            return await conn.ExecuteScalarAsync<long>(sql, new { groupId, qq });
        }

        public async Task<bool> GetIsBlackAsync(long qq)
        {
            return await GetValueAsync<bool>("is_black", qq);
        }

        public async Task<bool> GetIsFreezeAsync(long qq)
        {
            return await GetValueAsync<bool>("is_freeze", qq);
        }

        public async Task<bool> GetIsShutupAsync(long qq)
        {
            return await GetValueAsync<bool>("is_shutup", qq);
        }

        public async Task<bool> GetIsSuperAsync(long qq)
        {
            return await GetValueAsync<bool>("is_super", qq);
        }

        public async Task<bool> UpdateCszGameAsync(long qq, int cszRes, long cszCredit, int cszTimes)
        {
            string sql = $@"
                UPDATE {_tableName} SET 
                    csz_res = @cszRes, 
                    csz_credit = @cszCredit, 
                    csz_times = @cszTimes,
                    updated_at = CURRENT_TIMESTAMP 
                WHERE id = @qq";
            
            using var conn = CreateConnection();
            return await conn.ExecuteAsync(sql, new { cszRes, cszCredit, cszTimes, qq }) > 0;
        }

        public override async Task<int> SetValueAsync(string field, object value, long id, IDbTransaction? trans = null)
        {
            // 确保字段名安全
            if (!System.Text.RegularExpressions.Regex.IsMatch(field, @"^[a-zA-Z0-9_]+$"))
                throw new ArgumentException("Invalid field name");

            string dbField = ToSnakeCase(field);
            string sql = $"UPDATE {_tableName} SET {dbField} = @value, updated_at = CURRENT_TIMESTAMP WHERE id = @id";
            
            if (trans != null)
            {
                return await trans.Connection.ExecuteAsync(sql, new { value, id }, trans);
            }
            using var conn = CreateConnection();
            return await conn.ExecuteAsync(sql, new { value, id });
        }

        public async Task<long> GetCoinsAsync(long userId)
        {
            string sql = $"SELECT coins FROM {_tableName} WHERE id = @userId";
            using var conn = CreateConnection();
            return await conn.ExecuteScalarAsync<long>(sql, new { userId });
        }

        public async Task<long> GetBotUinByOpenidAsync(string userOpenid)
        {
            string sql = $"SELECT COALESCE(bot_uin, 0) FROM {_tableName} WHERE user_open_id = @userOpenid";
            using var conn = CreateConnection();
            return await conn.ExecuteScalarAsync<long>(sql, new { userOpenid });
        }

        public async Task<long> GetTargetUserIdAsync(string userOpenid)
        {
            string sql = $"SELECT COALESCE(target_user_id, id) FROM {_tableName} WHERE user_open_id = @userOpenid";
            using var conn = CreateConnection();
            return await conn.ExecuteScalarAsync<long>(sql, new { userOpenid });
        }

        public async Task<long> GetMaxIdInRangeAsync(long min, long max)
        {
            string sql = $"SELECT COALESCE(MAX(id), 0) FROM {_tableName} WHERE id > @min AND id < @max";
            using var conn = CreateConnection();
            return await conn.ExecuteScalarAsync<long>(sql, new { min, max });
        }

        public async Task<string> GetUserOpenidAsync(long botUin, long userId)
        {
            string sql = $"SELECT user_open_id FROM {_tableName} WHERE (target_user_id = @userId OR id = @userId) AND bot_uin = @botUin";
            using var conn = CreateConnection();
            return await conn.ExecuteScalarAsync<string>(sql, new { userId, botUin });
        }

        public async Task<decimal> GetBalanceAsync(long qq, IDbTransaction? trans = null)
        {
            return await GetValueAsync<decimal>("balance", qq, trans);
        }

        public async Task<decimal> GetBalanceForUpdateAsync(long qq, IDbTransaction? trans = null)
        {
            string sql = $"SELECT balance FROM {_tableName} WHERE id = @qq FOR UPDATE";
            if (trans != null)
            {
                return await trans.Connection.ExecuteScalarAsync<decimal>(sql, new { qq }, trans);
            }
            using var conn = CreateConnection();
            return await conn.ExecuteScalarAsync<decimal>(sql, new { qq });
        }

        public async Task<bool> AddBalanceAsync(long qq, decimal amount, IDbTransaction? trans = null)
        {
            return await IncrementValueAsync("balance", amount, qq, trans) > 0;
        }

        public async Task<decimal> GetFreezeBalanceAsync(long qq, IDbTransaction? trans = null)
        {
            return await GetValueAsync<decimal>("balance_freeze", qq, trans);
        }

        public async Task<decimal> GetFreezeBalanceForUpdateAsync(long qq, IDbTransaction? trans = null)
        {
            string sql = $"SELECT balance_freeze FROM {_tableName} WHERE id = @qq FOR UPDATE";
            if (trans != null)
            {
                return await trans.Connection.ExecuteScalarAsync<decimal>(sql, new { qq }, trans);
            }
            using var conn = CreateConnection();
            return await conn.ExecuteScalarAsync<decimal>(sql, new { qq });
        }

        public async Task<bool> FreezeBalanceAsync(long qq, decimal amount, IDbTransaction? trans = null)
        {
            string sql = $@"
                UPDATE {_tableName} SET 
                    balance = balance - @amount, 
                    balance_freeze = COALESCE(balance_freeze, 0) + @amount,
                    updated_at = CURRENT_TIMESTAMP 
                WHERE id = @qq";
            
            if (trans != null)
            {
                return await trans.Connection.ExecuteAsync(sql, new { amount, qq }, trans) > 0;
            }
            using var conn = CreateConnection();
            return await conn.ExecuteAsync(sql, new { amount, qq }) > 0;
        }

        public async Task<string> GetBalanceListAsync(long groupId, long qq)
        {
            // 获取前10名
            string sqlTop10 = $@"
                SELECT u.id as UserId, u.balance as Balance
                FROM {_tableName} u
                WHERE u.id IN (SELECT DISTINCT user_id FROM Balance WHERE group_id = @groupId)
                ORDER BY u.balance DESC
                LIMIT 10";

            using var conn = CreateConnection();
            var topList = await conn.QueryAsync<(long UserId, decimal Balance)>(sqlTop10, new { groupId });

            var sb = new System.Text.StringBuilder();
            int i = 1;
            bool foundMe = false;
            foreach (var item in topList)
            {
                sb.Append($"【第{i++}名】 [@:{item.UserId}] 余额：{item.Balance:N}\n");
                if (item.UserId == qq) foundMe = true;
            }

            if (!foundMe)
            {
                // 获取我的排名
                string sqlMyRank = $@"
                    SELECT count(*) + 1
                    FROM {_tableName} u
                    WHERE u.balance > (SELECT balance FROM {_tableName} WHERE id = @qq)
                    AND u.id IN (SELECT DISTINCT user_id FROM Balance WHERE group_id = @groupId)";
                
                var myRank = await conn.ExecuteScalarAsync<long>(sqlMyRank, new { groupId, qq });
                var myBalance = await GetBalanceAsync(qq);
                sb.Append($"【第{myRank}名】 [@:{qq}] 余额：{myBalance:N}");
            }

            return sb.ToString();
        }

        public async Task<string> GetRankAsync(long groupId, long qq)
        {
             string sqlMyRank = $@"
                SELECT count(*) + 1
                FROM {_tableName} u
                WHERE u.balance > (SELECT balance FROM {_tableName} WHERE id = @qq)
                AND u.id IN (SELECT DISTINCT user_id FROM Balance WHERE group_id = @groupId)";
            
            using var conn = CreateConnection();
            var myRank = await conn.ExecuteScalarAsync<long>(sqlMyRank, new { groupId, qq });
            var myBalance = await GetBalanceAsync(qq);
            return $"【第{myRank}名】 [@:{qq}] 余额：{myBalance:N}";
        }

        public async Task SyncCacheFieldAsync(long userId, string field, object value)
        {
            if (MetaData.CacheService == null || !MetaData.UseCache) return;
            
            // 模拟 MetaData 的缓存键生成逻辑
            string fullName = typeof(UserInfo).FullName ?? "BotWorker.Domain.Entities.UserInfo";
            string cacheKey = $"MetaData:{fullName}:Id:{userId}";
            
            // 注意：MetaData 的 SyncCacheField 逻辑是失效行级缓存并设置字段级缓存（如果支持）
            // 这里我们主要处理失效行级缓存，因为 Repository 模式下通常倾向于整体缓存或按需加载
            await MetaData.CacheService.RemoveAsync(cacheKey);
            
            // 如果有字段级缓存需求，可以根据 field 处理
            string fieldCacheKey = $"MetaData:{fullName}:Id:{userId}_{field}";
            await MetaData.CacheService.SetAsync(fieldCacheKey, value, TimeSpan.FromMinutes(1));
        }

        public async Task<int> AppendAsync(long botUin, long groupId, long qq, string name, long ownerId, IDbTransaction? trans = null)
        {
            var user = await GetByIdAsync(qq);
            if (user == null)
            {
                user = new UserInfo
                {
                    Id = qq,
                    Name = name,
                    BotUin = botUin,
                    GroupId = groupId,
                    InsertDate = DateTime.Now,
                    UpdatedAt = DateTime.Now,
                    // Set defaults as needed
                    IsCoins = true, 
                    IsSz84 = false
                };
                await InsertAsync(user, trans);
                return 1;
            }
            else
            {
                // Optionally update name or other fields
                if (user.Name != name)
                {
                    user.Name = name;
                    user.UpdatedAt = DateTime.Now;
                    await UpdateEntityAsync(user, trans);
                }
                return 0;
            }
        }
    }
}
