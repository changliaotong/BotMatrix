using System.Collections.Concurrent;
using System.Collections.Generic;

namespace BotWorker.Services
{
    public interface II18nService
    {
        string GetString(string key, string lang = "zh-CN");
    }

    public class I18nService : II18nService
    {
        private readonly ConcurrentDictionary<string, Dictionary<string, string>> _translations = new();

        public string GetString(string key, string lang = "zh-CN")
        {
            if (_translations.TryGetValue(lang, out var dict) && dict.TryGetValue(key, out var value))
            {
                return value;
            }
            return key;
        }
    }
}


