using BotWorker.Core.Interfaces;

namespace BotWorker.Bots.Models.Story
{
    public class StoryMetadata : IModuleMetadata
    {
        public string Name => "Story";
        public string Version => "1.0";
        public string Author => "光辉";
        public string Description => "章节式剧情推进系统";

        public IEnumerable<string> OptionalModules => ["UserProfile"];
    }

}
