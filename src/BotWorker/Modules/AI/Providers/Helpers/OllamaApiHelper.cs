using Microsoft.SemanticKernel;
using Microsoft.SemanticKernel.ChatCompletion;
using BotWorker.Modules.AI.Interfaces;
using BotWorker.Modules.AI.Providers.Configs;
using System.Collections.Generic;

namespace BotWorker.Modules.AI.Providers.Helpers
{
    public class OllamaApiHelper : OpenAIBaseProvider
    {
        public override string ProviderName => "Ollama";

        public OllamaApiHelper(OllamaConfig config) 
            : base("ollama", config.OllamaUrl.EndsWith("/v1") ? config.OllamaUrl : config.OllamaUrl.TrimEnd('/') + "/v1", config.ModelId)
        {
        }
    }
}
