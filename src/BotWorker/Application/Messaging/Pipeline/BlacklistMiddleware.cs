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

                // 检查新的全局黑名单表
                if (Domain.Entities.BlackList.IsSystemBlack(botMsg.UserId))
                {
                    botMsg.Answer = $"检测到黑名单用户 {botMsg.UserId}，已拦截其请求。";
                    context.Logger?.LogInformation("[Blacklist] Intercepted system black user {UserId}, message {MessageId}", botMsg.UserId, context.EventId);
                    if (botMsg.IsGroup)
                    {
                         await botMsg.KickOutAsync(botMsg.SelfId, botMsg.RealGroupId, botMsg.UserId);
                    }
                    return;
                }

                // 用户黑名单拦截(参考原 HandleBlackWarnAsync)
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
                        // 私聊：返回拉黑提示
                        botMsg.Answer = $"您已被群({botMsg.GroupId})拉黑";
                        if (botMsg.GroupId != BotInfo.GroupCrm)
                        {
                            // 这里可以根据需要保留或简化逻辑
                            botMsg.Answer += UserInfo.GetResetDefaultGroup(botMsg.UserId);
                        }
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
