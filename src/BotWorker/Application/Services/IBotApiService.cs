using System.Threading.Tasks;

namespace BotWorker.Application.Services
{
    /// <summary>
    /// 机器人协议 API 服务：负责与 OneBot 等底层协议交互
    /// </summary>
    public interface IBotApiService
    {
        Task KickMemberAsync(long selfId, long groupId, long userId, bool rejectAddRequest = false);
        Task MuteMemberAsync(long selfId, long groupId, long userId, int durationSeconds);
        Task UnmuteMemberAsync(long selfId, long groupId, long userId);
        Task SetGroupTitleAsync(long selfId, long groupId, long userId, string title);
        Task SendMessageAsync(long selfId, long groupId, long userId, string message, bool isGroup = true);
        Task SetGroupNameAsync(long selfId, long groupId, string groupName);
        Task LeaveGroupAsync(long selfId, long groupId);
        Task SetGroupSpecialTitleAsync(long selfId, long groupId, long userId, string title);
        // 其他 OneBot 标准 API...
    }

    public class BotApiService : IBotApiService
    {
        // 这里的实现将包含实际的 HTTP/WebSocket 发送逻辑
        // 目前为了重构，我们可以先通过 BotMessage 的实例方法代理，后续再彻底重构
        public async Task KickMemberAsync(long selfId, long groupId, long userId, bool rejectAddRequest = false)
        {
            // 实际调用底层协议
            await Task.CompletedTask;
        }

        public async Task MuteMemberAsync(long selfId, long groupId, long userId, int durationSeconds)
        {
            // 实际调用底层协议
            await Task.CompletedTask;
        }

        public async Task UnmuteMemberAsync(long selfId, long groupId, long userId)
        {
            // 实际调用底层协议
            await Task.CompletedTask;
        }

        public async Task SetGroupTitleAsync(long selfId, long groupId, long userId, string title)
        {
            // 实际调用底层协议
            await Task.CompletedTask;
        }

        public async Task SendMessageAsync(long selfId, long groupId, long userId, string message, bool isGroup = true)
        {
            // 实际调用底层协议
            await Task.CompletedTask;
        }

        public async Task SetGroupNameAsync(long selfId, long groupId, string groupName)
        {
            // 实际调用底层协议
            await Task.CompletedTask;
        }

        public async Task LeaveGroupAsync(long selfId, long groupId)
        {
            // 实际调用底层协议
            await Task.CompletedTask;
        }

        public async Task SetGroupSpecialTitleAsync(long selfId, long groupId, long userId, string title)
        {
            // 实际调用底层协议
            await Task.CompletedTask;
        }
    }
}
