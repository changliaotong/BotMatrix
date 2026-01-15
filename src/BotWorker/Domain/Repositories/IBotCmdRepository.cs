using System.Collections.Generic;
using System.Threading.Tasks;
using BotWorker.Domain.Entities;

namespace BotWorker.Domain.Repositories
{
    public interface IBotCmdRepository : IBaseRepository<BotCmd>
    {
        Task<IEnumerable<string>> GetAllCommandNamesAsync();
        Task<string?> GetCmdNameAsync(string cmdText);
        Task<IEnumerable<string>> GetClosedCommandsAsync();
        Task<bool> IsCmdCloseAllAsync(string cmdName);
        Task<string> GetCmdTextAsync(string cmdName);
        Task EnsureCommandExistsAsync(string name, string text);
    }
}
