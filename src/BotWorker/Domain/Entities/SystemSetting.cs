using System.Threading.Tasks;
using BotWorker.Common;
using BotWorker.Domain.Repositories;
using Dapper.Contrib.Extensions;
using Microsoft.Extensions.DependencyInjection;

namespace BotWorker.Domain.Entities
{
    [Table("system_setting")]
    public class SystemSetting
    {
        [ExplicitKey]
        public string Key { get; set; }
        public string Value { get; set; }

        public static async Task<bool> IsCloudLimitedAsync(BotMessage bm) => await GetBoolAsync(bm, "IsCloudLimited");

        public static async Task<bool> IsPrefixNameProxyAsync(BotMessage bm) => await GetBoolAsync(bm, "IsPrefixNameProxy");

        private static async Task<bool> GetBoolAsync(BotMessage bm, string key) => await bm.SystemSettingRepository.GetBoolAsync(key);
    }
}
