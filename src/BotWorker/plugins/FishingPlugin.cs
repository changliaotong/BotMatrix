using System.Threading.Tasks;
using BotWorker.Core.Plugin;

namespace BotWorker.Plugins
{
    public class FishingPlugin : IPlugin
    {
        public string Name => "Fishing";
        public string Description => "一个简单的钓鱼插件";

        public Task InitAsync(IRobot robot)
        {
            // 注册钓鱼相关技能
            return Task.CompletedTask;
        }
    }
}
