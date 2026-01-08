using BotWorker.Core.Interfaces;

namespace sz84.Bots.Models.Challenge
{
    public class ChallengeMetadata : IModuleMetadata
    {
        public string Name => "Challenge";
        public string Version => "1.0";
        public string Author => "derlin";
        public string Description => "单人挑战系统，含关卡推进、冷却与每日限制";
        public IEnumerable<string> OptionalModules => ["Credit"];
    }
}
