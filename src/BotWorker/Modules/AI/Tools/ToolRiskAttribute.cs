using System;

namespace BotWorker.Modules.AI.Tools
{
    /// <summary>
    /// 工具风险等级
    /// </summary>
    public enum ToolRiskLevel
    {
        /// <summary>
        /// 只读、无副作用 (文档检索、代码读取)
        /// </summary>
        Low,

        /// <summary>
        /// 有限副作用 (生成 diff、写草稿)
        /// </summary>
        Medium,

        /// <summary>
        /// 影响系统状态 (合并代码、发布配置)
        /// </summary>
        High
    }

    /// <summary>
    /// 用于标记工具函数及其风险等级
    /// </summary>
    [AttributeUsage(AttributeTargets.Method, AllowMultiple = false)]
    public class ToolRiskAttribute : Attribute
    {
        public ToolRiskLevel RiskLevel { get; }
        public string Description { get; }

        public ToolRiskAttribute(ToolRiskLevel riskLevel, string description = "")
        {
            RiskLevel = riskLevel;
            Description = description;
        }
    }
}
