using BotWorker.Infrastructure.Persistence.ORM;

namespace BotWorker.Modules.AI.Models.Evolution
{
    public class TaskExecution : MetaDataGuid<TaskExecution>
    {
        public override string TableName => "TaskExecution";
        public override string KeyField => "Id";

        public string ExecutionId { get; set; } = string.Empty;
        public string TaskId { get; set; } = string.Empty;
        
        public string StepName { get; set; } = string.Empty;
        public string SkillId { get; set; } = string.Empty;
        
        public string InputData { get; set; } = "{}"; // JSON
        public string OutputData { get; set; } = "{}"; // JSON
        
        public string Status { get; set; } = "Success"; // Success, Fail
        
        // 评估数据
        public int EvaluationScore { get; set; } = 0; // 0-100
        public string EvaluationFeedback { get; set; } = string.Empty;
        public string ErrorMessage { get; set; } = string.Empty;

        // LLM 原始数据
        public string RawPrompt { get; set; } = string.Empty;
        public string RawResponse { get; set; } = string.Empty;

        public DateTime StartedAt { get; set; }
        public DateTime FinishedAt { get; set; }
    }
}
