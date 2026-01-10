using System;
using System.Collections.Generic;
using System.Threading.Tasks;
using BotWorker.Modules.AI.Providers;
using BotWorker.Modules.AI.Interfaces;
using BotWorker.Modules.AI.Tools;
using BotWorker.Modules.AI.Filters;
using BotWorker.Modules.AI.Plugins;
using BotWorker.Domain.Interfaces;
using Microsoft.SemanticKernel;
using Microsoft.SemanticKernel.ChatCompletion;
using Microsoft.Extensions.Logging;
using Microsoft.Extensions.DependencyInjection;
using System.ComponentModel;
using System.Linq;

using BotWorker.Modules.AI.Models;
using BotWorker.Modules.AI.Providers.Helpers;
using BotWorker.Modules.Plugins;
using BotWorker.Infrastructure.Communication.OneBot;
using BotWorker.Domain.Entities;

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
        private readonly IToolAuditService _auditService;
        private readonly LLMApp _llmApp;
        private readonly ILogger<AIService> _logger;
        private readonly IServiceProvider _serviceProvider;
        private static readonly Random _random = new();

        public AIService(
            IMcpService mcpService, 
            IRagService ragService,
            IToolAuditService auditService,
            LLMApp llmApp, 
            ILogger<AIService> logger,
            IServiceProvider serviceProvider)
        {
            _mcpService = mcpService;
            _ragService = ragService;
            _auditService = auditService;
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
                // 默认使用 Doubao 模型，如果 model 为空
                // 如果 model 为 "Random"，则由 ModelProviderManager 随机选择
                var providerName = model ?? "Doubao";
                IModelProvider? provider = null;

                // 1. 优先检查用户是否提供了自己的 Key
                if (context is PluginContext pc && pc.Event is BotMessageEvent bme)
                {
                    var userId = bme.BotMessage.UserId;
                    var userConfig = await UserAIConfig.GetUserConfigAsync(userId, providerName);
                    if (userConfig != null && !string.IsNullOrEmpty(userConfig.ApiKey))
                    {
                        _logger.LogInformation("Using user-provided API key for user {UserId} and provider {ProviderName}", userId, providerName);
                        provider = new GenericOpenAIProvider(providerName, userConfig.ApiKey, userConfig.BaseUrl, providerName);
                        _ = UserAIConfig.UpdateUsageAsync(userConfig.Id);
                    }
                }

                // 2. 如果没有用户 Key，且允许租赁，尝试从租赁池中随机选择
                if (provider == null)
                {
                    var leasedConfigs = await UserAIConfig.GetLeasedConfigsAsync(providerName);
                    if (leasedConfigs.Count > 0)
                    {
                        var config = leasedConfigs[_random.Next(leasedConfigs.Count)];
                        _logger.LogInformation("Using leased API key from user {LeaserId} for provider {ProviderName}", config.UserId, providerName);
                        provider = new GenericOpenAIProvider(providerName, config.ApiKey, config.BaseUrl, providerName);
                        _ = UserAIConfig.UpdateUsageAsync(config.Id);
                        
                        // 奖励出租者：增加少量算力
                        _ = UserInfo.AddTokensAsync(0, 0, "算力租赁奖励", config.UserId, "系统", 100, $"您的 API Key 被使用，获得算力奖励");
                    }
                }

                // 3. 最后使用系统配置
                if (provider == null)
                {
                    provider = _llmApp._manager.GetProvider(providerName);
                }

                if (provider == null)
                {
                    // 如果找不到指定的 provider，尝试获取随机一个作为兜底
                    provider = _llmApp._manager.GetRandomProvider();
                    if (provider == null)
                    {
                        return "没有可用的 AI 提供商。";
                    }
                    _logger.LogWarning("Specified AI Provider '{providerName}' not found. Falling back to '{fallbackProvider}'.", providerName, provider.ProviderName);
                }

                // 1. 准备插件列表
                var plugins = new List<KernelPlugin>();

                // 1.0 获取基础参数
                long groupId = 0;
                if (context != null)
                {
                    long.TryParse(context.GroupId, out groupId);
                }

                // 1.1 注入本地技能插件 (如果上下文存在)
                if (context != null)
                {
                    using var scope = _serviceProvider.CreateScope();
                    var robot = scope.ServiceProvider.GetRequiredService<IRobot>();
                    plugins.Add(KernelPluginFactory.CreateFromObject(new BotSkillPlugin(robot, context), "BotSkills"));
                }

                // 1.2 注入 RAG 插件
                plugins.Add(KernelPluginFactory.CreateFromObject(new RagPlugin(_ragService, groupId), "RAG"));
                plugins.Add(KernelPluginFactory.CreateFromObject(new SystemToolPlugin(), "SystemTools"));
                plugins.Add(KernelPluginFactory.CreateFromObject(new SystemAdminPlugin(), "SystemAdmin"));

                // 1.3 注入 MCP 插件 (从 IMcpService 获取工具并转换为插件)
                var mcpPlugins = await GetMcpPluginsAsync(context);
                if (mcpPlugins != null) plugins.AddRange(mcpPlugins);

                // 2. 准备对话历史
                var history = new ChatHistory();
                
                // --- RAG 预检索优化 ---
                if (context != null)
                {
                    // 在调用大模型之前，先尝试检索相关知识并注入上下文
                    // 这样可以减少一次大模型的 Tool Call 回合，提高响应速度
                    var knowledge = await _ragService.GetFormattedKnowledgeAsync(prompt, groupId);
                    if (!string.IsNullOrEmpty(knowledge))
                    {
                        history.AddSystemMessage(knowledge);
                        _logger.LogInformation("RAG pre-retrieval success for group {GroupId}", groupId);
                    }
                }
                
                history.AddSystemMessage("你是一个全能的机器人助手。你可以调用本地技能、查询知识库或使用外部工具来回答问题。");
                history.AddUserMessage(prompt);

                // 3. 执行
                var options = new ModelExecutionOptions
                {
                    ModelId = null, // 使用 Provider 默认模型
                    Plugins = plugins,
                    Filters = context != null ? new[] { new DigitalEmployeeToolFilter(_auditService, context.UserId ?? "system", "staff") } : null,
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
                yield return $"❌ 错误：找不到 AI 提供商 '{providerName}'。";
                yield break;
            }

            // 1. 准备插件
            var plugins = new List<KernelPlugin>();

            long groupId = 0;
            if (context != null)
            {
                long.TryParse(context.GroupId, out groupId);
            }

            if (context != null)
            {
                using var scope = _serviceProvider.CreateScope();
                var robot = scope.ServiceProvider.GetRequiredService<IRobot>();
                plugins.Add(KernelPluginFactory.CreateFromObject(new BotSkillPlugin(robot, context), "BotSkills"));
            }
            plugins.Add(KernelPluginFactory.CreateFromObject(new RagPlugin(_ragService, groupId), "RAG"));
            plugins.Add(KernelPluginFactory.CreateFromObject(new SystemToolPlugin(), "SystemTools"));
            plugins.Add(KernelPluginFactory.CreateFromObject(new SystemAdminPlugin(), "SystemAdmin"));
            var mcpPlugins = await GetMcpPluginsAsync(context);
            if (mcpPlugins != null) plugins.AddRange(mcpPlugins);

            // 2. 准备对话历史
            var history = new ChatHistory();

            // --- RAG 预检索优化 ---
            if (context != null)
            {
                var knowledge = await _ragService.GetFormattedKnowledgeAsync(prompt, groupId);
                if (!string.IsNullOrEmpty(knowledge))
                {
                    history.AddSystemMessage(knowledge);
                }
            }

            history.AddSystemMessage("你是一个全能的机器人助手。你可以调用本地技能、查询知识库或使用外部工具来回答问题。");
            history.AddUserMessage(prompt);

            // 3. 执行流式调用
            var options = new ModelExecutionOptions
            {
                Plugins = plugins,
                Filters = context != null ? new[] { new DigitalEmployeeToolFilter(_auditService, context.UserId ?? "system", "staff") } : null
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
                        if (response.IsError) return $"❌ 错误：{string.Join("\n", response.Content.Select(c => c.Text))}";
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
        private readonly long _groupId;
        public RagPlugin(IRagService ragService, long groupId)
        {
            _ragService = ragService;
            _groupId = groupId;
        }

        [KernelFunction(name: "knowledge_search")]
        [Description("当用户的问题与本群所配置的知识库内容有关时，调用此函数（如学校政策、公司制度等）")]
        [ToolRisk(ToolRiskLevel.Low, "检索知识库中的文档内容")]
        public async Task<string> SearchKnowledge([Description("搜索关键词或问题")] string query)
        {
            // 对齐逻辑：使用统一的格式化输出
            return await _ragService.GetFormattedKnowledgeAsync(query, _groupId);
        }
    }
}



