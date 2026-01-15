using BotWorker.Infrastructure.Communication.OneBot;
using BotWorker.Modules.AI.Interfaces;

namespace BotWorker.Application.Messaging.Pipeline
{
    /// <summary>
    /// AI中间件：当其他逻辑未拦截时，尝试调用AI进行回复
    /// </summary>
    public class AiMiddleware : IMiddleware
    {
        private readonly IAgentExecutor _agentExecutor;

        public AiMiddleware(IAgentExecutor agentExecutor)
        {
            _agentExecutor = agentExecutor;
        }

        public async Task InvokeAsync(IPluginContext context, RequestDelegate next)
        {
            if (context is PluginContext pluginCtx && pluginCtx.Event is BotMessageEvent botMsgEvent)
            {
                var botMsg = botMsgEvent.BotMessage;
                Serilog.Log.Information("[AiMiddleware] Processing message: {MessageId}, Content: {Content}", botMsg.MsgId, botMsg.Message);

                // 1. 尝试解析智能体呼叫
                await botMsg.TryParseAgentCall();

                if (botMsg.IsCallAgent)
                {
                    Serilog.Log.Information("[AiMiddleware] Agent call detected: {AgentName}, Params: {Params}", botMsg.CurrentAgent?.Name, botMsg.CmdPara);
                    if (botMsg.CmdPara.Trim().IsNull())
                    {
                        // 仅切换智能体，不生成响应
                        botMsg.Answer = UserInfo.SetValue("AgentId", botMsg.CurrentAgent!.Id, botMsg.UserId) == -1
                            ? $"变身{RetryMsg}"
                            : $"【{botMsg.CurrentAgent.Name}】{botMsg.CurrentAgent.Info}";
                    }
                    else if (!botMsg.IsWeb)
                    {
                        // 既切换又生成响应
                        Serilog.Log.Information("[AiMiddleware] Calling agent: {AgentName}", botMsg.CurrentAgent?.Name);
                        
                        // 特殊处理：如果是 dev_orchestrator，启动自主开发循环
                        if (botMsg.CurrentAgent?.Name == "dev_orchestrator")
                        {
                            Serilog.Log.Information("[AiMiddleware] Triggering autonomous loop for dev_orchestrator");
                            botMsg.Answer = await _agentExecutor.ExecuteJobTaskAsync("dev_orchestrator", botMsg.CmdPara, context);
                        }
                        else
                        {
                            Serilog.Log.Information("[AiMiddleware] Calling GetAgentResAsync for agent: {AgentName}", botMsg.CurrentAgent?.Name);
                            await botMsg.GetAgentResAsync();
                        }
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


