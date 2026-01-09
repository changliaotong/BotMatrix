using System;
using System.Threading.Tasks;
using System.Collections.Generic;
using System.Linq;
using Microsoft.SemanticKernel;
using Microsoft.SemanticKernel.ChatCompletion;
using BotWorker.Modules.AI.Providers;
using BotWorker.Modules.AI.Interfaces;
using BotWorker.Domain.Models.Messages.BotMessages;

namespace BotWorker.Modules.AI.Services
{
    public interface IAgentExecutor
    {
        Task<string> ExecuteAsync(string prompt, Dictionary<string, object>? context = null);
    }

    public class AgentExecutor : IAgentExecutor
    {
        private readonly IAIService _aiService;

        public AgentExecutor(IAIService aiService)
        {
            _aiService = aiService;
        }

        public async Task<string> ExecuteAsync(string prompt, Dictionary<string, object>? context = null)
        {
            // 这里我们不再重复逻辑，而是直接调用 AIService 的新接口
            // 如果 context 中包含 IPluginContext，则使用它
            if (context != null && context.TryGetValue("PluginContext", out var ctx) && ctx is BotWorker.Domain.Interfaces.IPluginContext pluginContext)
            {
                return await _aiService.ChatWithContextAsync(prompt, pluginContext);
            }

            return await _aiService.ChatAsync(prompt);
        }
    }
}


