using BotWorker.Infrastructure.Communication.OneBot;

namespace BotWorker.Application.Messaging.Pipeline
{
    /// <summary>
    /// VIP与权限中间件：处理使用权限、续费提醒、进群确认等
    /// </summary>
    public class VipMiddleware : IMiddleware
    {
        public async Task InvokeAsync(IPluginContext context, RequestDelegate next)
        {
            if (context is PluginContext pluginCtx && pluginCtx.Event is Infrastructure.Communication.OneBot.BotMessageEvent botMsgEvent)
            {
                var botMsg = botMsgEvent.BotMessage;
                    Serilog.Log.Information("[VipMiddleware] Checking rights for message {MessageId}. IsGroup: {IsGroup}, GroupId: {GroupId}, UseRight: {UseRight}", 
                        botMsg.MsgId, botMsg.IsGroup, botMsg.GroupId, botMsg.Group?.UseRight);

                // 1. 检查使用权限 (私聊与频道默认有权，群聊需检查配置)
                var isHaveUseRight = !botMsg.IsGroup || botMsg.IsGuild || botMsg.HaveUseRight();
                if (!isHaveUseRight)
                {
                    Serilog.Log.Warning("[VipMiddleware] Access denied for message {MessageId}. HaveUseRight: {HaveUseRight}", botMsg.MsgId, botMsg.HaveUseRight());
                    botMsg.Reason += "[使用权限]";
                    botMsg.IsSend = false; // 不回消息
                    return; // 拦截
                }

                // 2. 群聊特殊检查
                if (botMsg.IsGroup && !botMsg.IsGuild)
                {
                    // 1. VIP 续费提醒
                    if (botMsg.IsVip && (botMsg.IsCmd || botMsg.IsAtMe) && GroupVip.RestDays(botMsg.GroupId) < 0)
                    {
                        Serilog.Log.Warning("[VipMiddleware] Group expired for message {MessageId}", botMsg.MsgId);
                        botMsg.IsCancelProxy = true;
                        botMsg.Answer = $"本群机器人已过期，请及时续费";
                        botMsg.Reason += "[通知续费]";
                        return;
                    }

                    // 2. Sz84 特殊群组逻辑 (管理权限与关注官号检查)
                    if (botMsg.Group?.IsSz84 == true)
                    {
                        if (botMsg.SelfPerm < 2)
                        {
                            GroupInfo.SetValue("IsSz84", false, botMsg.GroupId);
                        }
                        else
                        {
                            if (!UserInfo.SubscribedPublic(botMsg.UserId))
                            {
                                if ((botMsg.IsCmd || botMsg.IsAtMe) && !botMsg.CurrentMessage.IsMatch(Regexs.BindToken))
                                {
                                    botMsg.Answer = "请先设置我为管理员开启功能";
                                    botMsg.IsCancelProxy = true;
                                }
                                Serilog.Log.Warning("[VipMiddleware] User not subscribed for message {MessageId}", botMsg.MsgId);
                                botMsg.Reason += "[关注官号]";
                                return;
                            }
                        }
                    }

                    // 3. 官机不在场检查
                    if ((botMsg.IsCmd || botMsg.IsAtMe) && !botMsg.IsProxyInGroup && botMsg.IsRealProxy)
                    {
                        Serilog.Log.Warning("[VipMiddleware] Proxy not in group for message {MessageId}", botMsg.MsgId);
                        await botMsg.SendOfficalShareAsync();
                        botMsg.Reason += "[官机不在]";
                        return;
                    }

                    // 4. 进群确认
                    var confirmRes = await botMsg.GetConfirmNew();
                    if (!string.IsNullOrEmpty(confirmRes))
                    {
                        Serilog.Log.Warning("[VipMiddleware] Confirm new user for message {MessageId}", botMsg.MsgId);
                        botMsg.Answer = confirmRes;
                        return;
                    }
                }
                else if (!botMsg.IsGuild && !botMsg.IsGroup)
                {
                    // 私聊时的过期检查
                    if (botMsg.Group?.IsValid == false)
                    {
                        Serilog.Log.Warning("[VipMiddleware] Private message group invalid for message {MessageId}", botMsg.MsgId);
                        botMsg.Answer = GroupVip.IsVipOnce(botMsg.GroupId)
                            ? $"({botMsg.GroupId}) 机器人已过期"
                            : $"({botMsg.GroupId}) 机器人已过体验期";

                        botMsg.IsSend = false; // 不回消息
                        return;
                    }
                }
            }

            Serilog.Log.Information("[VipMiddleware] Passing message {MessageId} to next middleware", context is PluginContext pc && pc.Event is BotMessageEvent bme ? bme.BotMessage.MsgId : "unknown");
            await next(context);
        }
    }
}
