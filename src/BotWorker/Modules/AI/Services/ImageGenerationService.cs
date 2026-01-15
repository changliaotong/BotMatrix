using System;
using System.Collections.Generic;
using System.Net.Http;
using System.Text;
using System.Text.Json;
using System.Threading.Tasks;
using Microsoft.Extensions.Logging;
using Microsoft.SemanticKernel;
using Microsoft.SemanticKernel.TextToImage;
using BotWorker.Modules.AI.Plugins;
using BotWorker.Modules.AI.Interfaces;
using BotWorker.Modules.AI.Providers.Configs;

namespace BotWorker.Modules.AI.Services
{
    public class ImageGenerationService : IImageGenerationService
    {
        private readonly ILogger<ImageGenerationService> _logger;
        private readonly IServiceProvider _serviceProvider;
        private readonly IAgentExecutor _agentExecutor;
        private readonly ModelProviderManager _modelProviderManager;

        public ImageGenerationService(
            ILogger<ImageGenerationService> logger,
            IServiceProvider serviceProvider,
            IAgentExecutor agentExecutor,
            ModelProviderManager modelProviderManager)
        {
            _logger = logger;
            _serviceProvider = serviceProvider;
            _agentExecutor = agentExecutor;
            _modelProviderManager = modelProviderManager;
        }

        public async Task<string> GenerateImageAsync(string prompt, bool refinePrompt = false)
        {
            try
            {
                string finalPrompt = prompt;
                if (refinePrompt)
                {
                    finalPrompt = await RefinePromptAsync(prompt);
                    _logger.LogInformation("[ImageGenerationService] Refined prompt: {RefinedPrompt}", finalPrompt);
                }

                // 从数据库获取生图模型
                var (provider, modelId, _, _) = _modelProviderManager.GetProviderAndModel(null, LLMModelType.Image);
                if (provider == null)
                {
                    _logger.LogError("[ImageGenerationService] No image generation provider found in database.");
                    return string.Empty;
                }

                // 直接调用 provider 的 GenerateImageAsync，它内部已经实现了 SK 逻辑
                return await provider.GenerateImageAsync(finalPrompt, new ModelExecutionOptions { ModelId = modelId });
            }
            catch (Exception ex)
            {
                _logger.LogError(ex, "[ImageGenerationService] Error generating image");
            }

            return string.Empty;
        }

        public async Task<string> RefinePromptAsync(string prompt)
        {
            try
            {
                // 优先尝试使用专门的“文生图提示词生成器”智能体
                var refinedPrompt = await _agentExecutor.ExecuteByAgentGuidAsync(
                    BotWorker.Modules.AI.Models.AgentInfos.DallEAgent.Guid, 
                    prompt
                );

                if (!string.IsNullOrEmpty(refinedPrompt) && !refinedPrompt.StartsWith("❌"))
                {
                    return refinedPrompt.Trim();
                }

                // 如果智能体执行失败，则回退到内置的系统提示词逻辑
                _logger.LogWarning("[ImageGenerationService] Agent refinement failed, falling back to built-in logic.");
                
                var systemPrompt = @"你是一个专业的 AI 绘画提示词专家（ImagePromptRefinerAgent）。
你的任务是将用户简单的描述转化为详细、专业、具有电影感和艺术感的提示词。
请通过以下步骤进行优化：
1. 风格补全：根据用户描述，补全合适的艺术风格（如写实、插画、赛博朋克、吉卜力风等）。
2. 镜头语言：添加景深、构图方式、特写、广角等描述。
3. 光影与色彩：添加黄金时间、柔光、强烈对比、特定色调等。
4. 细节增强：添加材质、纹理、背景环境的细节。

请直接输出优化后的提示词，不要包含任何解释性文字。建议使用中文（豆包模型对中文支持极佳）。";

                var aiService = _serviceProvider.GetRequiredService<IAIService>();
                var fullPrompt = $"{systemPrompt}\n\n用户原始描述：{prompt}\n\n请输出优化后的提示词：";
                var result = await aiService.RawChatAsync(fullPrompt);
                return string.IsNullOrEmpty(result) ? prompt : result.Trim();
            }
            catch (Exception ex)
            {
                _logger.LogError(ex, "[ImageGenerationService] Error refining prompt");
                return prompt;
            }
        }

        private class ResponseData
        {
            public int code { get; set; }
            public Data? data { get; set; }
            public string message { get; set; } = string.Empty;
            public string request_id { get; set; } = string.Empty;
            public int status { get; set; }
            public string time_elapsed { get; set; } = string.Empty;
        }

        private class Data
        {
            public List<string> image_urls { get; set; } = [];
        }
    }
}
