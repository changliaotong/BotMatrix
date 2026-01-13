using System.Collections.Generic;
using System.Threading.Tasks;
using BotWorker.Modules.AI.Models.Evolution;

namespace BotWorker.Modules.AI.Interfaces
{
    public interface ITaskStepRepository : IRepository<TaskStep, long>
    {
        Task<IEnumerable<TaskStep>> GetByTaskIdAsync(long taskId);
    }
}
