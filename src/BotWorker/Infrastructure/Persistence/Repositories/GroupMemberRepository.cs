using System;
using System.Collections.Generic;
using System.Threading.Tasks;
using BotWorker.Infrastructure.Persistence.Repositories;
using BotWorker.Domain.Repositories;
using BotWorker.Domain.Entities;
using Dapper;

namespace BotWorker.Infrastructure.Persistence.Repositories
{
    public class GroupMemberRepository : BaseRepository<GroupMember>, IGroupMemberRepository
    {
        public GroupMemberRepository(string? connectionString = null) 
            : base("group_member", connectionString ?? GlobalConfig.BaseInfoConnection)
        {
        }

        public async Task<GroupMember?> GetAsync(long groupId, long userId, IDbTransaction? trans = null)
        {
            return await GetFirstOrDefaultAsync("WHERE group_id = @groupId AND user_id = @userId", new { groupId, userId }, trans);
        }

        public async Task<bool> ExistsAsync(long groupId, long userId, IDbTransaction? trans = null)
        {
            return await CountAsync("WHERE group_id = @groupId AND user_id = @userId", new { groupId, userId }, trans) > 0;
        }

        public async Task<long> AddAsync(GroupMember member)
        {
            member.UpdatedAt = DateTime.Now;
            return await InsertAsync(member);
        }

        public async Task<long> GetCoinsAsync(int coinsType, long groupId, long userId, IDbTransaction? trans = null)
        {
            string field = GetCoinsField(coinsType);
            return await GetValueAsync<long>(field, "WHERE group_id = @groupId AND user_id = @userId", new { groupId, userId }, trans);
        }

        public async Task<bool> UpdateAsync(GroupMember member)
        {
            if (member.Id == 0)
            {
                 var existing = await GetAsync(member.GroupId, member.UserId);
                 if (existing != null) member.Id = existing.Id;
            }
            
            member.UpdatedAt = DateTime.Now;
            return await UpdateEntityAsync(member);
        }

        public async Task<bool> AddCoinsAsync(long botUin, long groupId, long userId, int coinsType, long amount, string reason)
        {
            string field = GetCoinsField(coinsType);
            return await IncrementValueAsync(field, amount, "WHERE group_id = @groupId AND user_id = @userId", new { groupId, userId }) > 0;
        }

        public async Task<long> GetCoinsForUpdateAsync(int coinsType, long groupId, long userId, IDbTransaction trans)
        {
            string field = GetCoinsField(coinsType);
            return await GetValueAsync<long>(field, "WHERE group_id = @groupId AND user_id = @userId FOR UPDATE", new { groupId, userId }, trans);
        }

        public async Task<long> GetLongAsync(string field, long groupId, long userId, IDbTransaction? trans = null)
        {
            return await GetValueAsync<long>(field, "WHERE group_id = @groupId AND user_id = @userId", new { groupId, userId }, trans);
        }

        public async Task<T> GetValueAsync<T>(string field, long groupId, long userId, IDbTransaction? trans = null)
        {
            return await GetValueAsync<T>(field, "WHERE group_id = @groupId AND user_id = @userId", new { groupId, userId }, trans);
        }

        public async Task<int> SetValueAsync(string field, object value, long groupId, long userId, IDbTransaction? trans = null)
        {
            return await SetValueAsync(field, value, "WHERE group_id = @groupId AND user_id = @userId", new { groupId, userId }, trans);
        }

        public async Task<int> IncrementValueAsync(string field, object value, long groupId, long userId, IDbTransaction? trans = null)
        {
            return await IncrementValueAsync(field, value, "WHERE group_id = @groupId AND user_id = @userId", new { groupId, userId }, trans);
        }

        public async Task<long> GetForUpdateAsync(string field, long groupId, long userId, IDbTransaction trans)
        {
            return await GetValueAsync<long>(field, "WHERE group_id = @groupId AND user_id = @userId FOR UPDATE", new { groupId, userId }, trans);
        }

        public async Task<bool> UpdateSignInfoAsync(long groupId, long userId, int signTimes, int signLevel, IDbTransaction? trans = null)
        {
            var member = await GetAsync(groupId, userId, trans);
            if (member == null) return false;

            member.SignTimes = signTimes;
            member.SignLevel = signLevel;
            member.SignDate = DateTime.Now;
            member.SignTimesAll++;
            
            return await UpdateEntityAsync(member, trans);
        }

        public async Task<int> GetSignDateDiffAsync(long groupId, long userId)
        {
            var member = await GetAsync(groupId, userId);
            if (member == null) return 99999; 

            var signDate = member.SignDate == default ? new DateTime(2000, 1, 1) : member.SignDate;
            return (DateTime.Now.Date - signDate.Date).Days;
        }

        public async Task<string> GetSignListAsync(long groupId, int topN = 10)
        {
            var results = await GetListAsync($"WHERE group_id = @groupId AND sign_times > 0 ORDER BY sign_times DESC, sign_level DESC LIMIT {topN}", new { groupId });
            
            var sb = new System.Text.StringBuilder();
            int i = 1;
            foreach (var item in results)
            {
                sb.Append($"【第{i}名】 [@:{item.UserId}] 连续签到：{item.SignTimes}天(LV{item.SignLevel})\n");
                i++;
            }
            return sb.ToString();
        }

        public async Task<int> AppendAsync(long groupId, long userId, string name, string displayName = "", long groupCredit = 0, string confirmCode = "", IDbTransaction? trans = null)
        {
            var member = await GetAsync(groupId, userId, trans);

            if (member != null)
            {
                member.UserName = name;
                member.DisplayName = displayName;
                member.ConfirmCode = confirmCode;
                member.Status = 1;
                member.UpdatedAt = DateTime.Now;
                return await UpdateEntityAsync(member, trans) ? 1 : 0;
            }
            else
            {
                member = new GroupMember
                {
                    GroupId = groupId,
                    UserId = userId,
                    UserName = name,
                    DisplayName = displayName,
                    GroupCredit = groupCredit,
                    ConfirmCode = confirmCode,
                    Status = 1,
                    UpdatedAt = DateTime.Now
                };
                return (int)await InsertAsync(member, trans);
            }
        }

        public async Task<long> GetCoinsRankingAsync(long groupId, long userId)
        {
            string conditions = $@"WHERE group_id = @groupId AND gold_coins > (
                    SELECT gold_coins FROM {_tableName} WHERE group_id = @groupId AND user_id = @userId
                )";
            return await GetValueAsync<long>("COUNT(1) + 1", conditions, new { groupId, userId });
        }

        public async Task<long> GetCoinsRankingAllAsync(long userId)
        {
            string sql = $@"
                SELECT COUNT(1) + 1 
                FROM (
                    SELECT user_id, SUM(gold_coins) as total_gold 
                    FROM {_tableName} 
                    GROUP BY user_id
                ) t
                WHERE total_gold > (
                    SELECT SUM(gold_coins) FROM {_tableName} WHERE user_id = @userId
                )";
            return await ExecuteScalarAsync<long>(sql, new { userId });
        }

        public async Task<int> GetIntAsync(string field, long groupId, long userId, IDbTransaction? trans = null)
        {
            return await GetValueAsync<int>(field, "WHERE group_id = @groupId AND user_id = @userId", new { groupId, userId }, trans);
        }

        public async Task SyncCacheFieldAsync(long groupId, long userId, string field, object value)
        {
            if (BotWorker.Infrastructure.Persistence.ORM.MetaData.CacheService == null || !BotWorker.Infrastructure.Persistence.ORM.MetaData.UseCache) return;
            
            // 行级缓存 Key
            string fullName = typeof(GroupMember).FullName ?? "BotWorker.Domain.Entities.GroupMember";
            string cacheKey = $"MetaData:{fullName}:Id:{userId}_{groupId}";
            await BotWorker.Infrastructure.Persistence.ORM.MetaData.CacheService.RemoveAsync(cacheKey);
            
            // 列级缓存 Key
            string fieldCacheKey = $"MetaData:{fullName}:Id:{userId}_{groupId}_{field}";
            await BotWorker.Infrastructure.Persistence.ORM.MetaData.CacheService.SetAsync(fieldCacheKey, value, TimeSpan.FromMinutes(1));
        }

        private string GetCoinsField(int coinsType)
        {
            return coinsType switch
            {
                0 => "group_credit",
                1 => "gold_coins",
                2 => "black_coins",
                3 => "purple_coins",
                4 => "game_coins",
                _ => throw new ArgumentException("Invalid coins type")
            };
        }
    }
}
