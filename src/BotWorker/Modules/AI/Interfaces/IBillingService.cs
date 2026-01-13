using System.Threading.Tasks;
using BotWorker.Modules.AI.Models.Billing;

namespace BotWorker.Modules.AI.Interfaces
{
    public interface IBillingService
    {
        /// <summary>
        /// 获取或创建用户的钱包
        /// </summary>
        Task<Wallet> GetOrCreateWalletAsync(long ownerId);

        /// <summary>
        /// 检查余额是否足够
        /// </summary>
        Task<bool> HasSufficientBalanceAsync(long ownerId, decimal requiredAmount);

        /// <summary>
        /// 记录一次消费
        /// </summary>
        Task<bool> ConsumeAsync(long ownerId, decimal amount, long? relatedId = null, string? relatedType = null, string? remark = null);

        /// <summary>
        /// 充值
        /// </summary>
        Task<bool> RechargeAsync(long ownerId, decimal amount, string? remark = null);

        /// <summary>
        /// 检查用户是否有有效的算力资源租赁
        /// </summary>
        Task<bool> HasActiveLeaseAsync(long tenantId, string resourceType);
    }
}
