using BotWorker.Domain.Interfaces;

namespace BotWorker.Modules.JustKidding
{
    public class JustKiddingMetadata : IModuleMetadata
    {
        public string Name => "JustKidding";
        public string Version => "1.0";
        public string Author => "derlin";
        public string Description => "逗你玩。逗群里人玩不明觉厉";
        public IEnumerable<string> OptionalModules => ["Credit"];
    }
}
