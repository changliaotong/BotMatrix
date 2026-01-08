using BotWorker.Core.Interfaces;

namespace sz84.Bots.Models.PvP
{
    public class PvPMetadata : IModuleMetadata
    {
        public string Name => "PvP";
        public string Version => "1.0";
        public string Author => "光辉";
        public string Description => "玩家对战系统，含段位与积分排行";
        public IEnumerable<string> RequiredModules => ["Credit"];
    }
}
