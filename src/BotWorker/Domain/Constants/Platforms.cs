namespace BotWorker.Domain.Constants;

public static class Platforms
{
    public const string Mirai = "Mirai";
    public const string QQ = "qq";
    public const string Weixin = "weixin";
    public const string Public = "Public";
    public const string WeCom = "WeCom";
    public const string QQGuild = "qqguild";
    public const string Web = "Web";
    public const string Local = "Local";
    public const string Worker = "Worker";

    public static int BotType(string? platform) =>
    platform switch
    {
        Mirai => 0,
        QQ => 1,
        Weixin => 2,
        Public => 3,
        QQGuild => 4,
        Web => 5,
        Local => 6,
        Worker => 7,
        _ => 1
    };

    public static string ToPlatform(int value)
    {
        return value switch
        {
            0 => Mirai,
            1 => QQ,
            2 => Weixin,
            3 => Public,
            4 => QQGuild,
            5 => Web,
            6 => Local,
            7 => Worker,
            _ => QQ   // 默认平台，建议保持和正向一致
        };
    }
}
