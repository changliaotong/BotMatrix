using System.Threading.Tasks;
using BotWorker.Core.Plugin;
using BotWorker.BotWorker.BotWorker.Common.Exts;

namespace BotWorker.Core.Pipeline
{
    /// <summary>
    /// 自动签到中间件：处理用户的自动签�?打卡逻辑
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

            if (context is PluginContext pluginCtx && pluginCtx.Event is Core.OneBot.BotMessageEvent botMsgEvent)
            {
                var botMsg = botMsgEvent.BotMessage;
                var message = context.RawMessage;

                if (!string.IsNullOrEmpty(message))
                {
                    // 只有非显式签到指令时才尝试自动签�?                    var isAuto = botMsg.CmdName != "签到" && !message.Contains("签到") && !message.Contains("打卡");
                    if (isAuto)
                    {
                        var answer = botMsg.TrySignIn(isAuto) ?? string.Empty;
                        if (!string.IsNullOrWhiteSpace(answer))
                        {
                            botMsg.CostTime = CurrentStopwatch == null ? 0 : CurrentStopwatch.Elapsed.TotalSeconds;
                            // 自动签到成功，发送结�?                            botMsg.Answer = answer;
                            
                            // 自动签到通常作为附加动作，不一定需要终止管道�?                            // 但在原逻辑中，它会立即发送一条消息�?                            // 这里我们将其存入 Answer，让后续流程决定是否发送或继续�?                            // 如果原逻辑是立即发送并清空，我们可以模拟这个行为：
                            var isCmd = botMsg.IsCmd;
                            var isCancelProxy = botMsg.IsCancelProxy;
                            botMsg.IsCmd = true;
                            botMsg.IsCancelProxy = true;
                            //botMsg.IsRecall = CurrentGroup.IsRecall;
                            //botMsg.RecallAfterMs = CurrentGroup.RecallTime;
                            await botMsg.SendMessageAsync();
                            botMsg.IsCmd = isCmd;
                            //botMsg.IsRecall = false;
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


