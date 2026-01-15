using Microsoft.SemanticKernel;
using Microsoft.SemanticKernel.ChatCompletion;
using BotWorker.Modules.AI.Interfaces;
using BotWorker.Modules.AI.Providers.Configs;
using System.Collections.Generic;

namespace BotWorker.Modules.AI.Providers.Helpers
{
    public class OpenAIAzureApiHelper : OpenAIBaseProvider
    {
        public OpenAIAzureApiHelper(string deploymentName, string endpoint, string apiKey) 
            : base("Azure OpenAI", apiKey, endpoint, deploymentName)
        {
        }

        public override Kernel BuildKernel(ModelExecutionOptions options)
        {
            var builder = Kernel.CreateBuilder()
                .AddAzureOpenAIChatCompletion(_defaultModelId, _url, _apiKey);

            if (options.Plugins != null)
            {
                foreach (var plugin in options.Plugins)
                {
                    builder.Plugins.Add(plugin);
                }
            }

            return builder.Build();
        }
    }
}
