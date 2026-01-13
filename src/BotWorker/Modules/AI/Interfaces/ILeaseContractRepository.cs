using System.Collections.Generic;
using System.Threading.Tasks;
using BotWorker.Modules.AI.Models.Billing;

namespace BotWorker.Modules.AI.Interfaces
{
    public interface ILeaseContractRepository : IRepository<LeaseContract, long>
    {
        Task<IEnumerable<LeaseContract>> GetByTenantIdAsync(long tenantId);
        Task<IEnumerable<LeaseContract>> GetActiveContractsAsync();
    }
}
