using BotWorker.Domain.Models.Messages.BotMessages;

namespace BotWorker.Application.Services
{
    /// <summary>
    /// 权限校验服务：统一管理机器人系统内的各类权限判�?    /// </summary>
    public interface IPermissionService
    {
        bool IsMaster(long userId);
        bool IsAdmin(BotMessage botMsg);
        bool IsGroupOwner(BotMessage botMsg);
        bool HaveSetupRight(BotMessage botMsg);
        bool HaveUseRight(BotMessage botMsg);
    }

    public class PermissionService : IPermissionService
    {
        public bool IsMaster(long userId)
        {
            // 这里封装原本�?Master 判定逻辑
            return userId == 1000000; // 示例
        }

        public bool IsAdmin(BotMessage botMsg)
        {
            return botMsg.UserPerm >= 2;
        }

        public bool IsGroupOwner(BotMessage botMsg)
        {
            return botMsg.UserId == botMsg.Group.RobotOwner;
        }

        public bool HaveSetupRight(BotMessage botMsg)
        {
            // 封装原本�?BotMessage.HaveSetupRight 中的逻辑
            return IsMaster(botMsg.UserId) || IsGroupOwner(botMsg) || IsAdmin(botMsg);
        }

        public bool HaveUseRight(BotMessage botMsg)
        {
            // 封装原本�?BotMessage.HaveUseRight 中的逻辑
            return botMsg.IsVip || botMsg.IsGuild;
        }
    }
}


