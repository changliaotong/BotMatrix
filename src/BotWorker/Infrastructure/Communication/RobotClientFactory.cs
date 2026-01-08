using BotWorker.Domain.Interfaces;

namespace BotWorker.Infrastructure.Communication;

public static class RobotClientFactory
{
    private static readonly Dictionary<string, IRobotClient> clients = [];

    public static void Register(string platform, IRobotClient client)
    {
        clients[platform] = client;
    }

    public static IRobotClient Get(string platform)
    {
        Logger.Info($"获取平台 {platform} 的 IRobotClient");
        foreach (var c in clients)
        {
            Logger.Info($"{c.Key} => {c.Value}");
        }
        if (clients.TryGetValue(platform, out var client))
            return client;

        throw new NotSupportedException($"平台 {platform} 未注册 IRobotClient");
    }
}
