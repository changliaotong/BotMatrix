using System;
using System.Collections.Generic;
using System.Threading.Tasks;
using BotWorker.Domain.Entities;
using BotWorker.Domain.Repositories;
using Dapper;

namespace BotWorker.Infrastructure.Persistence.Repositories
{
    public class PublicUserRepository : BaseRepository<PublicUser>, IPublicUserRepository
    {
        public PublicUserRepository(string? connectionString = null)
            : base("public_user", connectionString ?? GlobalConfig.BaseInfoConnection)
        {
        }

        public async Task<long> GetUserIdAsync(string botKey, string userKey)
        {
            string sql = $"SELECT user_id FROM {_tableName} WHERE bot_key = @botKey AND user_key = @userKey";
            using var conn = CreateConnection();
            var userId = await conn.ExecuteScalarAsync<long?>(sql, new { botKey, userKey });
            
            if (userId.HasValue && userId.Value > 0)
                return userId.Value;

            // If not exists, append (simplified version of old Append)
            var newUser = new PublicUser
            {
                BotKey = botKey,
                UserKey = userKey,
                InsertDate = DateTime.Now,
                IsBind = false,
                BindToken = GetBindToken(botKey, userKey)
            };
            
            // Note: The old Append used a complex ID generation for UserId if it was new.
            // For now, we'll just insert and return. 
            // We might need to implement the full logic if UserId is required to be unique/specific.
            return await InsertAsync(newUser);
        }

        public async Task<bool> IsBindAsync(long userId)
        {
            // const long startUserId = 4104967295;
            return userId < 4104967295L || userId > 90000000000L;
        }

        public async Task<bool> IsSubscribedToOfficialPublicAsync(long userId)
        {
            if (!await IsBindAsync(userId)) return false;
            
            string[] officialKeys = { "gh_c9dfbd45d42f", "gh_2158fa6520a3", "gh_5696f9a0fae9", "gh_f184bf294a46" };
            string sql = $"SELECT COUNT(1) FROM {_tableName} WHERE bot_key = ANY(@officialKeys) AND user_id = @userId";
            using var conn = CreateConnection();
            return await conn.ExecuteScalarAsync<int>(sql, new { officialKeys, userId }) > 0;
        }

        public string GetBindToken(string botKey, string userKey)
        {
            return (botKey + userKey).MD5()[7..23];
        }

        public async Task<string> GetBindTokenAsync(string botKey, string userKey)
        {
            return GetBindToken(botKey, userKey);
        }

        public async Task<string> GetInviteCodeAsync(string botKey, string userKey)
        {
            return await GetValueAsync<string>("BindToken", $"WHERE bot_key = @botKey AND user_key = @userKey", new { botKey, userKey });
        }

        // These complex methods will be moved to BotMessage for better dependency access
        public Task<string> GetRecResAsync(long botUin, long groupId, string groupName, long userId, string name, string botKey, string clientKey, string message)
        {
            throw new NotImplementedException("Logic moved to BotMessage");
        }

        public Task<string> GetBindTokenResAsync(BotMessage bm, string tokenType, string bindToken)
        {
            throw new NotImplementedException("Logic moved to BotMessage");
        }
    }
}
