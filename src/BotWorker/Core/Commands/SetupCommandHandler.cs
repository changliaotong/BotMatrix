using System;
using System.Threading.Tasks;
using BotWorker.Bots.BotMessages;
using BotWorker.Common.Exts;
using BotWorker.Bots.Public;

namespace BotWorker.Core.Commands
{
    public class SetupCommandHandler
    {
        private readonly Services.IGroupService _groupService;
        private readonly Services.IUserService _userService;

        public SetupCommandHandler(Services.IGroupService groupService, Services.IUserService userService)
        {
            _groupService = groupService;
            _userService = userService;
        }

        public async Task<CommandResult> HandleAsync(BotMessage botMsg)
        {
            var message = botMsg.CurrentMessage;
            
            bool isCmdOpen = message.ToLower().In("开启", "#开启", "kq", "#kq");
            bool isCmdBlack = message.IsMatch(BlackList.regexBlack);
            bool isCmdKeyword = message.IsMatch(GroupWarn.RegexCmdWarn);

            if (!isCmdOpen && !isCmdBlack && !isCmdKeyword)
                return CommandResult.Continue();

            // 基础配置检查
            var answer = botMsg.SetupPrivate(true, false);
            if (!string.IsNullOrEmpty(answer))
                return CommandResult.Intercepted(answer);

            // 1. 开启机器人
            if (isCmdOpen && !botMsg.Group.IsOpen)
            {
                answer = await _groupService.SetRobotOpenStatusAsync(botMsg, "开启");
                return CommandResult.Intercepted(answer);
            }

            // 2. 黑名单管理
            if (isCmdBlack)
            {
                (botMsg.CmdName, botMsg.CmdPara) = botMsg.GetCmdPara(message, BlackList.regexBlack);
                botMsg.CmdName = botMsg.CmdName.Replace("黑名单", "拉黑").Replace("加黑", "拉黑").Replace("删黑", "取消拉黑");
                answer = await _userService.HandleBlacklistAsync(botMsg);
                answer += botMsg.GroupId == 0 ? "\n设置群 {默认群}" : "";
                return CommandResult.Intercepted(answer);
            }

            // 3. 敏感词管理
            if (isCmdKeyword)
            {
                answer = GroupWarn.GetEditKeyword(botMsg.GroupId, message);
                answer += !botMsg.IsGroup ? "\n设置群 {默认群}" : "";
                return CommandResult.Intercepted(answer);
            }

            return CommandResult.Continue();
        }
    }
}
