using BotWorker.Infrastructure.Persistence.ORM;
using System;
using System.Collections.Generic;
using System.Threading.Tasks;

namespace BotWorker.Modules.Games
{
    /// <summary>
    /// 账户类型
    /// </summary>
    public enum AccountType
    {
        User = 0,           // 普通用户账户
        SystemReserve = 1,  // 系统储备金账户 (发行积分)
        SystemRevenue = 2,  // 系统收入账户 (回收积分)
        Merchant = 3,       // 商家账户 (如游戏插件)
    }

    /// <summary>
    /// 积分账户
    /// </summary>
    public class PointAccount : MetaDataGuid<PointAccount>
    {
        public string AccountId { get; set; } = string.Empty; // UserId 或 系统预定义ID
        public string AccountName { get; set; } = string.Empty;
        public AccountType Type { get; set; }
        public long Balance { get; set; } // 当前余额
        public DateTime LastUpdateTime { get; set; } = DateTime.Now;

        public override string TableName => "PointAccounts";
        public override string KeyField => "Id";

        public static async Task<PointAccount?> GetByAccountIdAsync(string accountId)
        {
            return await GetSingleAsync("WHERE AccountId = @AccountId", new { AccountId = accountId });
        }
    }

    /// <summary>
    /// 积分流水 (复式记账分录)
    /// </summary>
    public class PointLedger : MetaDataGuid<PointLedger>
    {
        public string TransactionId { get; set; } = string.Empty; // 交易流水号 (关联借贷双方)
        
        // 借方信息 (积分流入方)
        public string DebitAccountId { get; set; } = string.Empty;
        public string DebitAccountName { get; set; } = string.Empty;
        
        // 贷方信息 (积分流出方)
        public string CreditAccountId { get; set; } = string.Empty;
        public string CreditAccountName { get; set; } = string.Empty;

        public long Amount { get; set; } // 交易金额
        public string Description { get; set; } = string.Empty; // 交易摘要 (分录说明)
        public DateTime TransactionTime { get; set; } = DateTime.Now;

        public override string TableName => "PointLedgers";
        public override string KeyField => "Id";

        public static async Task<List<PointLedger>> GetAccountHistoryAsync(string accountId, int limit = 20)
        {
            return (await QueryAsync($"WHERE DebitAccountId = @Id OR CreditAccountId = @Id ORDER BY TransactionTime DESC", new { Id = accountId })).ToList();
        }
    }
}
