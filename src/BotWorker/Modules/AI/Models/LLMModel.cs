using BotWorker.Common.Extensions;
using BotWorker.Infrastructure.Persistence.ORM;

namespace BotWorker.Modules.AI.Models
{
    public enum LLMModelType
    {
        Chat = 0,
        Image = 1
    }

    public class LLMModel : MetaDataGuid<LLMModel>
    {
        public override string TableName => "LLMModel";
        public override string KeyField => "Id";

        public string Name { get; set; } = string.Empty;
        public int ProviderId { get; set; }
        public string Memo { get; set; } = string.Empty;
        public int Status { get; set; } = 1;
        public int ModelType { get; set; } = 0; // 0: Chat, 1: Image

        public static int Append(string name, int providerId, string memo = "", int modelType = 0)
        {
            return Insert([
                new Cov("Name", name),
                new Cov("ProviderId", providerId),
                new Cov("Memo", memo),
                new Cov("Status", 1),
                new Cov("ModelType", modelType)
            ]);
        }

        public static (int, string, string) GetModelInfo(int modelId)
        {
            // 检查模型是否可用（模型本身激活 且 关联的提供者也激活）
            bool IsUsable(int id)
            {
                if (id <= 0) return false;
                var sql = $"SELECT COUNT(*) FROM {FullName} m JOIN LLMProvider p ON m.ProviderId = p.Id WHERE m.Id = {id} AND m.Status = 1 AND p.Status = 1";
                return QueryScalar<int>(sql) > 0;
            }

            if (!IsUsable(modelId))
            {
                // 随机选择一个模型激活的
                var sql = $"SELECT {SqlTop(1)} m.Id FROM {FullName} m JOIN LLMProvider p ON m.ProviderId = p.Id WHERE m.Status = 1 AND p.Status = 1 ORDER BY {SqlRandomOrder}{SqlLimit(1)}";
                modelId = QueryScalar<int>(sql);
            }

            if (modelId <= 0) return (0, "", "");

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
