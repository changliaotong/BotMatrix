using BotWorker.Modules.AI.Interfaces;
using BotWorker.Modules.AI.Models;
using BotWorker.Modules.AI.Providers;
using Microsoft.Extensions.DependencyInjection;

namespace BotWorker.Domain.Models.BotMessages
{
    public partial class BotMessage
    {
        public async Task<string> GetAiConfigResAsync()
        {
            using var scope = LLMApp.ServiceProvider.CreateScope();
            var sp = scope.ServiceProvider;
            var llmRepository = sp.GetRequiredService<ILLMRepository>();

            if (CmdName == "设置Key")
            {
                // 格式：设置Key [提供商] [Key] [BaseUrl(可选)]
                var parts = CmdPara.Split(' ', StringSplitOptions.RemoveEmptyEntries);
                if (parts.Length < 2)
                    return "格式错误。用法：设置Key [提供商] [Key] [BaseUrl(可选)]\n例如：设置Key DeepSeek sk-xxx https://api.deepseek.com/v1";

                var providerName = parts[0];
                var apiKey = parts[1];
                var baseUrl = parts.Length > 2 ? parts[2] : "";

                var config = await llmRepository.GetUserProviderAsync(UserId, providerName) ?? new LLMProvider
                {
                    OwnerId = UserId,
                    Name = providerName,
                    Type = "openai" // 默认设为 openai 兼容类型
                };

                config.SetEncryptedApiKey(apiKey);
                if (!string.IsNullOrEmpty(baseUrl)) config.Endpoint = baseUrl;
                
                var res = await llmRepository.SaveUserProviderAsync(config);
                return res ? $"✅ 已成功设置 {providerName} 的 API Key" : "❌ 设置失败，请稍后重试";
            }
            else if (CmdName == "岗位任务")
            {
                // 格式：岗位任务 [JobId] [任务描述]
                var parts = CmdPara.Split(' ', 2, StringSplitOptions.RemoveEmptyEntries);
                if (parts.Length < 2)
                    return "格式错误。用法：岗位任务 [JobId] [任务描述]\n当前可用岗位：image_refiner, code_reviewer";

                var jobId = parts[0];
                var taskPrompt = parts[1];

                // 获取 AgentExecutor
                var executor = sp.GetRequiredService<IAgentExecutor>();
                var aiService = sp.GetRequiredService<IAIService>();
                var i18nService = sp.GetRequiredService<II18nService>();
                var logger = sp.GetRequiredService<ILogger<BotMessage>>();

                var pluginContext = new PluginContext(
                    new Infrastructure.Communication.OneBot.BotMessageEvent(this),
                    Platform,
                    SelfId.ToString(),
                    aiService,
                    i18nService,
                    logger,
                    User,
                    Group,
                    null, // Member
                    SelfInfo,
                    async msg => { Answer = msg; await SendMessageAsync(); },
                    async (title, artist, jumpUrl, coverUrl, audioUrl) => { await SendMusicAsync(title, artist, jumpUrl, coverUrl, audioUrl); }
                );

                var result = await executor.ExecuteJobTaskAsync(jobId, taskPrompt, pluginContext);
                return result;
            }
            else if (CmdName == "开启租赁")
            {
                if (string.IsNullOrEmpty(CmdPara))
                    return "请指定要开启租赁的提供商名称。用法：开启租赁 [提供商]";

                var config = await llmRepository.GetUserProviderAsync(UserId, CmdPara);
                if (config == null || string.IsNullOrEmpty(config.ApiKey))
                    return $"❌ 您尚未设置 {CmdPara} 的 API Key，无法开启租赁";

                config.IsShared = true;
                await llmRepository.UpdateProviderAsync(config);
                return $"✅ 已开启 {CmdPara} 的算力租赁。当您的 Key 被系统使用时，您将获得算力奖励。";
            }
            else if (CmdName == "关闭租赁")
            {
                if (string.IsNullOrEmpty(CmdPara))
                    return "请指定要关闭租赁的提供商名称。用法：关闭租赁 [提供商]";

                var config = await llmRepository.GetUserProviderAsync(UserId, CmdPara);
                if (config == null)
                    return $"❌ 未找到 {CmdPara} 的配置信息";

                config.IsShared = false;
                await llmRepository.UpdateProviderAsync(config);
                return $"✅ 已关闭 {CmdPara} 的算力租赁。";
            }
            else if (CmdName == "我的Key")
            {
                var configs = await llmRepository.GetUserProvidersAsync(UserId);
                var configList = configs.ToList();
                if (configList.Count == 0)
                    return "您尚未设置任何个人 API Key。";

                var res = "🛠️ 您的 AI 配置信息：\n";
                foreach (var c in configList)
                {
                    var plainKey = c.GetDecryptedApiKey();
                    var keyMasked = plainKey.Length > 8 ? plainKey[..4] + "****" + plainKey[^4..] : "****";
                    res += $"- {c.Name}: {keyMasked} (租赁: {(c.IsShared ? "开" : "关")})\n";
                }
                return res;
            }

            return string.Empty;
        }
    }
}
