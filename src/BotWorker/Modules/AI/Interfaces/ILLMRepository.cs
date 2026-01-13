using System.Collections.Generic;
using System.Threading.Tasks;
using BotWorker.Modules.AI.Models;

namespace BotWorker.Modules.AI.Interfaces
{
    public interface ILLMRepository
    {
        // Provider methods
        Task<LLMProvider?> GetProviderByIdAsync(long id);
        Task<IEnumerable<LLMProvider>> GetAllProvidersAsync();
        Task<IEnumerable<LLMProvider>> GetActiveProvidersAsync();
        Task<long> AddProviderAsync(LLMProvider provider);
        Task<bool> UpdateProviderAsync(LLMProvider provider);
        Task<bool> DeleteProviderAsync(long id);

        // Model methods
        Task<LLMModel?> GetModelByIdAsync(long id);
        Task<IEnumerable<LLMModel>> GetModelsByProviderIdAsync(long providerId);
        Task<IEnumerable<LLMModel>> GetActiveModelsAsync();
        Task<long> AddModelAsync(LLMModel model);
        Task<bool> UpdateModelAsync(LLMModel model);
        Task<bool> DeleteModelAsync(long id);
        
        // Private/Shared Key support
        Task<LLMProvider?> GetUserProviderAsync(long userId, string providerName);
        Task<IEnumerable<LLMProvider>> GetSharedProvidersAsync(string providerName);
        Task<IEnumerable<LLMProvider>> GetUserProvidersAsync(long userId);
        Task<bool> SaveUserProviderAsync(LLMProvider provider);
        Task<bool> UpdateUsageAsync(long providerId);

        // Complex AI logic
        Task<(long ModelId, string ProviderName, string ModelName)> GetBestAvailableModelAsync(long preferredModelId);
        Task<LLMModel?> GetModelByNameAsync(string modelName);
    }
}
