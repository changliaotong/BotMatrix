using System.Threading.Tasks;
using BotWorker.Domain.Entities;

namespace BotWorker.Domain.Interfaces
{
    public interface IPluginContext
    {
        // 基础信息
        string Message { get; }
        string? GroupId { get; }
        string UserId { get; }
        string Platform { get; }
        string BotId { get; }
        string EventType { get; }
        bool IsMessage { get; }
        string RawMessage { get; set; }

        // 丰富实体
        UserInfo? User { get; }
        GroupInfo? Group { get; }
        GroupMember? Member { get; }
        BotInfo? Bot { get; }

        // 服务接口
        // IAIService AI { get; }
        // II18nService I18n { get; }

        // 动作
        Task ReplyAsync(string message);
    }
}


