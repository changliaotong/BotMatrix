using System.Threading.Tasks;
using BotWorker.Domain.Entities;

namespace BotWorker.Domain.Repositories
{
    public interface IMusicRepository : IBaseRepository<Music>
    {
        Task<Music?> GetByKindAndSongIdAsync(string kind, string songId);
        Task<string> GetMusicUrlAsync(string kind, string songId);
        Task<string> GetMusicUrlPublicAsync(string kind, string songId);
        Task<BotWorker.Models.MusicShareMessage?> GetMusicShareMessageAsync(long id);
        Task<string> GetPayloadAsync(long id);
        Task<long> GetMusicIdAsync(string kind, string songId);
        Task<string> GetMusicUrlByJumpUrlAsync(string jumpUrl);
    }

    public interface ISongOrderRepository : IBaseRepository<BotWorker.Modules.Games.SongOrder>
    {
        Task<List<BotWorker.Modules.Games.SongOrder>> GetHistoryAsync(string userId);
    }
}
