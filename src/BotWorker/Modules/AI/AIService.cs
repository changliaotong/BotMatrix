using System;
using System.Collections.Generic;
using System.Threading.Tasks;
using BotWorker.Agents.Providers;
using Microsoft.SemanticKernel.ChatCompletion;

namespace BotWorker.Services
{
    public interface IAIService
    {
        Task<string> ChatAsync(string prompt, string? model = null);
    }

    public class AIService : IAIService
    {
        private readonly IMcpService _mcpService;
        private readonly LLMApp _llmApp;

        public AIService(IMcpService mcpService, LLMApp llmApp)
        {
            _mcpService = mcpService;
            _llmApp = llmApp;
        }

        public async Task<string> ChatAsync(string prompt, string? model = null)
        {
            // 如果未指定模型，默认使用 DeepSeek
            var providerName = model ?? "DeepSeek";
            var provider = _llmApp._manager.GetProvider(providerName);

            if (provider == null)
            {
                // 如果找不到指定的 Provider，尝试返回第一个可用的
                // 这里为了简单，如果找不到就返回错误信息，或者可以根据需求回退到默认
                return $"Error: AI Provider '{providerName}' not found.";
            }

            var history = new ChatHistory();
            history.AddUserMessage(prompt);

            try
            {
                return await provider.ExecuteAsync(history, "");
            }
            catch (Exception ex)
            {
                return $"AI Error: {ex.Message}";
            }
        }
    }
}



