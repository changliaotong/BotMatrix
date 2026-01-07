using OneBotSharp.Objs.Message;
using sz84.Bots.Entries;
using BotWorker.Models;

namespace sz84.Bots.Services;

public static class MusicService
{
    public static OneBotSharp.Objs.Message.MsgMusic.MsgData ToMusicData(this SongResult result)
    {
        return new OneBotSharp.Objs.Message.MsgMusic.MsgData
        {
            Type = "qq",
            Title = result.Name,
            Content = result.Artist,
            Url = result.AudioUrl,
            Image = result.Cover,
            Audio = result.AudioUrl
        };
    }
}


