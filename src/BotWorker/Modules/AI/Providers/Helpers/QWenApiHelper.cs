using Microsoft.SemanticKernel;
using Microsoft.SemanticKernel.ChatCompletion;
using BotWorker.Modules.AI.Interfaces;
using BotWorker.Modules.AI.Providers.Configs;
using System.Collections.Generic;

namespace BotWorker.Modules.AI.Providers.Helpers
{
    public class QWenApiHelper : OpenAIBaseProvider
    {
        public override string ProviderName => "QWen";

        public QWenApiHelper(QWenConfig config) 
            : base(config.Key, config.Url, config.ModelId)
        {
        }
    }
}
