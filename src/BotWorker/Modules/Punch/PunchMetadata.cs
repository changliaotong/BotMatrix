using BotWorker.Core.Interfaces;

namespace BotWorker.Domain.Entities.Punch
{
    public class PunchMetadata : IModuleMetadata
    {
        public string Name => "Punch";
        public string Version => "1.0";
        public string Author => "derlin";
        public string Description => "用户可通过指令“揍群主”获得积分";

        public IEnumerable<string> RequiredModules => ["Credit"];
    }

}
