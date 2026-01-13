using System.Collections.Generic;
using System.Threading.Tasks;
using BotWorker.Modules.AI.Models.Evolution;

namespace BotWorker.Modules.AI.Interfaces
{
    public interface IJobDefinitionRepository : IRepository<JobDefinition, long>
    {
        Task<JobDefinition?> GetByKeyAsync(string jobKey);
        Task<IEnumerable<JobDefinition>> GetActiveJobsAsync();
    }
}
