namespace sz84.Bots.Platform;

public static class Platforms
{
    public const string NapCat = "NapCat";
    public const string Weixin = "weixin";
    public const string Public = "Public";
    public const string WeCom = "WeCom";
    public const string QQGuild = "qqguild";
    public const string Web = "Web";
    public const string Local = "Local";
    public const string Worker = "Worker";
    public const string Mirai = "Mirai";

    public static int BotType(string? platform) =>
    platform switch
    {
        NapCat => 1,
        Weixin => 2,
        Public => 3,
        QQGuild => 4,
        Web => 5,
        Local => 6,
        Worker => 7,
        Mirai => 8,
        _ => 1
    };

    public static string ToPlatform(int value)
    {
        return value switch
        {
            1 => NapCat,
            2 => Weixin,
            3 => Public,
            4 => QQGuild,
            5 => Web,
            6 => Local,
            7 => Worker,
            8 => Mirai,
            _ => NapCat   // 默认平台，建议保持和正向一致
        };
    }
}

