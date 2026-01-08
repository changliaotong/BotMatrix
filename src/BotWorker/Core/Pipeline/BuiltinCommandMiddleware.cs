using System;
using System.Threading.Tasks;
using BotWorker.Core.Plugin;
using BotWorker.Common;
using BotWorker.Common.Exts;
using BotWorker.Core.Commands;

namespace BotWorker.Core.Pipeline
{
    /// <summary>
    /// 内置指令中间件：作为指令路由器，调度各个 CommandHandler
    /// </summary>
    public class BuiltinCommandMiddleware : IMiddleware
    {
        private readonly AdminCommandHandler _adminHandler;
        private readonly SetupCommandHandler _setupHandler;
        private readonly HotCommandHandler _hotHandler;
        private readonly GameCommandHandler _gameHandler;

        public BuiltinCommandMiddleware(
            AdminCommandHandler adminHandler,
            SetupCommandHandler setupHandler,
            HotCommandHandler hotHandler,
            GameCommandHandler gameHandler)
        {
            _adminHandler = adminHandler;
            _setupHandler = setupHandler;
            _hotHandler = hotHandler;
            _gameHandler = gameHandler;
        }

        public async Task InvokeAsync(IPluginContext context, RequestDelegate next)
        {
            if (context is PluginContext pluginCtx && pluginCtx.Event is Core.OneBot.BotMessageEvent botMsgEvent)
            {
                var botMsg = botMsgEvent.BotMessage;

                // 1. 管理类指令 (踢人、禁言等)
                var adminRes = await _adminHandler.HandleAsync(botMsg);
                if (adminRes.Intercept)
                {
                    botMsg.Answer = adminRes.Message;
                    return;
                }

                // 2. 配置类指令 (开启/关闭、黑名单等)
                var setupRes = await _setupHandler.HandleAsync(botMsg);
                if (setupRes.Intercept)
                {
                    botMsg.Answer = setupRes.Message;
                    return;
                }

                // 3. 热门业务指令 (菜单、签到等)
                var hotRes = await _hotHandler.HandleAsync(botMsg);
                if (hotRes.Intercept)
                {
                    botMsg.Answer = hotRes.Message;
                    return;
                }

                // 4. 核心正则指令解析与判定
                var isCmdMsg = botMsg.CurrentMessage.IsMatch(Bots.Public.BotCmd.GetRegexCmd());
                botMsg.IsCmd = isCmdMsg;

                if (botMsg.IsCmd)
                {
                    (botMsg.CmdName, botMsg.CmdPara) = botMsg.GetCmdPara(botMsg.CurrentMessage, Bots.Public.BotCmd.GetRegexCmd());

                    // 指令有效性检查
                    if (IsInvalidCommand(botMsg))
                    {
                        botMsg.IsCmd = false;
                        botMsg.CmdName = "闲聊";
                        botMsg.CmdPara = botMsg.CurrentMessage;
                    }
                    else
                    {
                        if (botMsg.IsRefresh) botMsg.HandleRefresh();
                        else await botMsg.GetCmdResAsync();

                        if (!string.IsNullOrEmpty(botMsg.Answer)) return;
                    }
                }

                // 4. 确认指令状态
                await botMsg.ConfirmCmdAsync();
                if (!string.IsNullOrEmpty(botMsg.Answer)) return;

                // 5. 游戏类指令 (成语接龙等)
                var gameRes = await _gameHandler.HandleAsync(botMsg);
                if (gameRes.Intercept)
                {
                    botMsg.Answer = gameRes.Message;
                    return;
                }

                // 6. 默认降级处理
                botMsg.CmdPara = botMsg.CurrentMessage; 
                await botMsg.GetCmdResAsync();

                if (botMsg.IsRefresh && !botMsg.Answer.IsNull())
                    botMsg.HandleRefresh();

                if (!string.IsNullOrEmpty(botMsg.Answer))
                    return;
            }

            await next(context);
        }

        private bool IsInvalidCommand(Bots.BotMessages.BotMessage botMsg)
        {
            return (botMsg.CmdName.In("续费", "暗恋", "换群", "换主人", "警告") && !botMsg.CmdPara.IsNull() && !botMsg.CmdPara.IsNum())
                || (botMsg.CmdName.In("剪刀", "石头", "布", "抽奖", "三公", "红", "和", "蓝") && !botMsg.CmdPara.IsNull() && (botMsg.CmdPara.Trim() != "梭哈") && !botMsg.CmdPara.IsNum())
                || (botMsg.CmdName.In("菜单", "领积分", "签到", "爱群主", "笑话", "鬼故事", "早安", "午安", "晚安", "揍群主", "升级", "降级", "结算", "一键改名") && !botMsg.CmdPara.IsNull())
                || (botMsg.CmdName.In("计算") && !botMsg.CmdPara.IsMatch(Regexs.Formula));
        }
    }
}
