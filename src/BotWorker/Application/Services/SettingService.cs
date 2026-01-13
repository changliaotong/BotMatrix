using System.Collections.Concurrent;
using System.Threading.Tasks;

namespace BotWorker.Application.Services
{
    public interface ISettingService
    {
        Task<string?> GetSettingAsync(string key);
        Task SetSettingAsync(string key, string value);
    }

    public class SettingService : ISettingService
    {
        private readonly ConcurrentDictionary<string, string> _settings = new();

        public async Task<string?> GetSettingAsync(string key)
        {
            return await Task.FromResult(_settings.TryGetValue(key, out var value) ? value : null);
        }

        public async Task SetSettingAsync(string key, string value)
        {
            _settings[key] = value;
            await Task.CompletedTask;
        }
    }
}


