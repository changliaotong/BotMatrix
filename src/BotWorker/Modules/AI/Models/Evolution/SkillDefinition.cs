using System;
using System.ComponentModel.DataAnnotations.Schema;

namespace BotWorker.Modules.AI.Models.Evolution
{
    /// <summary>
    /// 技能/工具定义 (Evolution - Skill)
    /// 对应数据库表: ai_skill_definitions
    /// </summary>
    [Table("ai_skill_definitions")]
    public class SkillDefinition
    {
        public long Id { get; set; }

        /// <summary>
        /// 技能唯一标识，如 file_read
        /// </summary>
        public string SkillKey { get; set; } = string.Empty;

        public string Name { get; set; } = string.Empty;

        public string Description { get; set; } = string.Empty;

        /// <summary>
        /// 对应 Agent 输出的 action，如 READ
        /// </summary>
        public string ActionName { get; set; } = string.Empty;

        /// <summary>
        /// 参数 Schema (JSONB)
        /// </summary>
        public string ParameterSchema { get; set; } = "{}";

        /// <summary>
        /// 是否为内置技能（由代码实现）
        /// </summary>
        public bool IsBuiltin { get; set; } = true;

        /// <summary>
        /// 动态脚本内容 (可选)
        /// </summary>
        public string? ScriptContent { get; set; }

        public DateTime CreatedAt { get; set; } = DateTime.UtcNow;

        public DateTime UpdatedAt { get; set; } = DateTime.UtcNow;
    }
}
