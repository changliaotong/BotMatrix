using BotWorker.Domain.Interfaces;
using BotWorker.Modules.Plugins;
using System.Threading.Tasks;

namespace BotWorker.Plugins
{
    [BotPlugin(
        Id = "game.fishing",
        Name = "Fishing",
        Description = "一个简单的钓鱼插件",
        Version = "1.0.0",
        Author = "BotMatrix Team"
    )]
    public class FishingPlugin : IPlugin
    {
        public string Name => "Fishing";
        public string Description => "一个简单的钓鱼插件";

        public Task InitAsync(IRobot robot)
        {
            // 注册钓鱼相关技能
            return robot.RegisterSkillAsync(new SkillCapability
            {
                Name = "fish",
                Description = "钓鱼",
                Commands = new[] { "钓鱼" }
            }, async (ctx, args) => 
            {
                await ctx.ReplyAsync("🎣 你甩出了鱼竿...");
                await Task.Delay(1000);
                return "🐟 恭喜你钓到了一条小金鱼！";
            });
        }

        public async Task StopAsync() => await Task.CompletedTask;
    }
}
