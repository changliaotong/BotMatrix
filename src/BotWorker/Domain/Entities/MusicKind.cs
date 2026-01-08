using System.ComponentModel;

namespace BotWorker.Domain.Entities;

public enum MusicKind
{
    /// <summary>
    /// QQ音乐
    /// </summary>
    [Description("QQMusic")] QQMusic,

    /// <summary>
    /// 网易云音乐
    /// </summary>
    [Description("NeteaseCloudMusic")] NeteaseCloudMusic,

    /// <summary>
    /// 酷狗音乐
    /// </summary>
    [Description("KugouMusic")] KugouMusic,
    
    /// <summary>
    /// 早喵音乐
    /// </summary>
    [Description("ZMMusic")] ZMMusic
}
