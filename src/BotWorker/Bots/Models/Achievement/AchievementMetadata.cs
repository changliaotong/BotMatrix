using BotWorker.Core.Interfaces;

namespace BotWorker.Bots.Models.Achievement
{
    public class AchievementMetadata : IModuleMetadata
    {
        public string Name => "Achievement";
        public string Version => "1.0";
        public string Author => "derlin";
        public string Description => "用户成就系统，奖励积分";

        public IEnumerable<string> RequiredModules => ["Credit"];
    }
}
