using System;
using System.Collections.Generic;
using System.Threading.Tasks;
using BotWorker.Domain.Entities;
using BotWorker.Domain.Repositories;
using Dapper;

namespace BotWorker.Infrastructure.Persistence.Repositories
{
    public class BotPublicRepository : BaseRepository<BotPublic>, IBotPublicRepository
    {
        public BotPublicRepository(string? connectionString = null)
            : base("Public", connectionString ?? GlobalConfig.BaseInfoConnection)
        {
        }

        public async Task<BotPublic?> GetByPublicKeyAsync(string publicKey)
        {
            string sql = $"SELECT * FROM {_tableName} WHERE PublicKey = @publicKey";
            using var conn = CreateConnection();
            return await conn.QueryFirstOrDefaultAsync<BotPublic>(sql, new { publicKey });
        }

        public async Task<long> GetRobotQQAsync(string botKey)
        {
            return await GetValueAsync<long>("BotUin", botKey, "PublicKey");
        }

        public async Task<long> GetGroupIdAsync(string botKey)
        {
            return await GetValueAsync<long>("GroupId", botKey, "PublicKey");
        }

        public async Task<string> GetBotNameAsync(string botKey)
        {
            var res = await GetValueAsync<string>("PublicName", botKey, "PublicKey");
            return res ?? "[未知公众号]";
        }
    }
}
