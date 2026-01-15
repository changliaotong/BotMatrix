using System.Data;
using System.Threading.Tasks;
using BotWorker.Domain.Entities;
using Dapper;
using Dapper.Contrib.Extensions;
using BotWorker.Common;

namespace BotWorker.Infrastructure.Persistence.Repositories
{
    public class MusicVideoRepository : BaseRepository<MusicVideo>, IMusicVideoRepository
    {
        public MusicVideoRepository(string? connectionString = null) : base("MusicVideo", connectionString)
        {
        }

        public async Task<string> GetContentByVidAsync(string vid)
        {
            using var conn = CreateConnection();
            string sql = "SELECT \"MvContent\" FROM \"MusicVideo\" WHERE \"MvVid\" = @vid";
            return await conn.QueryFirstOrDefaultAsync<string>(sql, new { vid }) ?? string.Empty;
        }

        public async Task<bool> ExistsByVidAsync(string vid)
        {
            using var conn = CreateConnection();
            string sql = "SELECT COUNT(1) FROM \"MusicVideo\" WHERE \"MvVid\" = @vid";
            return await conn.ExecuteScalarAsync<int>(sql, new { vid }) > 0;
        }

        public async Task<int> AddAsync(MusicVideo musicVideo)
        {
            using var conn = CreateConnection();
            return await conn.InsertAsync(musicVideo);
        }
    }
}
