using System.Threading.Tasks;
using BotWorker.Domain.Entities;

namespace BotWorker.Domain.Repositories
{
    public interface ISystemSettingRepository : IBaseRepository<SystemSetting>
    {
        Task<string> GetValueAsync(string key);
        Task<bool> GetBoolAsync(string key);
    }
}
