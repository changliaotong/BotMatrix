using System.Linq;
using Microsoft.SemanticKernel;
using Microsoft.SemanticKernel.ChatCompletion;
using BotWorker.Modules.AI.Providers;
using BotWorker.Modules.AI.Interfaces;
using BotWorker.Modules.AI.Tools;
using BotWorker.Domain.Models.Messages.BotMessages;
using System.Reflection;
using Microsoft.Extensions.DependencyInjection;
using System.Text.Json;

namespace BotWorker.Modules.AI.Services
{
    public interface IAgentExecutor
    {
        Task<string> ExecuteAsync(string prompt, Dictionary<string, object>? context = null);
        Task<string> ExecuteAgentTaskAsync(string taskId, string staffId, string prompt, IPluginContext context);
    }

    public class AgentExecutor : IAgentExecutor
    {
        private readonly IAIService _aiService;
        private readonly IToolAuditService _auditService;
        private readonly IServiceProvider _serviceProvider;

        public AgentExecutor(IAIService aiService, IToolAuditService auditService, IServiceProvider serviceProvider)
        {
            _aiService = aiService;
            _auditService = auditService;
            _serviceProvider = serviceProvider;
        }

        public async Task<string> ExecuteAsync(string prompt, Dictionary<string, object>? context = null)
        {
            // 如果 context 中包含 IPluginContext，则尝试使用更严格的 Agentic 流程
            if (context != null && context.TryGetValue("PluginContext", out var ctx) && ctx is IPluginContext pluginContext)
            {
                var taskId = context.TryGetValue("TaskId", out var tid) ? tid.ToString() : Guid.NewGuid().ToString();
                var staffId = context.TryGetValue("StaffId", out var sid) ? sid.ToString() : "system";
                
                return await ExecuteAgentTaskAsync(taskId!, staffId!, prompt, pluginContext);
            }

            return await _aiService.ChatAsync(prompt);
        }

        /// <summary>
        /// 遵循“数字员工工具接口规范 v1”的执行流程
        /// </summary>
        public async Task<string> ExecuteAgentTaskAsync(string taskId, string staffId, string prompt, IPluginContext context)
        {
            // 1. Planner: 任务分析与规划
            // 在复杂任务中，Planner 负责将任务拆解。目前我们通过系统提示词引导 AI 进入规划状态
            var planningPrompt = $@"
[任务 ID: {taskId}]
[员工 ID: {staffId}]
你现在作为 'Planner' 角色。请分析以下用户需求，并规划执行步骤。
如果需求简单，直接开始执行。
如果需求复杂，请先在心中拆解步骤。

用户需求：{prompt}";

            // 2. Executor: 执行任务 (带审计拦截器)
            // AIService 内部已经通过 DigitalEmployeeToolFilter 实现了 Executor 职责和风险控制
            var result = await _aiService.ChatWithContextAsync(planningPrompt, context);

            // 3. Reviewer: 结果校验与总结
            // 规范要求 Reviewer 判定结果是否合规、是否完成了用户目标
            var reviewPrompt = $@"
你现在作为 'Reviewer' 角色。
请评估以下任务执行结果是否完整且符合预期。
如果结果中包含 'ERROR: 该操作属于高风险行为'，请向用户解释原因并引导其联系管理员审批。

任务需求：{prompt}
执行结果：{result}

请给出最终的用户答复：";

            return await _aiService.ChatAsync(reviewPrompt);
        }
    }
}


