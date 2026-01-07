using BotWorker.Core.Data;

namespace BotWorker.Bots.Entries
{
    public class Song
    {
        public long MusicId { get; set; }

        public MusicKind Kind { get; set; }

        public string? SongId { get; set; }
    }
}
