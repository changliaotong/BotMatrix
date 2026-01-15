using System.Collections.Generic;
using System.Threading.Tasks;
using BotWorker.Modules.AI.Tools;

namespace BotWorker.Modules.AI.Interfaces
{
    public interface IToolAuditRepository
    {
        Task<long> AddAsync(ToolAuditLog log);
        Task<bool> UpdateAsync(ToolAuditLog log);
        Task<ToolAuditLog?> GetByGuidAsync(string guid);
        Task<IEnumerable<ToolAuditLog>> GetPendingApprovalsAsync();
    }
}
