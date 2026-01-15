using System.Data;
using System.Threading.Tasks;
using BotWorker.Domain.Entities;
using Dapper;
using Dapper.Contrib.Extensions;
using BotWorker.Common;
using BotWorker.Models;

namespace BotWorker.Infrastructure.Persistence.Repositories
{
    public class MusicRepository : BaseRepository<Music>, IMusicRepository
    {
        public MusicRepository(string? connectionString = null) : base("Music", connectionString)
        {
        }

        public async Task<Music?> GetByKindAndSongIdAsync(string kind, string songId)
        {
            using var conn = CreateConnection();
            string sql = $"SELECT * FROM {_tableName} WHERE \"Kind\" = @kind AND \"SongId\" = @songId ORDER BY \"Id\" DESC";
            return await conn.QueryFirstOrDefaultAsync<Music>(sql, new { kind, songId });
        }

        public async Task<string> GetMusicUrlAsync(string kind, string songId)
        {
            using var conn = CreateConnection();
            string sql = $"SELECT TOP 1 * FROM {_tableName} WHERE \"Kind\" = @kind AND \"SongId\" = @songId";
            var music = await conn.QueryFirstOrDefaultAsync<Music>(sql, new { kind, songId });
            
            if (music == null) return "";
            
            // Format: <a href="{4}" target="_blank">{2}:{0}<br /><img width="75" src="{3}"></a>
            // 0: Title, 2: Summary, 3: PictureUrl, 4: MusicUrl
            return $"<a href=\"{music.MusicUrl}\" target=\"_blank\">{music.Summary}:{music.Title}<br /><img width=\"75\" src=\"{music.PictureUrl}\"></a>";
        }

        public async Task<string> GetMusicUrlPublicAsync(string kind, string songId)
        {
            using var conn = CreateConnection();
            string sql = $"SELECT TOP 1 * FROM {_tableName} WHERE \"Kind\" = @kind AND \"SongId\" = @songId";
            var music = await conn.QueryFirstOrDefaultAsync<Music>(sql, new { kind, songId });

            if (music == null) return "";

            // Format: ✅ 点歌成功!\n<a href="{4}">{2}:{0}\n点击此链接开始播放</a>
            return $"✅ 点歌成功!\n<a href=\"{music.MusicUrl}\">{music.Summary}:{music.Title}\n点击此链接开始播放</a>";
        }

        public async Task<MusicShareMessage?> GetMusicShareMessageAsync(long id)
        {
            using var conn = CreateConnection();
            var music = await conn.GetAsync<Music>(id);
            if (music == null) return null;

            return new MusicShareMessage
            {
                Title = music.Title,
                Brief = music.Brief,
                Summary = music.Summary,
                PictureUrl = music.PictureUrl,
                MusicUrl = music.IsVIP ? music.MusicUrl2 : music.MusicUrl,
                JumpUrl = music.JumpUrl,
                Kind = music.Kind
            };
        }

        public async Task<string> GetPayloadAsync(long id)
        {
            using var conn = CreateConnection();
            string sql = $"SELECT COALESCE(\"Payload\", \"JumpUrl\") FROM {_tableName} WHERE \"Id\" = @id";
            return await conn.ExecuteScalarAsync<string>(sql, new { id }) ?? "";
        }

        public async Task<long> GetMusicIdAsync(string kind, string songId)
        {
            using var conn = CreateConnection();
            string sql = $"SELECT \"Id\" FROM {_tableName} WHERE \"Kind\" = @kind AND \"SongId\" = @songId ORDER BY \"Id\" DESC"; // Sort by Id desc to match 'GetWhere' behavior?
            // Original code: GetWhere<long>("Id", $"Kind = ... and SongId=...", "") -> default order? usually by PK.
            // But GetWhere usually returns first match.
            return await conn.ExecuteScalarAsync<long>(sql, new { kind, songId });
        }

        public async Task<string> GetMusicUrlByJumpUrlAsync(string jumpUrl)
        {
            using var conn = CreateConnection();
            string sql = $"SELECT \"MusicUrl\" FROM {_tableName} WHERE \"JumpUrl\" = @jumpUrl ORDER BY \"Id\" DESC";
            return await conn.ExecuteScalarAsync<string>(sql, new { jumpUrl }) ?? "";
        }
    }
}
