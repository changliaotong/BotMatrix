using sz84.Core.Data;

namespace sz84.Bots.Entries
{
    public class Song
    {
        public long MusicId { get; set; }

        public MusicKind Kind { get; set; }

        public string? SongId { get; set; }
    }
}
