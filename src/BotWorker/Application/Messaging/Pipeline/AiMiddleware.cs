using System;
using System.Threading.Tasks;
using BotWorker.Core.Plugin;
using BotWorker.BotWorker.BotWorker.Common.Exts;

namespace BotWorker.Core.Pipeline
{
    /// <summary>
    /// AI/智能体中间件：处理智能体切换、状态检查及 AI 响应生成
    /// </summary>
    public class AiMiddleware : IMiddleware
    {
        public async Task InvokeAsync(IPluginContext context, RequestDelegate next)
        {
            if (context is PluginContext pluginCtx && pluginCtx.Event is Core.OneBot.BotMessageEvent botMsgEvent)
            {
                var botMsg = botMsgEvent.BotMessage;

                // 1. 尝试解析智能体呼�?
                await botMsg.TryParseAgentCall();

                if (botMsg.IsCallAgent)
                {
                    if (botMsg.CmdPara.Trim().IsNull())
                    {
                        // 仅切换智能体，不生成响应
                        botMsg.Answer = botMsg.UserInfo.SetValue("AgentId", botMsg.CurrentAgent.Id, botMsg.UserId) == -1
                            ? $"变身{botMsg.RetryMsg}"
                            : $"【{botMsg.CurrentAgent.Name}】{botMsg.CurrentAgent.Info}";
                    }
                    else if (!botMsg.IsWeb)
                    {
                        // 既切换又生成响应
                        await botMsg.GetAgentResAsync();
                    }
                    return; // 拦截，由 AI 负责后续处理
                }

                // 2. 检查用户当前状态是否为 AI 模式
                var userStateRes = botMsg.UserInfo.GetStateRes(botMsg.User.State);
                if (userStateRes == "AI")
                {
                    await botMsg.GetAgentResAsync();
                    return; // 拦截
                }
            }

            await next(context);
        }
    }
}


