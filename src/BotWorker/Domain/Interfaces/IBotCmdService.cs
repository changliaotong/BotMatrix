using System.Collections.Generic;
using System.Threading.Tasks;

namespace BotWorker.Domain.Interfaces
{
    public interface IBotCmdService
    {
        Task<string> GetRegexCmdAsync();
        Task<string> GetCmdNameAsync(string cmdText);
        Task<string> GetClosedCmdAsync();
        Task<bool> IsClosedCmdAsync(long groupId, string message);
        Task<bool> IsCmdCloseAllAsync(string cmdName);
        Task<string> GetCmdTextAsync(string cmdName);
        Task EnsureCommandExistsAsync(string name, string text);
        Task<int> SetCmdCloseAllAsync(string cmdName, int isClose);
        void RegisterExtraCommands(IEnumerable<string> commands);
    }
}
