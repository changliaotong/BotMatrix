namespace BotWorker.Domain.Interfaces;

/// <summary>
/// 跨平台机器人客户端接�?/// </summary>
public interface IRobotClient
{
    Task SendMessageAsync(BotMessage ctx);
    Task<string> MuteAsync(long selfId, long group, long target, int seconds);
    Task<string> KickAsync(long selfId, long group, long target, bool isReject = false);
    Task<string> RecallAsync(long selfId, long group, string message);
    Task<string> RecallForwardAsync(long selfId, long group, string message, string forward);
    Task<string> ChangeNameAsync(long selfId, long group, long target, string newName, string prefixBoy, string prefixGirl, string prefixAdmin);
    Task<string> ChangeNameAllAsync(long selfId, long group, string prefixBoy, string prefixGirl, string prefixAdmin);
    Task<string> SetTitleAsync(long selfId, long group, long target, string title);
    Task<string> SetGroupCardAsync(long selfId, long group, long target, string card);
    Task<string> SetGroupAdminAsync(long selfId, long group, long target, bool admin);
    Task<string> SetGroupWholeMuteAsync(long selfId, long group, bool mute);
    Task<string> LeaveAsync(long selfId, long group);
    Task<bool> IsInGroupAsync(long selfId, long group, long target);
}


