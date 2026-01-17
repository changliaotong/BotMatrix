using System.Threading.Tasks;
using BotWorker.Common.Extensions;
using BotWorker.Domain.Interfaces;
using BotWorker.Infrastructure.Communication.OneBot;
using BotWorker.Modules.Plugins;
using BotWorker.Infrastructure.Utils;

namespace BotWorker.Application.Messaging.Pipeline
{
    /// <summary>
    /// 预处理中间件：清洗消息（移除广告、繁转简、处理@me等）
    /// </summary>
    public class PreProcessMiddleware : IMiddleware
    {
        private readonly IUserService _userService;

        public PreProcessMiddleware(IUserService userService)
        {
            _userService = userService;
        }

        public async Task InvokeAsync(IPluginContext context, RequestDelegate next)
        {
            if (context is PluginContext pluginCtx && pluginCtx.Event is BotMessageEvent botMsgEvent)
            {
                var botMsg = botMsgEvent.BotMessage;

                // 1. 同行过滤
                if (await _userService.StartWith285or300Async(botMsg.UserId))
                {
                    botMsg.Reason += "[同行]";
                    context.Logger?.LogInformation("[PreProcess] Filtered: Peer user {UserId} for message {MessageId}", botMsg.UserId, context.EventId);
                    return; // 彻底拦截
                }

                // 2. 处理 @me
                if (botMsg.IsAtMe)
                {
                    botMsg.CurrentMessage = botMsg.CurrentMessage.RemoveUserId(botMsg.SelfId);
                }

                // 3. 识别是否 @ 其它
                botMsg.IsAtOthers = botMsg.IsGroup && botMsg.CurrentMessage.RemoveQqImage().IsHaveUserId();

                // 4. 强制前缀检查
                if (botMsg.Group.IsRequirePrefix)
                {
                    if (!botMsg.CurrentMessage.IsMatch(Regexs.Prefix))
                    {
                        botMsg.Reason += "[前缀]";
                        context.Logger?.LogInformation("[PreProcess] Filtered: Missing required prefix for group {GroupId}, message {MessageId}", botMsg.GroupId, context.EventId);
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


