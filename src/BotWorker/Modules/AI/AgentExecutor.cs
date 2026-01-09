using System;
using System.Threading.Tasks;
using System.Collections.Generic;
using System.Linq;
using Microsoft.SemanticKernel;
using Microsoft.SemanticKernel.ChatCompletion;
using BotWorker.Agents.Providers;
using BotWorker.Domain.Models.Messages.BotMessages;

namespace BotWorker.Services
{
    public interface IAgentExecutor
    {
        Task<string> ExecuteAsync(string prompt, Dictionary<string, object>? context = null);
    }

    public class AgentExecutor : IAgentExecutor
    {
        private readonly IMcpService _mcpService;
        private readonly LLMApp _llmApp;

        public AgentExecutor(IMcpService mcpService, LLMApp llmApp)
        {
            _mcpService = mcpService;
            _llmApp = llmApp;
        }

        public async Task<string> ExecuteAsync(string prompt, Dictionary<string, object>? context = null)
        {
            long userId = 0;
            long orgId = 0;

            if (context != null)
            {
                if (context.TryGetValue("UserId", out var uid)) userId = Convert.ToInt64(uid);
                if (context.TryGetValue("OrgId", out var oid)) orgId = Convert.ToInt64(oid);
            }

            // 1. 获取 MCP 工具并转换为 SK 插件
            var mcpTools = await _mcpService.GetToolsForContextAsync(userId, orgId);
            var functions = new List<KernelFunction>();

            foreach (var tool in mcpTools)
            {
                var function = KernelFunctionFactory.CreateFromMethod(
                    async (KernelArguments args) =>
                    {
                        var dictArgs = new Dictionary<string, object>();
                        foreach (var arg in args)
                        {
                            dictArgs[arg.Key] = arg.Value ?? "";
                        }
                        var response = await _mcpService.CallToolAsync(tool.ServerId, tool.Name, dictArgs);
                        if (response.IsError)
                            return $"Error calling {tool.Name}: {string.Join("\n", response.Content.Select(c => c.Text))}";
                        return string.Join("\n", response.Content.Select(c => c.Text));
                    },
                    tool.Name,
                    tool.Description
                );
                functions.Add(function);
            }

            var plugins = new List<KernelPlugin>();
            if (functions.Any())
            {
                plugins.Add(KernelPluginFactory.CreateFromFunctions("MCPTools", functions));
            }

            // 2. 获取 AI 提供商
            var provider = _llmApp._manager.GetProvider("DeepSeek") 
                        ?? _llmApp._manager.GetProvider("Azure OpenAI")
                        ?? _llmApp._manager.GetProvider("OpenAI");

            if (provider == null) return "Error: No AI provider available.";

            // 3. 准备对话历史
            var history = new ChatHistory();
            history.AddUserMessage(prompt);

            // 4. 准备上下文对象
            var botMessage = new BotMessage 
            { 
                User = new UserInfo { Id = userId },
                Group = new GroupInfo { Id = orgId }
            };

            // 5. 执行
            try
            {
                return await provider.ExecuteAsync(history, "", botMessage, plugins);
            }
            catch (Exception ex)
            {
                return $"Agent Execution Error: {ex.Message}";
            }
        }
    }
}


