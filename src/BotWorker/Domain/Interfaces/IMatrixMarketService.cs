using BotWorker.Domain.Interfaces;
using BotWorker.Modules.Games;
using System.Collections.Generic;
using System.Threading.Tasks;

namespace BotWorker.Domain.Interfaces
{
    public interface IMatrixMarketService
    {
        Task<string> GetMarketDisplayAsync(string userId);
        Task<string> UnlockModuleAsync(IPluginContext ctx, string moduleName);
        Task<List<UserModuleAccess>> GetUserAccessAsync(string userId);
    }
}
