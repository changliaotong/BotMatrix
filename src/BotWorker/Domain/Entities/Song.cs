using sz84.Core.Data;

namespace BotWorker.Domain.Entities
{
    public class Song
    {
        public long MusicId { get; set; }

        public MusicKind Kind { get; set; }

        public string? SongId { get; set; }
    }
}
