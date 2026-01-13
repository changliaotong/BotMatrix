using System;
using System.Collections.Generic;
using System.Net.Http;
using System.Text;
using System.Text.Json;
using System.Threading.Tasks;
using Microsoft.Extensions.Logging;
using BotWorker.Modules.AI.Providers.Configs;
using BotWorker.Modules.AI.Interfaces;

namespace BotWorker.Modules.AI.Services
{
    public class ImageGenerationService : IImageGenerationService
    {
        private readonly ILogger<ImageGenerationService> _logger;
        private readonly IServiceProvider _serviceProvider;
        private readonly IHttpClientFactory _httpClientFactory;

        public ImageGenerationService(
            ILogger<ImageGenerationService> logger,
            IServiceProvider serviceProvider,
            IHttpClientFactory httpClientFactory)
        {
            _logger = logger;
            _serviceProvider = serviceProvider;
            _httpClientFactory = httpClientFactory;
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

                var apiUrl = DoubaoTxt2Img.Url;
                var requestData = new
                {
                    req_key = "high_aes_general_v21_L",
                    prompt = finalPrompt,
                    model_version = "general_v2.1_L",
                    req_schedule_conf = "general_v20_9B_pe",
                    seed = -1,
                    scale = 3.5,
                    ddim_steps = 25,
                    width = 512,
                    height = 512,
                    use_pre_llm = true,
                    use_sr = true,
                    return_url = true,
                    logo_info = new
                    {
                        add_logo = false,
                        position = 0,
                        language = 0,
                        opacity = 0.3,
                        logo_text_content = ""
                    }
                };

                var jsonRequest = JsonSerializer.Serialize(requestData);
                using var httpClient = _httpClientFactory.CreateClient();
                httpClient.DefaultRequestHeaders.Add("Authorization", $"Bearer {DoubaoTxt2Img.Secret}");

                var content = new StringContent(jsonRequest, Encoding.UTF8, "application/json");
                var response = await httpClient.PostAsync(apiUrl, content);
                var responseContent = await response.Content.ReadAsStringAsync();

                if (response.IsSuccessStatusCode)
                {
                    var responseData = JsonSerializer.Deserialize<ResponseData>(responseContent);
                    if (responseData?.data?.image_urls != null && responseData.data.image_urls.Count > 0)
                    {
                        return responseData.data.image_urls[0];
                    }
                }
                else
                {
                    _logger.LogError("[ImageGenerationService] API call failed: {StatusCode}, {Content}", response.StatusCode, responseContent);
                }
            }
            catch (Exception ex)
            {
                _logger.LogError(ex, "[ImageGenerationService] Error generating image");
            }

            return string.Empty;
        }

        public async Task<string> RefinePromptAsync(string prompt)
        {
            var systemPrompt = @"你是一个专业的 AI 绘画提示词专家（ImagePromptRefinerAgent）。
你的任务是将用户简单的描述转化为详细、专业、具有电影感和艺术感的提示词。
请通过以下步骤进行优化：
1. 风格补全：根据用户描述，补全合适的艺术风格（如写实、插画、赛博朋克、吉卜力风等）。
2. 镜头语言：添加景深、构图方式、特写、广角等描述。
3. 光影与色彩：添加黄金时间、柔光、强烈对比、特定色调等。
4. 细节增强：添加材质、纹理、背景环境的细节。

请直接输出优化后的提示词，不要包含任何解释性文字。建议使用中文（豆包模型对中文支持极佳）。";

            try
            {
                // 使用 IServiceProvider 延迟获取 IAIService，避免循环依赖
                var aiService = _serviceProvider.GetRequiredService<IAIService>();
                
                // 使用 RawChatAsync 并带上系统提示词
                var fullPrompt = $"{systemPrompt}\n\n用户原始描述：{prompt}\n\n请输出优化后的提示词：";
                var refinedPrompt = await aiService.RawChatAsync(fullPrompt);
                return string.IsNullOrEmpty(refinedPrompt) ? prompt : refinedPrompt.Trim();
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
