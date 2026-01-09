using System;
using System.Collections.Generic;
using System.Threading.Tasks;
using BotWorker.Modules.AI.Providers;
using BotWorker.Modules.AI.Interfaces;
using BotWorker.Modules.AI.Plugins;
using BotWorker.Domain.Interfaces;
using Microsoft.SemanticKernel;
using Microsoft.SemanticKernel.ChatCompletion;
using Microsoft.Extensions.Logging;
using Microsoft.Extensions.DependencyInjection;
using System.ComponentModel;
using System.Linq;

namespace BotWorker.Modules.AI.Services
{
    public interface IAIService
    {
        Task<string> ChatAsync(string prompt, string? model = null);
        Task<string> ChatWithContextAsync(string prompt, IPluginContext context, string? model = null);
        IAsyncEnumerable<string> StreamChatAsync(string prompt, IPluginContext? context = null, string? model = null);
    }

    public class AIService : IAIService
    {
        private readonly IMcpService _mcpService;
        private readonly IRagService _ragService;
        private readonly LLMApp _llmApp;
        private readonly ILogger<AIService> _logger;
        private readonly IServiceProvider _serviceProvider;

        public AIService(
            IMcpService mcpService, 
            IRagService ragService,
            LLMApp llmApp, 
            ILogger<AIService> logger,
            IServiceProvider serviceProvider)
        {
            _mcpService = mcpService;
            _ragService = ragService;
            _llmApp = llmApp;
            _logger = logger;
            _serviceProvider = serviceProvider;
        }

        public async Task<string> ChatAsync(string prompt, string? model = null)
        {
            // 兼容旧接口，使用系统默认上下文
            return await ChatWithContextAsync(prompt, null!, model);
        }

        public async Task<string> ChatWithContextAsync(string prompt, IPluginContext? context, string? model = null)
        {
            try
            {
                var providerName = model ?? "DeepSeek";
                var provider = _llmApp._manager.GetProvider(providerName);

                if (provider == null)
                {
                    return $"Error: AI Provider '{providerName}' not found.";
                }

                // 1. 准备插件列表
                var plugins = new List<KernelPlugin>();

                // 1.1 注入本地技能插件 (如果上下文存在)
                if (context != null)
                {
                    using var scope = _serviceProvider.CreateScope();
                    var robot = scope.ServiceProvider.GetRequiredService<IRobot>();
                    plugins.Add(KernelPluginFactory.CreateFromObject(new BotSkillPlugin(robot, context), "BotSkills"));
                }

                // 1.2 注入 RAG 插件
                plugins.Add(KernelPluginFactory.CreateFromObject(new RagPlugin(_ragService), "RAG"));

                // 1.3 注入 MCP 插件 (从 IMcpService 获取工具并转换为插件)
                var mcpPlugins = await GetMcpPluginsAsync(context);
                if (mcpPlugins != null) plugins.AddRange(mcpPlugins);

                // 2. 准备对话历史
                var history = new ChatHistory();
                
                // 如果有 RAG，可以先进行预检索（或者让 AI 通过插件决定）
                // 这里我们采用 AI 决定模式，但也注入一些基础系统提示词
                history.AddSystemMessage("你是一个全能的机器人助手。你可以调用本地技能、查询知识库或使用外部工具来回答问题。");
                history.AddUserMessage(prompt);

                // 3. 执行
                var options = new ModelExecutionOptions
                {
                    ModelId = null, // 使用 Provider 默认模型
                    Plugins = plugins,
                    CancellationToken = default
                };

                return await provider.ExecuteAsync(history, options);
            }
            catch (Exception ex)
            {
                _logger.LogError(ex, "AIService Chat Error");
                return $"AI Error: {ex.Message}";
            }
        }

        public async IAsyncEnumerable<string> StreamChatAsync(string prompt, IPluginContext? context = null, string? model = null)
        {
            var providerName = model ?? "DeepSeek";
            var provider = _llmApp._manager.GetProvider(providerName);

            if (provider == null)
            {
                yield return $"Error: AI Provider '{providerName}' not found.";
                yield break;
            }

            // 1. 准备插件
            var plugins = new List<KernelPlugin>();
            if (context != null)
            {
                using var scope = _serviceProvider.CreateScope();
                var robot = scope.ServiceProvider.GetRequiredService<IRobot>();
                plugins.Add(KernelPluginFactory.CreateFromObject(new BotSkillPlugin(robot, context), "BotSkills"));
            }
            plugins.Add(KernelPluginFactory.CreateFromObject(new RagPlugin(_ragService), "RAG"));
            var mcpPlugins = await GetMcpPluginsAsync(context);
            if (mcpPlugins != null) plugins.AddRange(mcpPlugins);

            // 2. 准备对话历史
            var history = new ChatHistory();
            history.AddSystemMessage("你是一个全能的机器人助手。你可以调用本地技能、查询知识库或使用外部工具来回答问题。");
            history.AddUserMessage(prompt);

            // 3. 执行流式调用
            var options = new ModelExecutionOptions
            {
                Plugins = plugins
            };

            await foreach (var chunk in provider.StreamExecuteAsync(history, options))
            {
                yield return chunk;
            }
        }

        private async Task<List<KernelPlugin>> GetMcpPluginsAsync(IPluginContext? context)
        {
            long userId = 0;
            long orgId = 0;
            if (context != null)
            {
                long.TryParse(context.UserId, out userId);
                if (context.GroupId != null) long.TryParse(context.GroupId, out orgId);
            }

            var mcpTools = await _mcpService.GetToolsForContextAsync(userId, orgId);
            if (mcpTools == null || !mcpTools.Any()) return null!;

            var functions = new List<KernelFunction>();
            foreach (var tool in mcpTools)
            {
                var function = KernelFunctionFactory.CreateFromMethod(
                    async (KernelArguments args) =>
                    {
                        var dictArgs = new Dictionary<string, object>();
                        foreach (var arg in args) dictArgs[arg.Key] = arg.Value ?? "";
                        var response = await _mcpService.CallToolAsync(tool.ServerId, tool.Name, dictArgs);
                        if (response.IsError) return $"Error: {string.Join("\n", response.Content.Select(c => c.Text))}";
                        return string.Join("\n", response.Content.Select(c => c.Text));
                    },
                    tool.Name,
                    tool.Description
                );
                functions.Add(function);
            }

            return new List<KernelPlugin> { KernelPluginFactory.CreateFromFunctions("MCP", functions) };
        }
    }

    // 辅助 RAG 插件类
    public class RagPlugin
    {
        private readonly IRagService _ragService;
        public RagPlugin(IRagService ragService) => _ragService = ragService;

        [KernelFunction]
        [Description("从知识库中搜索相关信息以回答用户问题。")]
        public async Task<string> SearchKnowledge([Description("搜索关键词或问题")] string query)
        {
            var results = await _ragService.SearchAsync(query);
            if (!results.Any()) return "未找到相关知识。";
            return string.Join("\n---\n", results.Select(r => r.Content));
        }
    }
}



