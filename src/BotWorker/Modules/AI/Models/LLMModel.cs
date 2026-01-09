using BotWorker.Common.Extensions;
using BotWorker.Infrastructure.Persistence.ORM;

namespace BotWorker.Modules.AI.Models
{
    public class LLMModel : MetaDataGuid<LLMModel>
    {
        public override string TableName => "LLMModel";
        public override string KeyField => "Id";

        public string Name { get; set; } = string.Empty;
        public int ProviderId { get; set; }
        public string Memo { get; set; } = string.Empty;
        public int Status { get; set; } = 1;

        public static int Append(string name, int providerId, string memo = "")
        {
            return Insert([
                new Cov("Name", name),
                new Cov("ProviderId", providerId),
                new Cov("Memo", memo),
                new Cov("Status", 1)
            ]);
        }

        public static (int, string, string) GetModelInfo(int modelId)
        {
            if (modelId == 0) 
                modelId = GetWhere<int>("Id", "Status = 1", "NEWID()");
            
            var providerId = GetInt("ProviderId", modelId);
            var providerName = LLMProvider.GetValue("Name", providerId);
            var modelName = GetValue("Name", modelId);
            
            return (modelId, providerName, modelName);
        }

        public static async Task<List<LLMModel>> GetAllActiveAsync()
        {
            return await QueryListAsync<LLMModel>($"SELECT * FROM {FullName} WHERE Status = 1");
        }
    }
}
