using System;
using System.Threading.Tasks;
using BotWorker.Bots.BotMessages;

namespace BotWorker.Core.Commands
{
    /// <summary>
    /// 热门指令处理器：处理菜单、签到、积分、笑话等常用业务指令
    /// </summary>
    public class HotCommandHandler
    {
        private readonly IHotCmdService _hotCmdService;

        public HotCommandHandler(IHotCmdService hotCmdService)
        {
            _hotCmdService = hotCmdService;
        }

        public async Task<CommandResult> HandleAsync(BotMessage botMsg)
        {
            // 1. 判定是否为热门指令
            if (_hotCmdService.IsHot(botMsg))
            {
                // 2. 执行热门指令逻辑
                return await _hotCmdService.HandleHotCmdAsync(botMsg);
            }

            return CommandResult.Continue();
        }
    }
}
