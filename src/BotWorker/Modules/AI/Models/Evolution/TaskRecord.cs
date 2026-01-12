using BotWorker.Infrastructure.Persistence.ORM;

namespace BotWorker.Modules.AI.Models.Evolution
{
    public class TaskRecord : MetaDataGuid<TaskRecord>
    {
        public override string TableName => "TaskRecord";
        public override string KeyField => "Id";

        public string TaskId { get; set; } = string.Empty;
        public string EmployeeId { get; set; } = string.Empty;
        
        public string InputPayload { get; set; } = "{}"; // JSON
        public string ResultOutput { get; set; } = string.Empty;
        public string Status { get; set; } = "Pending"; // Pending, InProgress, Completed, Failed
        
        public int FinalScore { get; set; } = 0;
        public bool IsEvolved { get; set; } = false; // 是否已被用于进化分析

        public DateTime CreatedAt { get; set; } = DateTime.Now;
    }
}
