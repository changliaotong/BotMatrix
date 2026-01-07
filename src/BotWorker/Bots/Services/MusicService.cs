using OneBotSharp.Objs.Message;
using BotWorker.Bots.Entries;
using BotWorker.Models;

namespace BotWorker.Bots.Services;

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


