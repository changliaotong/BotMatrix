namespace BotWorker.Application.Services;

public static class MusicService
{
    public static OneBotSharp.Objs.Message.MsgMusic.MsgData ToMusicData(this BotWorker.Modules.Games.SongResult result)
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


