using System;
using System.ComponentModel.DataAnnotations.Schema;

namespace BotWorker.Modules.AI.Models.Evolution
{
    /// <summary>
    /// 任务执行步骤
    /// 对应数据库表: ai_task_steps
    /// </summary>
    [Table("ai_task_steps")]
    public class TaskStep
    {
        public long Id { get; set; }

        /// <summary>
        /// 关联的任务 ID
        /// </summary>
        public long TaskId { get; set; }

        /// <summary>
        /// 步骤索引 (从 0 开始)
        /// </summary>
        public int StepIndex { get; set; }

        public string? Name { get; set; }

        /// <summary>
        /// 输入数据 (JSONB)
        /// </summary>
        public string? InputData { get; set; }

        /// <summary>
        /// 输出数据 (JSONB)
        /// </summary>
        public string? OutputData { get; set; }

        /// <summary>
        /// 状态
        /// </summary>
        public string? Status { get; set; }

        public int DurationMs { get; set; }

        public string? ErrorMessage { get; set; }
        public DateTime CreatedAt { get; set; } = DateTime.UtcNow;
        public DateTime UpdatedAt { get; set; } = DateTime.UtcNow;
    }
}
