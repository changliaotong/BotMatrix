using System.Threading.Tasks;
using BotWorker.Core.Plugin;
using BotWorker.Common.Exts;

namespace BotWorker.Core.Pipeline
{
    /// <summary>
    /// 预处理中间件：清洗消息（移除广告、繁转简、处理@me等）
    /// </summary>
    public class PreProcessMiddleware : IMiddleware
    {
        public async Task InvokeAsync(IPluginContext context, RequestDelegate next)
        {
            if (context is PluginContext pluginCtx && pluginCtx.Event is Core.OneBot.BotMessageEvent botMsgEvent)
            {
                var botMsg = botMsgEvent.BotMessage;

                // 1. 同行过滤
                if (botMsg.UserInfo.StartWith285or300(botMsg.UserId))
                {
                    botMsg.Reason += "[同行]";
                    return; // 彻底拦截
                }

                // 2. 处理 @me
                if (botMsg.IsAtMe)
                {
                    botMsg.CurrentMessage = botMsg.CurrentMessage.RemoveUserId(botMsg.SelfId);
                }

                // 3. 识别是否 @ 其它人
                botMsg.IsAtOthers = botMsg.IsGroup && botMsg.CurrentMessage.RemoveQqImage().IsHaveUserId();

                // 4. 强制前缀检查
                if (botMsg.Group.IsRequirePrefix)
                {
                    if (!botMsg.CurrentMessage.IsMatch(Common.Regexs.Prefix))
                    {
                        botMsg.Reason += "[前缀]";
                        return; // 拦截
                    }
                    // 如果不是指令，移除前缀以便后续处理
                    if (!botMsg.IsCmd)
                        botMsg.CurrentMessage = botMsg.CurrentMessage[1..];
                }

                // 5. 消息清洗
                botMsg.CurrentMessage = botMsg.CurrentMessage.RemoveQqAds();
                botMsg.CurrentMessage = botMsg.CurrentMessage.AsJianti();
            }

            await next(context);
        }
    }
}
