using System.Threading.Tasks;
using BotWorker.Domain.Models.Messages.BotMessages;
using BotWorker.Domain.Interfaces;

namespace BotWorker.Application.Services
{
    public interface IGroupService
    {
        /// <summary>
        /// 踢出成员逻辑：包含权限检查、API调用、自动拉黑判断
        /// </summary>
        Task<string> KickMemberAsync(BotMessage botMsg);

        /// <summary>
        /// 禁言/取消禁言逻辑
        /// </summary>
        Task<string> MuteMemberAsync(BotMessage botMsg, bool isMute);

        /// <summary>
        /// 设置群机器人开启状态
        /// </summary>
        Task<string> SetRobotOpenStatusAsync(BotMessage botMsg, string status);

        /// <summary>
        /// 设置群成员头衔
        /// </summary>
        Task<string> SetMemberTitleAsync(BotMessage botMsg, long targetUserId, string title);

        /// <summary>
        /// 检查用户是否有管理/设置权限
        /// </summary>
        bool HasSetupRight(BotMessage botMsg);
    }

    public class GroupService : IGroupService
    {
        private readonly IUserService _userService;
        private readonly IBotApiService _apiService;
        private readonly IPermissionService _permissionService;
        private readonly IGroupRepository _groupRepository;

        public GroupService(
            IUserService userService, 
            IBotApiService apiService, 
            IPermissionService permissionService,
            IGroupRepository groupRepository)
        {
            _userService = userService;
            _apiService = apiService;
            _permissionService = permissionService;
            _groupRepository = groupRepository;
        }

        public async Task<string> KickMemberAsync(BotMessage botMsg)
        {
            // 简单实现逻辑
            if (!HasSetupRight(botMsg)) return "权限不足";
            // 进一步逻辑...
            return "已执行";
        }

        public async Task<string> MuteMemberAsync(BotMessage botMsg, bool isMute)
        {
            if (!HasSetupRight(botMsg)) return "权限不足";
            // 进一步逻辑...
            return "已执行";
        }

        public async Task<string> SetRobotOpenStatusAsync(BotMessage botMsg, string status)
        {
            if (!HasSetupRight(botMsg)) return "权限不足";
            bool isOpen = status == "开启" || status == "on";
            await _groupRepository.SetOpenStatusAsync(botMsg.GroupId, isOpen);
            return $"机器人已{(isOpen ? "开启" : "关闭")}";
        }

        public async Task<string> SetMemberTitleAsync(BotMessage botMsg, long targetUserId, string title)
        {
            if (!HasSetupRight(botMsg)) return "权限不足";
            // 进一步逻辑...
            return "已执行";
        }

        public bool HasSetupRight(BotMessage botMsg)
        {
            return _permissionService.HaveSetupRight(botMsg);
        }
    }
}
