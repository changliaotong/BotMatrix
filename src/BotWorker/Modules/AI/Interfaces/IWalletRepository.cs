using System.Threading.Tasks;
using BotWorker.Modules.AI.Models.Billing;

namespace BotWorker.Modules.AI.Interfaces
{
    public interface IWalletRepository : IRepository<Wallet, long>
    {
        Task<Wallet?> GetByOwnerIdAsync(long ownerId);
        Task<bool> UpdateBalanceAsync(long ownerId, decimal amount, bool isFreeze = false);
    }
}
