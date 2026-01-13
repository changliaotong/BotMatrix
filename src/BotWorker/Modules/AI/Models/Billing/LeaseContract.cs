using System;
using System.ComponentModel.DataAnnotations.Schema;

namespace BotWorker.Modules.AI.Models.Billing
{
    /// <summary>
    /// 租赁合同/订单
    /// 对应数据库表: ai_lease_contracts
    /// </summary>
    [Table("ai_lease_contracts")]
    public class LeaseContract
    {
        public long Id { get; set; }

        public long ResourceId { get; set; }

        /// <summary>
        /// 承租人 ID
        /// </summary>
        public long TenantId { get; set; }

        public DateTime StartTime { get; set; }

        public DateTime? EndTime { get; set; }

        /// <summary>
        /// 状态: active, completed, terminated
        /// </summary>
        public string Status { get; set; } = "active";

        public bool AutoRenew { get; set; } = false;

        public decimal TotalPaid { get; set; } = 0.0000m;

        public string Config { get; set; } = "{}";

        public DateTime CreatedAt { get; set; } = DateTime.UtcNow;

        public DateTime UpdatedAt { get; set; } = DateTime.UtcNow;
    }
}
