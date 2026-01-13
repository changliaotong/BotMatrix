using System;
using System.Collections.Generic;
using System.Threading.Tasks;
using BotWorker.Modules.AI.Models.Evolution;

namespace BotWorker.Modules.AI.Interfaces
{
    public interface ITaskRecordRepository : IRepository<TaskRecord, long>
    {
        Task<TaskRecord?> GetByExecutionIdAsync(Guid executionId);
        Task<IEnumerable<TaskRecord>> GetByAssigneeIdAsync(long assigneeId);
        Task<bool> UpdateStatusAsync(long id, string status, int progress);
    }
}
