using System;
using System.Threading.Tasks;
using BotWorker.Domain.Interfaces;
using BotWorker.Common;
using BotWorker.Common.Extensions;

namespace BotWorker.Application.Messaging.Pipeline
{
    /// <summary>
    /// 内置指令中间件：作为指令路由器，调度各个逻辑分支
    /// </summary>
    public class BuiltinCommandMiddleware : IMiddleware
    {
        public async Task InvokeAsync(IMessageContext context, MessageDelegate next)
        {
            var botMsg = context.Message;

            // 1. 核心正则指令解析与判定
            var isCmdMsg = botMsg.CurrentMessage.IsMatch(botMsg.GetRegexCmd());
            botMsg.IsCmd = isCmdMsg;

            if (botMsg.IsCmd)
            {
                (botMsg.CmdName, botMsg.CmdPara) = botMsg.GetCmdPara(botMsg.CurrentMessage, botMsg.GetRegexCmd());

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

            // 2. 确认指令状态
            await botMsg.ConfirmCmdAsync();
            if (!string.IsNullOrEmpty(botMsg.Answer)) return;

            // 3. 默认降级处理
            botMsg.CmdPara = botMsg.CurrentMessage; 
            await botMsg.GetCmdResAsync();

            if (botMsg.IsRefresh && !string.IsNullOrEmpty(botMsg.Answer))
                botMsg.HandleRefresh();

            if (!string.IsNullOrEmpty(botMsg.Answer))
                return;

            await next(context);
        }

        private bool IsInvalidCommand(BotWorker.Domain.Models.Messages.BotMessages.BotMessage botMsg)
        {
            return (botMsg.CmdName.In("续费", "暗恋", "换群", "换主人", "警告") && !string.IsNullOrEmpty(botMsg.CmdPara) && !botMsg.CmdPara.IsNum())
                || (botMsg.CmdName.In("剪刀", "石头", "布", "抽奖", "三公", "牛牛", "牌九", "骰子") && !string.IsNullOrEmpty(botMsg.CmdPara) && (botMsg.CmdPara.Trim() != "梭哈") && !botMsg.CmdPara.IsNum())
                || (botMsg.CmdName.In("菜单", "领积分", "签到", "爱群主", "笑话", "鬼故事", "早安", "午安", "晚安", "揍群主", "升级", "降级", "结算", "一键改名") && !string.IsNullOrEmpty(botMsg.CmdPara))
                || (botMsg.CmdName.In("计算") && !botMsg.CmdPara.IsMatch(Regexs.Formula));
        }
    }
}
