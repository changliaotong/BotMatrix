using System;

namespace BotWorker.Domain.Models
{
    /// <summary>
    /// 所有 BotMatrix 事件的基类
    /// </summary>
    public abstract class BaseEvent
    {
        public string EventId { get; set; } = Guid.NewGuid().ToString("N");
        public DateTime Timestamp { get; set; } = DateTime.Now;
    }

    /// <summary>
    /// 系统审计日志事件
    /// 用于记录系统内部的关键变更，可用于实时监控看板
    /// </summary>
    public class SystemAuditEvent : BaseEvent
    {
        public string Level { get; set; } = "Info"; // Info, Warning, Success, Critical
        public string Source { get; set; } = string.Empty; // 来源插件
        public string Message { get; set; } = string.Empty;
        public string TargetUser { get; set; } = string.Empty;
    }

    /// <summary>
    /// 积分交易事件
    /// 当 PointsService 发生任何资金变动时触发
    /// </summary>
    public class PointTransactionEvent : BaseEvent
    {
        public string UserId { get; set; } = string.Empty;
        public string AccountType { get; set; } = string.Empty; // User, System, Revenue
        public decimal Amount { get; set; }
        public string Description { get; set; } = string.Empty;
        public string TransactionType { get; set; } = string.Empty; // Income, Expense
    }

    /// <summary>
    /// 等级提升事件
    /// 当 EvolutionService 检测到用户升级时触发
    /// </summary>
    public class LevelUpEvent : BaseEvent
    {
        public string UserId { get; set; } = string.Empty;
        public int OldLevel { get; set; }
        public int NewLevel { get; set; }
        public string RankName { get; set; } = string.Empty;
    }

    /// <summary>
    /// 全服光环 (Global Buff) 类型
    /// </summary>
    public enum BuffType
    {
        ExperienceMultiplier, // 经验倍率
        PointsMultiplier,     // 积分倍率
        DropRateMultiplier    // 掉率倍率
    }

    /// <summary>
    /// 全服光环变更事件
    /// </summary>
    public class GlobalBuffEvent : BaseEvent
    {
        public BuffType Type { get; set; }
        public double Multiplier { get; set; }
        public DateTime ExpireTime { get; set; }
        public string Reason { get; set; } = string.Empty;
        public bool IsActive => DateTime.Now < ExpireTime;
    }

    /// <summary>
    /// 系统交互事件
    /// 用于追踪用户在系统中的各种行为
    /// </summary>
    public class SystemInteractionEvent : BaseEvent
    {
        public string UserId { get; set; } = string.Empty;
        public string InteractionType { get; set; } = string.Empty; // OpenMenu, UseCommand, etc.
        public string Details { get; set; } = string.Empty;
    }
}
