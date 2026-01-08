using BotWorker.Domain.Interfaces;

namespace BotWorker.Modules.Gacha
{
    public class GachaMetadata : IModuleMetadata
    {
        public string Name => "Gacha";
        public string Version => "1.0";
        public string Author => "derlin";
        public string Description => "抽卡系统，带图鉴与概率池";
        public IEnumerable<string> OptionalModules => ["Credit"];
    }

}
