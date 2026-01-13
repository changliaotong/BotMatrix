using System;
using System.ComponentModel.DataAnnotations.Schema;

namespace BotWorker.Modules.AI.Models.Billing
{
    /// <summary>
    /// 计费流水 (包括 Token 消耗和租赁费用)
    /// 对应数据库表: ai_billing_transactions
    /// </summary>
    [Table("ai_billing_transactions")]
    public class BillingTransaction
    {
        public long Id { get; set; }

        public long WalletId { get; set; }

        /// <summary>
        /// 类型: consume, recharge, refund, lease_fee
        /// </summary>
        public string Type { get; set; } = "consume";

        public decimal Amount { get; set; }

        /// <summary>
        /// 关联 ID (任务 ID 或 租赁合同 ID)
        /// </summary>
        public long? RelatedId { get; set; }

        /// <summary>
        /// 关联类型: task, lease, recharge
        /// </summary>
        public string? RelatedType { get; set; }

        public string? Remark { get; set; }
        public DateTime CreatedAt { get; set; } = DateTime.UtcNow;
        public DateTime UpdatedAt { get; set; } = DateTime.UtcNow;
    }
}
