using BotWorker.Infrastructure.Communication.OneBot;

namespace BotWorker.Application.Messaging.Pipeline
{
    /// <summary>
    /// AI中间件：当其他逻辑未拦截时，尝试调用AI进行回复
    /// </summary>
    public class AiMiddleware : IMiddleware
    {
        public async Task InvokeAsync(IPluginContext context, RequestDelegate next)
        {
            if (context is PluginContext pluginCtx && pluginCtx.Event is BotMessageEvent botMsgEvent)
            {
                var botMsg = botMsgEvent.BotMessage;

                // 1. 尝试解析智能体呼�?
                await botMsg.TryParseAgentCall();

                if (botMsg.IsCallAgent)
                {
                    if (botMsg.CmdPara.Trim().IsNull())
                    {
                        // 仅切换智能体，不生成响应
                        botMsg.Answer = UserInfo.SetValue("AgentId", botMsg.CurrentAgent.Id, botMsg.UserId) == -1
                            ? $"变身{RetryMsg}"
                            : $"【{botMsg.CurrentAgent.Name}】{botMsg.CurrentAgent.Info}";
                    }
                    else if (!botMsg.IsWeb)
                    {
                        // 既切换又生成响应
                        await botMsg.GetAgentResAsync();
                    }
                    return; // 拦截，由 AI 负责后续处理
                }

                // 2. 检查用户当前状态是否为 AI 模式，或者是否需要 AI 兜底
                var userStateRes = UserInfo.GetStateRes(botMsg.User.State);
                if (userStateRes == "AI")
                {
                    await botMsg.GetAgentResAsync();
                    return; // 拦截
                }

                // 3. 问答系统未命中时的 AI 兜底 (从 AnswerMessage.cs 移过来的逻辑)
                if (string.IsNullOrEmpty(botMsg.Answer) && (!botMsg.IsCmd || botMsg.CmdName == "闲聊") && !botMsg.IsDup && !botMsg.IsMusic)
                {
                    int cloud = !botMsg.IsGroup || botMsg.IsGuild ? 5 : !botMsg.User.IsShutup ? botMsg.Group.IsCloudAnswer : 0;
                    
                    if ((botMsg.IsAgent || botMsg.IsCallAgent || botMsg.IsAtMe || botMsg.IsGuild || !botMsg.IsGroup || botMsg.IsPublic || (cloud >= 5 && !botMsg.IsAtOthers)) && !botMsg.IsWeb)
                    {
                        await botMsg.GetAgentResAsync();
                        if (!string.IsNullOrEmpty(botMsg.Answer))
                        {
                            return; // 如果 AI 生成了回答，则拦截
                        }
                    }
                }
            }

            await next(context);
        }
    }
}


