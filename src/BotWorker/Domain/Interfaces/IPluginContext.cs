using System.Threading.Tasks;
using BotWorker.Domain.Entities;
using Microsoft.Extensions.Logging;

namespace BotWorker.Domain.Interfaces
{
    public interface IPluginContext
    {
        // 基础信息
        string Message { get; }
        string? GroupId { get; }
        string UserId { get; }
        string UserName { get; }
        string Platform { get; }
        string BotId { get; }
        string EventType { get; }
        bool IsMessage { get; }
        string RawMessage { get; set; }
        string? GroupName { get; }

        // 丰富实体
        UserInfo? User { get; }
        GroupInfo? Group { get; }
        GroupMember? Member { get; }
        BotInfo? Bot { get; }

        // 服务接口
        Services.IAIService AI { get; }
        Application.Services.II18nService I18n { get; }
        ILogger Logger { get; }

        // 动作
        Task ReplyAsync(string message);
        
        // 状态存储（插件内共享）
        void SetState<T>(string key, T value);
        T? GetState<T>(string key);

        // 会话支持
        bool IsConfirmed { get; set; }
        string? SessionAction { get; set; }
        string? SessionStep { get; set; }
        object? SessionData { get; set; }
    }
}


