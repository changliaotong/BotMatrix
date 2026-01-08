using BotWorker.Bots.Interfaces;

namespace BotWorker.Bots.Platform;

public static class RobotClientFactory
{
    private static readonly Dictionary<string, IRobotClient> clients = [];

    public static void Register(string platform, IRobotClient client)
    {
        clients[platform] = client;
    }

    public static IRobotClient Get(string platform)
    {
        ShowMessage($"��ȡƽ̨ {platform} �� IRobotClient");
        foreach (var c in clients)
        {
            ShowMessage($"{c.Key} => {c.Value}");
        }
        if (clients.TryGetValue(platform, out var client))
            return client;

        throw new NotSupportedException($"ƽ̨ {platform} δע�� IRobotClient");
    }
}

