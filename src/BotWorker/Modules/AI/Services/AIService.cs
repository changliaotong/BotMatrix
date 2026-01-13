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
        Task<string> ChatWithContextAsync(string prompt, IPluginContext? context, string? model = null);
        IAsyncEnumerable<string> StreamChatAsync(string prompt, IPluginContext? context = null, string? model = null);
        Task<string> GenerateImageAsync(string prompt, IPluginContext? context = null, string? model = null);
        Task<string> RawChatAsync(string prompt, string? model = null);
        Task<float[]> GenerateEmbeddingAsync(string text, string? model = null);
    }

    public class AIService : IAIService
    {
        private readonly IMcpService _mcpService;
        private readonly IRagService _ragService;
        private readonly IToolAuditService _auditService;
        private readonly IImageGenerationService _imageService;
        private readonly IBillingService _billingService;
        private readonly ILLMRepository _llmRepository;
        private readonly ILLMCallLogRepository _callLogRepository;
        private readonly LLMApp _llmApp;
        private readonly ILogger<AIService> _logger;
        private readonly IServiceProvider _serviceProvider;
        private static readonly Random _random = new();

        public AIService(
            IMcpService mcpService, 
            IRagService ragService,
            IToolAuditService auditService,
            IImageGenerationService imageService,
            IBillingService billingService,
            ILLMRepository llmRepository,
            ILLMCallLogRepository callLogRepository,
            LLMApp llmApp, 
            ILogger<AIService> logger,
            IServiceProvider serviceProvider)
        {
            _mcpService = mcpService;
            _ragService = ragService;
            _auditService = auditService;
            _imageService = imageService;
            _billingService = billingService;
            _llmRepository = llmRepository;
            _callLogRepository = callLogRepository;
            _llmApp = llmApp;
            _logger = logger;
            _serviceProvider = serviceProvider;
        }

        public async Task<string> ChatAsync(string prompt, string? model = null)
        {
            // 兼容旧接口，使用系统默认上下文
            return await ChatWithContextAsync(prompt, null!, model);
        }

        private async Task<(IModelProvider? Provider, string? ModelId, long? ProviderId)> GetEffectiveProviderAsync(string? model, IPluginContext? context, LLMModelType type = LLMModelType.Chat)
        {
            var (provider, modelId) = _llmApp._manager.SelectByStrategy(model ?? "random", type);

            long currentUserId = 0;
            if (context != null && !string.IsNullOrEmpty(context.UserId))
            {
                long.TryParse(context.UserId, out currentUserId);
            }

            // 1. 优先检查用户是否提供了自己的 Key (BYOK)
            if (currentUserId > 0)
            {
                var providerName = provider?.ProviderName ?? model ?? "Doubao";
                var userProvider = await _llmRepository.GetUserProviderAsync(currentUserId, providerName);

                if (userProvider != null && !string.IsNullOrEmpty(userProvider.ApiKey))
                {
                    var decryptedKey = userProvider.GetDecryptedApiKey();
                    var effectiveProvider = new GenericOpenAIProvider(providerName, decryptedKey, userProvider.Endpoint, modelId ?? providerName);
                    _logger.LogInformation("[AIService] Using BYOK for user {UserId}, provider {ProviderName}", currentUserId, providerName);
                    return (effectiveProvider, modelId, userProvider.Id);
                }
            }

            // 2. 如果没有用户 Key，尝试从租赁池中随机选择
            if (provider == null || (currentUserId > 0 && provider.ProviderName == "Doubao" && model == null)) // 默认 Provider 且没有指定模型时，尝试寻找共享 Key
            {
                var providerName = model ?? "Doubao";
                var sharedProviders = (await _llmRepository.GetSharedProvidersAsync(providerName)).ToList();

                if (sharedProviders.Count > 0)
                {
                    var config = sharedProviders[_random.Next(sharedProviders.Count)];
                    var decryptedKey = config.GetDecryptedApiKey();
                    var effectiveProvider = new GenericOpenAIProvider(providerName, decryptedKey, config.Endpoint, modelId ?? providerName);
                    _logger.LogInformation("[AIService] Using Shared Key from provider {ProviderId} for user {UserId}", config.Id, currentUserId);
                    return (effectiveProvider, modelId, config.Id);
                }
            }

            return (provider, modelId, null);
        }

        public async Task<string> ChatWithContextAsync(string prompt, IPluginContext? context, string? model = null)
        {
            try
            {
                // 0. 基础计费与租赁检查
                long currentUserId = 0;
                long tenantId = 0;
                if (context != null)
                {
                    if (!string.IsNullOrEmpty(context.UserId)) long.TryParse(context.UserId, out currentUserId);
                    if (!string.IsNullOrEmpty(context.GroupId)) long.TryParse(context.GroupId, out tenantId);
                }

                // 如果在群组中，优先检查群组（租户）是否有租赁
                long billingId = tenantId > 0 ? tenantId : currentUserId;

                if (billingId > 0)
                {
                    // 检查是否有活跃租赁
                    var hasLease = await _billingService.HasActiveLeaseAsync(billingId, "ai_service");
                    if (!hasLease)
                    {
                        // 检查余额
                        if (!await _billingService.HasSufficientBalanceAsync(billingId, 0.01m))
                        {
                            return "您的账户余额不足且没有有效的 AI 服务租赁，请充值或租赁后再尝试。";
                        }
                    }
                }

                // 1. 获取有效 Provider (支持 BYOK 和共享 Key)
                var (provider, modelId, providerId) = await GetEffectiveProviderAsync(model, context, LLMModelType.Chat);

                if (provider == null)
                {
                    return "没有可用的 AI 提供商。";
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
                plugins.Add(KernelPluginFactory.CreateFromObject(new ImageGenerationPlugin(_imageService), "ImageGeneration"));

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
                {                    ModelId = modelId, // 使用映射到的具体模型 ID
                    Plugins = plugins,
                    Filters = context != null ? new[] { new DigitalEmployeeToolFilter(_auditService, context.UserId ?? "system", "staff") } : null,
                    CancellationToken = default
                };

                var startTime = DateTime.UtcNow;
                var result = await provider.ExecuteAsync(history, options);
                var duration = (int)(DateTime.UtcNow - startTime).TotalMilliseconds;

                // 4. 记录消费与使用情况
                int inputTokens = EstimateTokens(prompt);
                int outputTokens = EstimateTokens(result);

                if (billingId > 0)
                {
                    // 计算 Token 消耗并计费
                    decimal cost = await CalculateCostAsync(inputTokens, outputTokens, modelId ?? provider.ProviderName);
                    await _billingService.ConsumeAsync(billingId, cost, relatedType: "ai_chat", remark: $"AI 聊天调用: {modelId ?? provider.ProviderName} (Token 计费)");
                    
                    // 记录审计日志
                    long? agentId = null;
                    if (context is PluginContext pc && pc.Event is BotMessageEvent bme)
                    {
                        agentId = bme.BotMessage.AgentId;
                    }

                    await _callLogRepository.AddAsync(new LLMCallLog
                    {
                        AgentId = agentId,
                        ModelId = (await _llmRepository.GetModelByNameAsync(modelId ?? provider.ProviderName))?.Id,
                        PromptTokens = inputTokens,
                        CompletionTokens = outputTokens,
                        TotalCost = cost,
                        LatencyMs = duration,
                        IsSuccess = true,
                        CreatedAt = DateTime.UtcNow
                    });
                }

                if (providerId.HasValue)
                {
                    await _llmRepository.UpdateUsageAsync(providerId.Value);
                }

                return result;
            }
            catch (Exception ex)
            {
                _logger.LogError(ex, "AIService Chat Error");
                return $"AI Error: {ex.Message}";
            }
        }

        public async Task<string> RawChatAsync(string prompt, string? model = null)
        {
            try
            {
                var (provider, modelId) = _llmApp._manager.SelectByStrategy(model ?? "random", LLMModelType.Chat);
                if (provider == null) return "没有可用的 AI 提供商。";

                var history = new ChatHistory();
                history.AddUserMessage(prompt);

                var options = new ModelExecutionOptions
                {
                    ModelId = modelId,
                    CancellationToken = default
                };

                return await provider.ExecuteAsync(history, options);
            }
            catch (Exception ex)
            {
                _logger.LogError(ex, "AIService RawChat Error");
                return $"AI Error: {ex.Message}";
            }
        }

        public async Task<float[]> GenerateEmbeddingAsync(string text, string? model = null)
        {
            try
            {
                var (provider, modelId) = _llmApp._manager.SelectByStrategy(model ?? "random", LLMModelType.Embedding);
                if (provider == null)
                {
                    // 如果没找到专用的 Embedding 模型，尝试用默认 Chat 模型（SK 通常支持）
                    (provider, modelId) = _llmApp._manager.SelectByStrategy(model ?? "random", LLMModelType.Chat);
                }

                if (provider == null) return Array.Empty<float>();

                return await provider.GenerateEmbeddingAsync(text, new ModelExecutionOptions { ModelId = modelId });
            }
            catch (Exception ex)
            {
                _logger.LogError(ex, "GenerateEmbeddingAsync error");
                return Array.Empty<float>();
            }
        }

        public async Task<string> GenerateImageAsync(string prompt, IPluginContext? context = null, string? model = null)
        {
            try
            {
                // 0. 基础计费与租赁检查
                long currentUserId = 0;
                long tenantId = 0;
                if (context != null)
                {
                    if (!string.IsNullOrEmpty(context.UserId)) long.TryParse(context.UserId, out currentUserId);
                    if (!string.IsNullOrEmpty(context.GroupId)) long.TryParse(context.GroupId, out tenantId);
                }

                // 如果在群组中，优先检查群组（租户）是否有租赁
                long billingId = tenantId > 0 ? tenantId : currentUserId;

                if (billingId > 0)
                {
                    // 检查是否有活跃租赁
                    var hasLease = await _billingService.HasActiveLeaseAsync(billingId, "ai_image");
                    if (!hasLease)
                    {
                        // 检查余额 (生图费用较高，默认 0.1)
                        if (!await _billingService.HasSufficientBalanceAsync(billingId, 0.1m))
                        {
                            return "❌ 您的账户余额不足且没有有效的 AI 生图租赁，请充值或租赁后再尝试。";
                        }
                    }
                }

                string result = string.Empty;
                string usedModel = model ?? "Doubao";

                // 如果指定了模型，且不是默认的生图模型，则走原有逻辑
                if (!string.IsNullOrEmpty(model) && model != "Doubao")
                {
                    var (provider, modelId, providerId) = await GetEffectiveProviderAsync(model, context, LLMModelType.Image);
                    if (provider != null)
                    {
                        var options = new ModelExecutionOptions
                        {
                            ModelId = modelId,
                            CancellationToken = default
                        };
                        _logger.LogInformation("[AIService] Generating image with provider {ProviderName}, model {ModelId}, prompt: {Prompt}", 
                            provider.ProviderName, modelId, prompt);
                        var imageUrl = await provider.GenerateImageAsync(prompt, options);

                        if (providerId.HasValue)
                        {
                            await _llmRepository.UpdateUsageAsync(providerId.Value);
                        }

                        result = imageUrl.StartsWith("http") ? $"[CQ:image,file={imageUrl}]" : imageUrl;
                        usedModel = modelId ?? provider.ProviderName;
                    }
                }
                else
                {
                    // 默认使用新封装 of ImageGenerationService (带 Prompt 优化)
                    var imageResult = await _imageService.GenerateImageAsync(prompt, true);
                    if (string.IsNullOrEmpty(imageResult))
                    {
                        return "❌ 图像生成失败。";
                    }
                    result = imageResult.StartsWith("[CQ:image") ? imageResult : $"[CQ:image,file={imageResult}]";
                }

                // 4. 记录消费与使用情况
                if (billingId > 0 && !string.IsNullOrEmpty(result))
                {
                    await _billingService.ConsumeAsync(billingId, 0.1m, relatedType: "ai_image", remark: $"AI 生图: {usedModel}");
                }

                return result;
            }
            catch (Exception ex)
            {
                _logger.LogError(ex, "AIService GenerateImage Error");
                return $"AI Image Error: {ex.Message}";
            }
        }

        public async IAsyncEnumerable<string> StreamChatAsync(string prompt, IPluginContext? context = null, string? model = null)
        {
            // 0. 基础计费与租赁检查
            long currentUserId = 0;
            long tenantId = 0;
            if (context != null)
            {
                if (!string.IsNullOrEmpty(context.UserId)) long.TryParse(context.UserId, out currentUserId);
                if (!string.IsNullOrEmpty(context.GroupId)) long.TryParse(context.GroupId, out tenantId);
            }

            // 如果在群组中，优先检查群组（租户）是否有租赁
            long billingId = tenantId > 0 ? tenantId : currentUserId;

            if (billingId > 0)
            {
                // 检查是否有活跃租赁
                var hasLease = await _billingService.HasActiveLeaseAsync(billingId, "ai_service");
                if (!hasLease)
                {
                    // 检查余额
                    if (!await _billingService.HasSufficientBalanceAsync(billingId, 0.01m))
                    {
                        yield return "❌ 您的账户余额不足且没有有效的 AI 服务租赁，请充值或租赁后再尝试。";
                        yield break;
                    }
                }
            }

            var (provider, modelId, providerId) = await GetEffectiveProviderAsync(model, context, LLMModelType.Chat);

            if (provider == null)
            {
                yield return $"❌ 错误：找不到 AI 提供商 '{model}'。";
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
            plugins.Add(KernelPluginFactory.CreateFromObject(new ImageGenerationPlugin(_imageService), "ImageGeneration"));
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
                ModelId = modelId,
                Plugins = plugins,
                Filters = context != null ? new[] { new DigitalEmployeeToolFilter(_auditService, context.UserId ?? "system", "staff") } : null
            };

            var fullContent = new System.Text.StringBuilder();
            var startTime = DateTime.UtcNow;
            await foreach (var chunk in provider.StreamExecuteAsync(history, options))
            {
                fullContent.Append(chunk);
                yield return chunk;
            }
            var duration = (int)(DateTime.UtcNow - startTime).TotalMilliseconds;

            // 4. 记录消费与使用情况
            int inputTokens = EstimateTokens(prompt);
            int outputTokens = EstimateTokens(fullContent.ToString());

            if (billingId > 0)
            {
                decimal cost = await CalculateCostAsync(inputTokens, outputTokens, modelId ?? provider.ProviderName);
                await _billingService.ConsumeAsync(billingId, cost, relatedType: "ai_chat_stream", remark: $"AI 流式对话: {modelId ?? provider.ProviderName}");

                // 记录审计日志
                long? agentId = null;
                if (context is PluginContext pc && pc.Event is BotMessageEvent bme)
                {
                    agentId = bme.BotMessage.AgentId;
                }

                await _callLogRepository.AddAsync(new LLMCallLog
                {
                    AgentId = agentId,
                    ModelId = (await _llmRepository.GetModelByNameAsync(modelId ?? provider.ProviderName))?.Id,
                    PromptTokens = inputTokens,
                    CompletionTokens = outputTokens,
                    TotalCost = cost,
                    LatencyMs = duration,
                    IsSuccess = true,
                    CreatedAt = DateTime.UtcNow
                });
            }

            if (providerId.HasValue)
            {
                await _llmRepository.UpdateUsageAsync(providerId.Value);
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

        private async Task<decimal> CalculateCostAsync(int inputTokens, int outputTokens, string modelName)
        {
            var model = await _llmRepository.GetModelByNameAsync(modelName);
            if (model == null) return 0.01m; // 找不到模型信息，使用保底费用

            decimal inputCost = (inputTokens / 1000m) * model.InputPricePer1kTokens;
            decimal outputCost = (outputTokens / 1000m) * model.OutputPricePer1kTokens;

            decimal totalCost = inputCost + outputCost;
            return totalCost > 0 ? totalCost : 0.01m; // 至少扣 0.01
        }

        private int EstimateTokens(string text)
        {
            if (string.IsNullOrEmpty(text)) return 0;
            // 简单估算：中文字符计 1.5 token，英文字符/数字/标点计 0.3 token
            int chineseChars = text.Count(c => c >= 0x4E00 && c <= 0x9FFF);
            int otherChars = text.Length - chineseChars;
            return (int)Math.Ceiling(chineseChars * 1.5 + otherChars * 0.3);
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



