using System.Threading.Tasks;
using BotWorker.Common;
using BotWorker.Domain.Repositories;
using Dapper.Contrib.Extensions;
using Microsoft.Extensions.DependencyInjection;

namespace BotWorker.Domain.Entities
{
    [Table("SystemSetting")]
    public class SystemSetting
    {
        [ExplicitKey]
        public string Key { get; set; }
        public string Value { get; set; }

        private static ISystemSettingRepository Repo => GlobalConfig.ServiceProvider!.GetRequiredService<ISystemSettingRepository>();

        public static bool IsCloudLimited => GetBool("IsCloudLimited");

        public static bool IsPrefixNameProxy => GetBool("IsPrefixNameProxy");

        private static bool GetBool(string key) => Repo.GetBoolAsync(key).GetAwaiter().GetResult();
    }
}
