using System;
using System.ComponentModel.DataAnnotations.Schema;

namespace BotWorker.Modules.AI.Models.Billing
{
    /// <summary>
    /// AI 账户/钱包
    /// 对应数据库表: ai_wallets
    /// </summary>
    [Table("ai_wallets")]
    public class Wallet
    {
        public long Id { get; set; }

        /// <summary>
        /// 关联用户或组织 ID
        /// </summary>
        public long OwnerId { get; set; }

        /// <summary>
        /// 余额
        /// </summary>
        public decimal Balance { get; set; } = 0.0000m;

        public string Currency { get; set; } = "CNY";

        /// <summary>
        /// 冻结余额 (如租赁中)
        /// </summary>
        public decimal FrozenBalance { get; set; } = 0.0000m;

        /// <summary>
        /// 累计支出
        /// </summary>
        public decimal TotalSpent { get; set; } = 0.0000m;

        /// <summary>
        /// 配置 (JSONB): 预警线、自动充值等
        /// </summary>
        public string Config { get; set; } = "{}";
        public DateTime CreatedAt { get; set; } = DateTime.UtcNow;
        public DateTime UpdatedAt { get; set; } = DateTime.UtcNow;
    }
}
