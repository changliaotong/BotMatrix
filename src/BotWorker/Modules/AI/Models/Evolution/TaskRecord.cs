using System;
using System.ComponentModel.DataAnnotations.Schema;

namespace BotWorker.Modules.AI.Models.Evolution
{
    /// <summary>
    /// 任务记录
    /// 对应数据库表: ai_task_records
    /// </summary>
    [Table("ai_task_records")]
    public class TaskRecord
    {
        public long Id { get; set; }

        /// <summary>
        /// 外部引用的执行 ID
        /// </summary>
        public Guid ExecutionId { get; set; }

        public string? Title { get; set; }

        public string? Description { get; set; }

        /// <summary>
        /// 发起者 ID
        /// </summary>
        public long? InitiatorId { get; set; }

        /// <summary>
        /// 承办人 (数字员工实例 ID)
        /// </summary>
        public long? AssigneeId { get; set; }

        /// <summary>
        /// 状态: pending, executing, completed, failed, cancelled
        /// </summary>
        public string Status { get; set; } = "pending";

        public int Progress { get; set; } = 0;

        /// <summary>
        /// 任务计划数据 (JSONB)
        /// </summary>
        public string PlanData { get; set; } = "{}";

        /// <summary>
        /// 任务结果数据 (JSONB)
        /// </summary>
        public string ResultData { get; set; } = "{}";

        public long? ParentTaskId { get; set; }

        public DateTime? StartedAt { get; set; }

        public DateTime? FinishedAt { get; set; }

        public DateTime CreatedAt { get; set; } = DateTime.UtcNow;

        public DateTime UpdatedAt { get; set; } = DateTime.UtcNow;
    }
}
