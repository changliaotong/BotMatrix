using sz84.Bots.BotMessages;

namespace sz84.Bots.Interfaces;

/// <summary>
/// 跨平台机器人客户端接口
/// </summary>
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
    Task<string> LeaveAsync(long selfId, long group);
    Task<bool> IsInGroupAsync(long selfId, long group, long target);

}