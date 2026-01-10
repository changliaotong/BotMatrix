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
            // 1. Planner: 决定行动方案
            // 这里我们利用 Semantic Kernel 的自动函数调用能力，但通过拦截器实现规范中的“Executor”和“Reviewer”职责
            
            // 注意：AIService.ChatWithContextAsync 内部已经配置了 RAG, BotSkills 和 MCP 插件
            // 我们需要一种方式在插件执行前进行审计和风险检查
            
            // 由于 AIService 内部使用了 Semantic Kernel，最优雅的方式是使用 Kernel 的 Function Invoking/Invoked 事件
            // 或者我们可以手动实现一个循环来精确控制 Planner/Executor/Reviewer 职责
            
            return await _aiService.ChatWithContextAsync(prompt, context);
        }
    }
}


