using System.Threading.Tasks;
using BotWorker.Bots.BotMessages;

namespace BotWorker.Core.Services
{
    public interface IGroupService
    {
        /// <summary>
        /// 踢出成员逻辑：包含权限检查、API调用、自动拉黑判�?        /// </summary>
        Task<string> KickMemberAsync(BotMessage botMsg);

        /// <summary>
        /// 禁言/取消禁言逻辑
        /// </summary>
        Task<string> MuteMemberAsync(BotMessage botMsg, bool isMute);

        /// <summary>
        /// 设置群机器人开关状�?        /// </summary>
        Task<string> SetRobotOpenStatusAsync(BotMessage botMsg, string status);

        /// <summary>
        /// 设置群成员头�?        /// </summary>
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
        private readonly Repositories.IGroupRepository _groupRepository;

        public GroupService(
            IUserService userService, 
            IBotApiService apiService, 
            IPermissionService permissionService,
            Repositories.IGroupRepository groupRepository)
        {
            _userService = userService;
            _apiService = apiService;
            _permissionService = permissionService;
            _groupRepository = groupRepository;
        }

        public async Task<string> KickMemberAsync(BotMessage botMsg)
        {
            botMsg.IsCancelProxy = true;

            // 1. 权限检�?            if (!_permissionService.HaveSetupRight(botMsg))            
                return "您没有踢人权�?;            

            // 2. 提取目标 QQ
            var qqMatches = System.Text.RegularExpressions.Regex.Matches(botMsg.CurrentMessage, Common.Regexs.QqNumberPattern);
            if (qqMatches.Count == 0)
                return "未识别到有效的目�?QQ";

            var kickedCount = 0;
            var resultMessage = "";
            var isBlackKick = botMsg.Group.IsBlackKick;

            foreach (System.Text.RegularExpressions.Match match in qqMatches)
            {
                var qq = match.Value.BotWorker.BotWorker.Common.Exts.AsLong();

                // 3. 执行踢人 API (现在通过服务调用)
                await _apiService.KickMemberAsync(botMsg.SelfId, botMsg.RealGroupId, qq);
                kickedCount++;

                // 4. 自动拉黑逻辑
                if (isBlackKick)
                {
                    if (await _userService.AddBlackAsync(botMsg, qq, "被踢拉黑") != -1)
                        resultMessage += $"\n{qq} 已拉黑！";
                }
            }

            return kickedCount > 0 
                ? $"�?已T�?{kickedCount} 个成�? + resultMessage 
                : "操作未完�?;
        }

        public async Task<string> MuteMemberAsync(BotMessage botMsg, bool isMute)
        {
            // 这里原本调用 botMsg.GetMuteResAsync()，现在我们将逻辑搬迁过来
            if (!_permissionService.HaveSetupRight(botMsg))
                return "您没有禁言权限";

            var duration = isMute ? 30 * 60 : 0; // 默认30分钟或取�?            var targetQq = botMsg.CurrentMessage.BotWorker.BotWorker.Common.Exts.GetQq();
            
            if (targetQq == 0) return "未识别到目标QQ";

            if (isMute)
                await _apiService.MuteMemberAsync(botMsg.SelfId, botMsg.RealGroupId, targetQq, duration);
            else
                await _apiService.UnmuteMemberAsync(botMsg.SelfId, botMsg.RealGroupId, targetQq);

            return isMute ? $"已禁言 {targetQq}" : $"已取消禁言 {targetQq}";
        }

        public async Task<string> SetRobotOpenStatusAsync(BotMessage botMsg, string status)
        {
            // 通过 Repository 操作数据�?            var isOpen = status == "开�?;
            await _groupRepository.SetIsOpenAsync(botMsg.GroupId, isOpen);
            
            return (isOpen ? "机器人已开�? : "机器人已关闭") + (botMsg.GroupId == 0 ? "\n设置�?{默认群}" : "");
        }

        public async Task<string> SetMemberTitleAsync(BotMessage botMsg, long targetUserId, string title)
        {
            if (!_permissionService.HaveSetupRight(botMsg))
                return "您没有设置头衔的权限";

            if (targetUserId == 0) return "未指定目标QQ";

            await _apiService.SetGroupSpecialTitleAsync(botMsg.SelfId, botMsg.RealGroupId, targetUserId, title);
            return $"已设�?{targetUserId} 的头衔为: {title}";
        }

        public bool HasSetupRight(BotMessage botMsg)
        {
            return _permissionService.HaveSetupRight(botMsg);
        }
    }
}


