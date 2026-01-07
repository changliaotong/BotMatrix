using BotWorker.Common.Exts;
using BotWorker.Core.MetaDatas;

namespace BotWorker.Agents.Entries
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
