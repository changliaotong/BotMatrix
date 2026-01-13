using System;
using System.ComponentModel.DataAnnotations.Schema;

namespace BotWorker.Modules.AI.Models.Evolution
{
    /// <summary>
    /// 岗位定义 (Evolution - Job)
    /// 对应数据库表: ai_job_definitions
    /// </summary>
    [Table("ai_job_definitions")]
    public class JobDefinition
    {
        public long Id { get; set; }

        /// <summary>
        /// 岗位唯一标识，如 software_engineer
        /// </summary>
        public string JobKey { get; set; } = string.Empty;

        public string Name { get; set; } = string.Empty;

        public string Purpose { get; set; } = string.Empty;

        /// <summary>
        /// 输入 Schema (JSONB)
        /// </summary>
        public string InputsSchema { get; set; } = "{}";

        /// <summary>
        /// 输出 Schema (JSONB)
        /// </summary>
        public string OutputsSchema { get; set; } = "{}";

        /// <summary>
        /// 岗位约束 (JSONB Array)
        /// </summary>
        public string Constraints { get; set; } = "[]";

        /// <summary>
        /// 系统提示词模板
        /// </summary>
        public string SystemPrompt { get; set; } = string.Empty;

        /// <summary>
        /// 技能工具定义 (JSONB Array)
        /// </summary>
        public string ToolSchema { get; set; } = "[]";

        /// <summary>
        /// 标准执行步骤 (JSONB Array)
        /// </summary>
        public string Workflow { get; set; } = "[]";

        /// <summary>
        /// 模型选择策略: specified, random, provider:xxx, cheapest, fastest
        /// </summary>
        public string ModelSelectionStrategy { get; set; } = "random";

        public int Version { get; set; } = 1;

        public bool IsActive { get; set; } = true;

        public DateTime CreatedAt { get; set; } = DateTime.UtcNow;

        public DateTime UpdatedAt { get; set; } = DateTime.UtcNow;
    }
}
