using Microsoft.SemanticKernel;
using Microsoft.SemanticKernel.ChatCompletion;
using BotWorker.Modules.AI.Interfaces;
using BotWorker.Modules.AI.Providers.Configs;
using System.Collections.Generic;

namespace BotWorker.Modules.AI.Providers.Helpers
{
    public class OpenAIAzureApiHelper : OpenAIBaseProvider
    {
        private readonly OpenAIAzureConfig _azureConfig;
        public override string ProviderName => "Azure OpenAI";

        public OpenAIAzureApiHelper(OpenAIAzureConfig config) 
            : base(config.ApiKey, config.Endpoint, config.DeploymentName)
        {
            _azureConfig = config;
        }

        protected override Kernel BuildKernel(ModelExecutionOptions options)
        {
            var builder = Kernel.CreateBuilder()
                .AddAzureOpenAIChatCompletion(_azureConfig.DeploymentName, _azureConfig.Endpoint, _azureConfig.ApiKey);

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
