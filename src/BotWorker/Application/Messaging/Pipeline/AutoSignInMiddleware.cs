using System.Diagnostics;
using System.Threading.Tasks;
using BotWorker.Common.Extensions;
using BotWorker.Domain.Interfaces;
using BotWorker.Infrastructure.Communication.OneBot;

namespace BotWorker.Application.Messaging.Pipeline
{
    /// <summary>
    /// 自动签到中间件：处理用户的自动签到/打卡逻辑
    /// </summary>
    public class AutoSignInMiddleware : IMiddleware
    {
        public async Task InvokeAsync(IPluginContext context, RequestDelegate next)
        {
            if (!context.IsMessage)
            {
                await next(context);
                return;
            }

            if (context is PluginContext pluginCtx && pluginCtx.Event is BotMessageEvent botMsgEvent)
            {
                var botMsg = botMsgEvent.BotMessage;
                var message = context.RawMessage;

                if (!string.IsNullOrEmpty(message))
                {
                    var sw = Stopwatch.StartNew();
                    // 只有非显式签到指令时才尝试自动签到
                    var isAuto = botMsg.CmdName != "签到" && !message.Contains("签到") && !message.Contains("打卡");
                    if (isAuto)
                    {
                        var answer = botMsg.TrySignIn(isAuto) ?? string.Empty;
                        if (!string.IsNullOrWhiteSpace(answer))
                        {
                            botMsg.CostTime = sw.Elapsed.TotalSeconds;
                            // 自动签到成功，发送结果
                            botMsg.Answer = answer;
                            
                            var isCmd = botMsg.IsCmd;
                            var isCancelProxy = botMsg.IsCancelProxy;
                            botMsg.IsCmd = true;
                            botMsg.IsCancelProxy = true;
                            await botMsg.SendMessageAsync();
                            botMsg.IsCmd = isCmd;
                            botMsg.IsCancelProxy = isCancelProxy;
                            botMsg.Answer = "";
                        }
                    }
                }
            }

            await next(context);
        }
    }
}


