using BotWorker.Common.Extensions;
using BotWorker.Infrastructure.Utils;

namespace BotWorker.Application.Messaging.Handlers
{
    public class AdminCommandHandler
    {
        private readonly Services.IGroupService _groupService;

        public AdminCommandHandler(Services.IGroupService groupService)
        {
            _groupService = groupService;
        }

        public async Task<CommandResult> HandleAsync(BotMessage botMsg)
        {
            if (botMsg.CmdName.In("踢", "T", "t", "剔除", "移除"))
            {
                var answer = await _groupService.KickMemberAsync(botMsg);
                return CommandResult.Intercepted(answer);
            }

            if (botMsg.CmdName.In("禁言", "取消禁言"))
            {
                var isMute = botMsg.CmdName == "禁言";
                var answer = await _groupService.MuteMemberAsync(botMsg, isMute);
                return CommandResult.Intercepted(answer);
            }

            if (botMsg.CmdName.In("设置头衔", "头衔"))
            {
                var targetUserId = botMsg.CurrentMessage.GetQq();
                var title = botMsg.CurrentMessage.RegexGetValue(Regexs.SetTitle, "title");
                var answer = await _groupService.SetMemberTitleAsync(botMsg, targetUserId, title);
                return CommandResult.Intercepted(answer);
            }

            return CommandResult.Continue();
        }
    }
}
