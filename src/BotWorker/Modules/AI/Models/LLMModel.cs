using BotWorker.Common.Extensions;
using BotWorker.Infrastructure.Persistence.ORM;

namespace BotWorker.Modules.AI.Models
{
    public class LLMModel : MetaDataGuid<LLMModel>
    {
        public override string TableName => "LLMModel";
        public override string KeyField => "Id";

        public static int Append(string name, int providerId, string memo = "")
        {
            return Insert([
                new Cov("Name", name),
                new Cov("ProviderId", providerId),
                new Cov("Memo", memo) ]);
        }

        public static (int, string, string) GetModelInfo(int modelId)
        {
            if (modelId == 0) 
                modelId = GetWhere<int>("Id", "Status = 1", "NEWID()");
            return (modelId, LLMProvider.GetValue("Name", GetInt("ProviderId", modelId)), GetValue("Name", modelId));
        }

    }
}
