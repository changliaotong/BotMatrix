using System;
using System.ComponentModel.DataAnnotations.Schema;

namespace BotWorker.Modules.AI.Models.Billing
{
    /// <summary>
    /// 算力租赁资源项
    /// 对应数据库表: ai_lease_resources
    /// </summary>
    [Table("ai_lease_resources")]
    public class LeaseResource
    {
        public long Id { get; set; }

        public string Name { get; set; } = string.Empty;

        /// <summary>
        /// 类型: gpu_worker, agent_instance, employee_service
        /// </summary>
        public string Type { get; set; } = "gpu_worker";

        public string? Description { get; set; }

        /// <summary>
        /// 资源提供者 (0 为系统，其它为用户 ID)
        /// </summary>
        public long ProviderId { get; set; }

        public decimal PricePerHour { get; set; }

        public string UnitName { get; set; } = "hour";

        public int MaxCapacity { get; set; } = 1;

        public int CurrentUsage { get; set; } = 0;

        /// <summary>
        /// 状态: available, busy, maintenance
        /// </summary>
        public string Status { get; set; } = "available";

        public bool IsActive { get; set; } = true;

        /// <summary>
        /// 规格配置 (JSONB): VRAM, CUDA cores 等
        /// </summary>
        public string Config { get; set; } = "{}";

        public DateTime CreatedAt { get; set; } = DateTime.UtcNow;

        public DateTime UpdatedAt { get; set; } = DateTime.UtcNow;
    }
}
