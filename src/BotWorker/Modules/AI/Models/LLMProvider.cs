using BotWorker.Infrastructure.Persistence.ORM;

namespace BotWorker.Modules.AI.Models
{
    public class LLMProvider : MetaDataGuid<LLMProvider>
    {
        public override string TableName => "LLMProvider";

        public override string KeyField => "Id";

        public string Name { get; set; } = string.Empty;
        public string BaseUrl { get; set; } = string.Empty;
        public string ApiKey { get; set; } = string.Empty;
        public string ProviderType { get; set; } = "openai";
        public string LogoUrl { get; set; } = string.Empty;
        public string Description { get; set; } = string.Empty;

        public static async Task<Dictionary<string, object>> AppendAsync(LLMProvider llmProvider, params string[] fields)
        { 
            return await InsertReturnFieldsAsync(new
            {
                llmProvider.Name,
                llmProvider.BaseUrl,
                llmProvider.ApiKey,
                llmProvider.ProviderType,
                llmProvider.LogoUrl,
                llmProvider.Description
            }, fields);
        }

        public static async Task<int?> UpdateAsync(LLMProvider llmProvider)
        {
            return await UpdateObjectAsync(new
            {
                llmProvider.Name,
                llmProvider.BaseUrl,
                llmProvider.ApiKey,
                llmProvider.ProviderType,
                llmProvider.LogoUrl,
                llmProvider.Description,
                UpdateAt = DateTime.Now,
            }, llmProvider.Id);
        }

        public static async Task<int?> RemoveAsync(LLMProvider llmProvider)
        {
            return await DeleteAsync(llmProvider.Id);
        }

        public static async Task<List<LLMProvider>?> GetAllAsync()
        {
            var sql = $"SELECT * FROM {FullName}";
            return await QueryListAsync<LLMProvider>(sql);
        }

        public static async Task<List<LLMProvider>?> GetAllActiveAsync()
        {
            return await GetAllAsync();
        }
    }
}
