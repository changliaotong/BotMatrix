using System.Collections.Generic;
using System.Threading.Tasks;
using BotWorker.Modules.AI.Models.Billing;

namespace BotWorker.Modules.AI.Interfaces
{
    public interface ILeaseResourceRepository : IRepository<LeaseResource, long>
    {
        Task<IEnumerable<LeaseResource>> GetAvailableResourcesAsync(string? type = null);
        Task<IEnumerable<LeaseResource>> GetByProviderIdAsync(long providerId);
    }
}
