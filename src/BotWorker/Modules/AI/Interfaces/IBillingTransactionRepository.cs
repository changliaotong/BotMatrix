using System.Collections.Generic;
using System.Threading.Tasks;
using BotWorker.Modules.AI.Models.Billing;

namespace BotWorker.Modules.AI.Interfaces
{
    public interface IBillingTransactionRepository : IRepository<BillingTransaction, long>
    {
        Task<IEnumerable<BillingTransaction>> GetByWalletIdAsync(long walletId);
    }
}
