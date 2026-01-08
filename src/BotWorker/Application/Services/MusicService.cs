using OneBotSharp.Objs.Message;
using BotWorker.Domain.Entities;
using BotWorker.Domain.Models; // Assume I will move it to Domain/Models

namespace BotWorker.Application.Services;

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


