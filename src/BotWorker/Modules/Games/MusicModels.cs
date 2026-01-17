using System;
using Dapper.Contrib.Extensions;

namespace BotWorker.Modules.Games
{
    public class SongResult
    {
        public string Name { get; set; } = string.Empty;
        public string Artist { get; set; } = string.Empty;
        public string Cover { get; set; } = string.Empty;
        public string AudioUrl { get; set; } = string.Empty;
        public string Source { get; set; } = "kuwo";

        public BotWorker.Models.MusicShareMessage ToMusicShareMessage()
        {
            return new BotWorker.Models.MusicShareMessage
            {
                Title = Name,
                Summary = Artist,
                PictureUrl = Cover,
                JumpUrl = AudioUrl,
                MusicUrl = AudioUrl,
                Brief = $"[分享]{Name}",
                Kind = "QQMusic"
            };
        }
    }

    [Table("user_song_orders")]
    public class SongOrder
    {
        [ExplicitKey]
        public Guid Id { get; set; } = Guid.NewGuid();
        public string FromUserId { get; set; } = string.Empty;
        public string FromNickname { get; set; } = string.Empty;
        public string ToUserId { get; set; } = string.Empty; // 如果是点给自己，则与 FromUserId 相同
        public string ToNickname { get; set; } = string.Empty;
        
        public string SongName { get; set; } = string.Empty;
        public string Artist { get; set; } = string.Empty;
        public string Message { get; set; } = string.Empty; // 寄语
        
        public DateTime OrderTime { get; set; } = DateTime.Now;
    }
}
