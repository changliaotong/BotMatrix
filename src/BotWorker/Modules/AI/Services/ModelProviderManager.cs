using BotWorker.Modules.AI.Interfaces;

namespace BotWorker.Modules.AI.Services
{
    public class ModelProviderManager
    {
        private readonly Dictionary<string, IModelProvider> _providers = [];

        public void RegisterProvider(IModelProvider provider)
        {
            _providers[provider.ProviderName] = provider;
        }

        public IModelProvider? GetProvider(string providerName)
        {
            if (!_providers.TryGetValue(providerName, out var provider))
            {
                return null;
            }

            return provider;
        }
    }

}
