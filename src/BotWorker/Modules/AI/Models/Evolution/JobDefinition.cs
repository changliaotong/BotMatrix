using BotWorker.Infrastructure.Persistence.ORM;
using Newtonsoft.Json;

namespace BotWorker.Modules.AI.Models.Evolution
{
    public class JobDefinition : MetaDataGuid<JobDefinition>
    {
        public override string TableName => "JobDefinition";
        public override string KeyField => "Id";

        public string JobId { get; set; } = string.Empty; // 岗位唯一标识，如 software_dev_junior
        public string Name { get; set; } = string.Empty;
        public string Purpose { get; set; } = string.Empty;
        
        public string InputsSchema { get; set; } = "{}"; // JSON
        public string OutputsSchema { get; set; } = "{}"; // JSON
        public string Constraints { get; set; } = "{}"; // JSON
        public string Workflow { get; set; } = "[]"; // JSON Array
        public string EvaluationRule { get; set; } = "{}"; // JSON
        
        public int Version { get; set; } = 1;
        public bool IsActive { get; set; } = true;
        public DateTime CreatedAt { get; set; } = DateTime.Now;

        [JsonIgnore]
        public dynamic? InputsSchemaObj => JsonConvert.DeserializeObject(InputsSchema);
        [JsonIgnore]
        public dynamic? OutputsSchemaObj => JsonConvert.DeserializeObject(OutputsSchema);
        [JsonIgnore]
        public dynamic? ConstraintsObj => JsonConvert.DeserializeObject(Constraints);
        [JsonIgnore]
        public dynamic? WorkflowObj => JsonConvert.DeserializeObject(Workflow);
    }
}
