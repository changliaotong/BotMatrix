using System.Collections.Generic;
using System.Threading.Tasks;
using BotWorker.Modules.AI.Models.Evolution;

namespace BotWorker.Modules.AI.Interfaces
{
    public interface IEmployeeInstanceRepository : IRepository<EmployeeInstance, long>
    {
        Task<EmployeeInstance?> GetByEmployeeIdAsync(string employeeId);
        Task<IEnumerable<EmployeeInstance>> GetByBotIdAsync(string botId);
        Task<IEnumerable<EmployeeInstance>> GetByJobIdAsync(long jobId);
        Task<bool> UpdateStatusAsync(long id, string onlineStatus, string state);
    }
}
