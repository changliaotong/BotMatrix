using System.Threading.Tasks;
using BotWorker.Domain.Entities;
using BotWorker.Domain.Repositories;
using Dapper;

namespace BotWorker.Infrastructure.Persistence.Repositories
{
    public class SystemSettingRepository : BaseRepository<SystemSetting>, ISystemSettingRepository
    {
        public SystemSettingRepository(string? connectionString = null) 
            : base("SystemSetting", connectionString ?? GlobalConfig.DbConnection)
        {
        }

        protected override string KeyField => "Key";

        public async Task<string> GetValueAsync(string key)
        {
            string sql = $"SELECT \"Value\" FROM {_tableName} WHERE \"Key\" = @key";
            using var conn = CreateConnection();
            return await conn.QueryFirstOrDefaultAsync<string>(sql, new { key }) ?? string.Empty;
        }

        public async Task<bool> GetBoolAsync(string key)
        {
            string val = await GetValueAsync(key);
            return val?.ToLower() == "true" || val == "1";
        }
    }
}
