using sz84.Bots.Interfaces;

namespace sz84.Bots.Platform;

public static class RobotClientFactory
{
    private static readonly Dictionary<string, IRobotClient> clients = [];

    public static void Register(string platform, IRobotClient client)
    {
        clients[platform] = client;
    }

    public static IRobotClient Get(string platform)
    {
        ShowMessage($"获取平台 {platform} 的 IRobotClient");
        foreach (var c in clients)
        {
            ShowMessage($"{c.Key} => {c.Value}");
        }
        if (clients.TryGetValue(platform, out var client))
            return client;

        throw new NotSupportedException($"平台 {platform} 未注册 IRobotClient");
    }
}