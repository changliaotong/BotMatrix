using BotWorker.Modules.AI.Interfaces;
using BotWorker.Modules.AI.Models;
using Microsoft.Extensions.Logging;

namespace BotWorker.Modules.AI.Interfaces
{
    public interface IModelProviderFactory
    {
        IModelProvider? CreateProvider(LLMProvider provider, string defaultModel);
    }
}
