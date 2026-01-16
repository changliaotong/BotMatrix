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
    public class GroupPropsRepository : BaseRepository<GroupProps>, IGroupPropsRepository
    {
        public GroupPropsRepository() : base("props")
        {
        }

        public async Task<long> GetIdAsync(long groupId, long userId, long propId)
        {
            const string sql = "SELECT id FROM props WHERE group_id = @groupId AND user_id = @userId AND prop_id = @propId AND is_used = 0 LIMIT 1";
            using var conn = CreateConnection();
            return await conn.QueryFirstOrDefaultAsync<long>(sql, new { groupId, userId, propId });
        }

        public async Task<bool> HavePropAsync(long groupId, long userId, long propId)
        {
            return await GetIdAsync(groupId, userId, propId) != 0;
        }

        public async Task<int> UsePropAsync(long groupId, long userId, long propId, long qqProp)
        {
            long id = await GetIdAsync(groupId, userId, propId);
            if (id == 0) return 0;

            const string sql = "UPDATE props SET used_date = CURRENT_TIMESTAMP, used_user_id = @qqProp, is_used = 1 WHERE id = @id";
            using var conn = CreateConnection();
            return await conn.ExecuteAsync(sql, new { qqProp, id });
        }

        public async Task<string> GetMyPropListAsync(long groupId, long userId)
        {
            const string sql = @"
                SELECT b.prop_name, b.prop_price 
                FROM props a 
                INNER JOIN prop b ON a.prop_id = b.id 
                WHERE a.group_id = @groupId AND a.user_id = @userId 
                LIMIT 10";
            
            using var conn = CreateConnection();
            var results = await conn.QueryAsync<(string PropName, int PropPrice)>(sql, new { groupId, userId });
            
            if (!results.Any()) return "您没有任何道具";
            
            return string.Join("\n", results.Select(r => $"{r.PropName} 价格：{r.PropPrice}分"));
        }

        public async Task<int> InsertAsync(long groupId, long userId, long propId, IDbTransaction? trans = null)
        {
            const string sql = "INSERT INTO props (group_id, user_id, prop_id) VALUES (@groupId, @userId, @propId)";
            if (trans != null)
            {
                return await trans.Connection.ExecuteAsync(sql, new { groupId, userId, propId }, trans);
            }
            using var conn = CreateConnection();
            return await conn.ExecuteAsync(sql, new { groupId, userId, propId });
        }
    }

    public class PropRepository : BaseRepository<Prop>, IPropRepository
    {
        public PropRepository() : base("prop")
        {
        }

        public async Task<long> GetIdAsync(string propName)
        {
            const string sql = "SELECT id FROM prop WHERE prop_name = @propName LIMIT 1";
            using var conn = CreateConnection();
            return await conn.QueryFirstOrDefaultAsync<long>(sql, new { propName });
        }

        public async Task<string> GetPropListAsync()
        {
            const string sql = "SELECT prop_name, prop_price FROM prop WHERE is_valid = 1 ORDER BY prop_name LIMIT 10";
            using var conn = CreateConnection();
            var results = await conn.QueryAsync<(string PropName, int PropPrice)>(sql);
            
            if (!results.Any()) return "暂无可购道具";
            
            return string.Join("\n", results.Select(r => $"{r.PropName} 价格：{r.PropPrice}分"));
        }

        public async Task<int> GetPropPriceAsync(long propId)
        {
            const string sql = "SELECT prop_price FROM prop WHERE id = @propId LIMIT 1";
            using var conn = CreateConnection();
            return await conn.QueryFirstOrDefaultAsync<int>(sql, new { propId });
        }
    }
}
