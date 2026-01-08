using System.Threading.Tasks;
using BotWorker.Core.Plugin;
using BotWorker.Bots.Public;

namespace BotWorker.Core.Pipeline
{
    /// <summary>
    /// 暗语中间件：处理“天王盖地虎”等全局身份验证和状态查询指令
    /// </summary>
    public class SecretSignalMiddleware : IMiddleware
    {
        public async Task InvokeAsync(IPluginContext context, RequestDelegate next)
        {
            if (!context.IsMessage)
            {
                await next(context);
                return;
            }

            var message = context.RawMessage;

            if (context is PluginContext pluginCtx && pluginCtx.Event is Core.OneBot.BotMessageEvent botMsgEvent)
            {
                var botMsg = botMsgEvent.BotMessage;

                // 暗语：天王盖地虎
                if (message.Contains("天王盖地虎"))
                {
                    if (!botMsg.IsProxyInGroup && botMsg.IsRealProxy)
                    {
                        await botMsg.SendOfficalShareAsync();
                    }
                    else if (botMsg.IsBlackSystem)
                    {
                        botMsg.Answer = "你已被列入官方黑名单";
                    }
                    else if (botMsg.IsGreySystem)
                    {
                        botMsg.Answer = "你已被列入官方灰名单";
                    }
                    else if (botMsg.IsBlack)
                    {
                        botMsg.Answer = "你已被列入黑名单";
                    }
                    else if (botMsg.IsGrey)
                    {
                        botMsg.Answer = "你已被列入灰名单";
                    }
                    else if (!botMsg.Group.IsPowerOn && !botMsg.IsGuild)
                    {
                        botMsg.Answer = "机器人已关机，请先开机";
                    }
                    else if (ClientPublic.SubscribeCompayPublic(botMsg.UserId))
                    {
                        botMsg.Answer = "✅ 你已确认身份";
                    }
                    else
                    {
                        botMsg.Answer = "微信搜【早喵AI】公众号，关注后留言【领积分】可领5000积分并完成身份确认";
                    }
                    
                    return; // 命中了暗语，终止管道
                }
            }

            await next(context);
        }
    }
}
