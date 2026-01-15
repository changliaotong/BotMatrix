using BotWorker.Infrastructure.Communication.OneBot;

namespace BotWorker.Application.Messaging.Pipeline
{
    /// <summary>
    /// 黑名单中间件：拦截处于黑名单中的用户
    /// </summary>
    public class BlacklistMiddleware : IMiddleware
    {
        public async Task InvokeAsync(IPluginContext context, RequestDelegate next)
        {
            if (context is PluginContext pluginCtx && pluginCtx.Event is BotMessageEvent botMsgEvent)
            {
                var botMsg = botMsgEvent.BotMessage;

                // 官方黑名单
                if (botMsg.IsBlackSystem)
                {                    
                    context.Logger?.LogInformation("[Blacklist] Intercepted system black user {UserId}, message {MessageId}", botMsg.UserId, context.EventId);
                    if (botMsg.IsGroup && botMsg.Group.IsCloudBlack)
                    {
                         await botMsg.KickOutAsync(botMsg.SelfId, botMsg.RealGroupId, botMsg.UserId);
                    }
                    return;
                }

                // 群黑名单拦截
                if (botMsg.IsBlack)
                {
                    context.Logger?.LogInformation("[Blacklist] Intercepted black user {UserId}, message {MessageId}", botMsg.UserId, context.EventId);
                    if (botMsg.IsGroup)
                    {
                        // 群聊：如果权限足够则踢人
                        if (botMsg.SelfPerm < botMsg.UserPerm && botMsg.SelfPerm < 2)
                        {
                            botMsg.Answer = $"黑名单成员 {botMsg.UserId} 将被踢出群";
                            await botMsg.KickOutAsync(botMsg.SelfId, botMsg.RealGroupId, botMsg.UserId);
                            botMsg.IsRecall = botMsg.Group.IsRecall;
                            botMsg.RecallAfterMs = botMsg.Group.RecallTime * 1000;
                            return; 
                        }
                    }
                    else
                    {
                        // 私聊
                        botMsg.Answer = $"您已被群({botMsg.GroupId})拉黑";
                        return;
                    }
                }

                // 3. 用户灰名单拦截
                if (botMsg.IsGrey)
                {
                    context.Logger?.LogInformation("[Blacklist] Intercepted grey user {UserId}, message {MessageId}", botMsg.UserId, context.EventId);
                    return; // 灰名单静默拦截
                }

                // 4. 敏感词告警(参考原 HandleBlackWarnAsync 中的 Group.IsWarn 分支)
                if (botMsg.IsGroup && botMsg.Group.IsWarn)
                {
                    await botMsg.GetKeywordWarnAsync();
                    if (!string.IsNullOrEmpty(botMsg.Answer))
                    {
                        context.Logger?.LogInformation("[Blacklist] Intercepted keyword warn for user {UserId}, message {MessageId}", botMsg.UserId, context.EventId);
                        botMsg.IsRecall = botMsg.Group.IsRecall;
                        botMsg.RecallAfterMs = botMsg.Group.RecallTime * 1000;
                        return; // 命中了敏感词拦截，停止后续插件执行
                    }
                }
            }

            await next(context);
        }
    }
}
