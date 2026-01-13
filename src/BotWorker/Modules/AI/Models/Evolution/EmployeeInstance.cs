using System;
using System.ComponentModel.DataAnnotations.Schema;

namespace BotWorker.Modules.AI.Models.Evolution
{
    /// <summary>
    /// 数字员工实例 (Evolution - Employee)
    /// 对应数据库表: ai_employee_instances
    /// </summary>
    [Table("ai_employee_instances")]
    public class EmployeeInstance
    {
        public long Id { get; set; }

        /// <summary>
        /// 工号
        /// </summary>
        public string EmployeeId { get; set; } = string.Empty;

        /// <summary>
        /// 关联的机器人底层 ID
        /// </summary>
        public string BotId { get; set; } = string.Empty;

        /// <summary>
        /// 关联的智能体 ID
        /// </summary>
        public long? AgentId { get; set; }

        /// <summary>
        /// 关联的岗位 ID
        /// </summary>
        public long? JobId { get; set; }

        public string? Name { get; set; }

        public string? Title { get; set; }

        public string? Department { get; set; }

        /// <summary>
        /// 在线状态: online, offline, busy
        /// </summary>
        public string OnlineStatus { get; set; } = "offline";

        /// <summary>
        /// 业务状态: idle, working, paused
        /// </summary>
        public string State { get; set; } = "idle";

        /// <summary>
        /// 已消耗 Token (薪酬)
        /// </summary>
        public long SalaryTokenUsed { get; set; } = 0;

        /// <summary>
        /// Token 限制
        /// </summary>
        public long SalaryTokenLimit { get; set; } = 1000000;

        /// <summary>
        /// KPI 评分
        /// </summary>
        public decimal KpiScore { get; set; } = 100.00m;

        /// <summary>
        /// 进化的元数据 (JSONB)
        /// </summary>
        public string ExperienceData { get; set; } = "{}";

        public DateTime CreatedAt { get; set; } = DateTime.UtcNow;

        public DateTime UpdatedAt { get; set; } = DateTime.UtcNow;
    }
}
