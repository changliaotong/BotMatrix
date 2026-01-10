using BotWorker.Infrastructure.Persistence.ORM;

namespace BotWorker.Domain.Entities
{
    public class SystemSetting : MetaData<SystemSetting>
    {
        public override string TableName => "SystemSetting";

        public override string KeyField => "Key";


        public static bool IsCloudLimited => GetBool("Value", "IsCloudLimited");

        public static bool IsPrefixNameProxy => GetBool("Value", "IsPrefixNameProxy");
    }
}
