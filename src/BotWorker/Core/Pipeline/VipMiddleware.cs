using System.Threading.Tasks;
using BotWorker.Core.Plugin;
using BotWorker.Bots.Groups;

namespace BotWorker.Core.Pipeline
{
    /// <summary>
    /// VIP与权限中间件：处理使用权限、续费提醒、进群确认等
    /// </summary>
    public class VipMiddleware : IMiddleware
    {
        public async Task InvokeAsync(IPluginContext context, RequestDelegate next)
        {
            if (context is PluginContext pluginCtx && pluginCtx.Event is Core.OneBot.BotMessageEvent botMsgEvent)
            {
                var botMsg = botMsgEvent.BotMessage;

                // 1. 检查使用权限
                var isHaveUseRight = botMsg.IsGuild || botMsg.HaveUseRight();
                if (!isHaveUseRight)
                {
                    botMsg.Reason += "[使用权限]";
                    if (!botMsg.IsGroup)
                    {
                        botMsg.Answer = $"群({botMsg.GroupId})你没有使用权限，请联系群主授权";
                        if (botMsg.UserPerm != 0 && botMsg.UserId != botMsg.Group.RobotOwner)
                            botMsg.Answer += botMsg.UserInfo.GetResetDefaultGroup(botMsg.UserId);
                    }
                    return; // 拦截
                }

                // 2. 群聊特殊检查
                if (botMsg.IsGroup && !botMsg.IsGuild)
                {
                    // 通知续费
                    if (botMsg.IsGroup && !botMsg.IsGuild)
                {
                    // 1. VIP 续费提醒
                    if (botMsg.IsVip && (botMsg.IsCmd || botMsg.IsAtMe) && GroupVip.RestDays(botMsg.GroupId) < 0)
                    {
                        botMsg.IsCancelProxy = true;
                        botMsg.Answer = $"本群机器人已过期，请及时续费";                    
                        botMsg.Reason += "[通知续费]";
                        return;
                    }

                    // 2. Sz84 特殊群组逻辑 (管理权限与关注官号检查)
                    if (botMsg.Group.IsSz84)
                    {
                        if (botMsg.SelfPerm < 2)
                        {
                            botMsg.GroupInfo.SetValue("IsSz84", false, botMsg.GroupId);
                        }
                        else
                        {
                            if (!botMsg.UserInfo.SubscribedPublic(botMsg.UserId))
                            {
                                if ((botMsg.IsCmd || botMsg.IsAtMe) && !botMsg.CurrentMessage.IsMatch(Common.Regexs.BindToken))
                                {
                                    botMsg.Answer = "请先设置我为管理员开启功能";
                                    botMsg.IsCancelProxy = true;
                                }
                                botMsg.Reason += "[关注官号]";
                                return;
                            }
                        }
                    }

                    // 3. 官机不在场检查
                    if ((botMsg.IsCmd || botMsg.IsAtMe) && !botMsg.IsProxyInGroup && botMsg.IsRealProxy)
                    {
                        await botMsg.SendOfficalShareAsync();
                        botMsg.Reason += "[官机不在]";
                        return;
                    }

                    // 4. 进群确认
                    var confirmRes = await botMsg.GetConfirmNew();
                    if (!string.IsNullOrEmpty(confirmRes))
                    {
                        botMsg.Answer = confirmRes;
                        return;
                    }
                }
                else if (!botMsg.IsGuild && !botMsg.IsGroup)
                {
                    // 私聊时的过期检查
                    if (!botMsg.Group.IsValid)
                    {
                        botMsg.Answer = GroupVip.IsVipOnce(botMsg.GroupId) 
                            ? $"群({botMsg.GroupId}) 机器人已过期" 
                            : $"群({botMsg.GroupId}) 机器人已过体验期";

                        botMsg.Answer += botMsg.UserInfo.GetResetDefaultGroup(botMsg.UserId);
                        return;
                    }
                }
            }

            await next(context);
        }
    }
}
