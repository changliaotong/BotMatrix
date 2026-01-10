using System.ComponentModel;

namespace BotWorker.Domain.Constants;

public enum MusicKind
{
    /// <summary>
    /// QQ����
    /// </summary>
    [Description("QQMusic")] QQMusic,

    /// <summary>
    /// ����������
    /// </summary>
    [Description("NeteaseCloudMusic")] NeteaseCloudMusic,

    /// <summary>
    /// �ṷ����
    /// </summary>
    [Description("KugouMusic")] KugouMusic,
    
    /// <summary>
    /// ��������
    /// </summary>
    [Description("ZMMusic")] ZMMusic
}

