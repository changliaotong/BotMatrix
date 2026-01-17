using System;
using System.Threading.Tasks;
using BotWorker.Common;
using BotWorker.Domain.Interfaces;
using BotWorker.Modules.Plugins;

namespace BotWorker.Modules.Games
{
    [BotPlugin(
        Id = "game.simple",
        Name = "基础游戏集",
        Version = "1.0.0",
        Author = "Matrix",
        Description = "包含抢楼、打飞机、打地鼠、打群主等基础趣味互动游戏",
        Category = "Games"
    )]
    public class SimpleGamePlugin : IPlugin
    {
        private readonly ISimpleGameService _gameService;

        public SimpleGamePlugin(ISimpleGameService gameService)
        {
            _gameService = gameService;
        }

        public async Task InitAsync(IRobot robot)
        {
            await robot.RegisterSkillAsync(new SkillCapability
            {
                Name = "基础互动游戏",
                Commands = new[] { "抢楼", "打飞机", "打地鼠", "打群主", "抢救群主", "爱群主", "群主最伟大", "群主最伟大了", "我爱群主" },
                Description = "包含抢楼、打飞机、打地鼠、打群主、抢救群主、爱群主等趣味互动"
            }, HandleSimpleGameAsync);
        }

        public Task StopAsync() => Task.CompletedTask;

        private async Task<string> HandleSimpleGameAsync(IPluginContext ctx, string[] args)
        {
            var userId = long.Parse(ctx.UserId);
            var cmd = ctx.RawMessage.Trim();

            // 尝试从问答库获取回复，实现动态内容配置
            if (ctx is PluginContext pluginCtx && pluginCtx.Event is Infrastructure.Communication.OneBot.BotMessageEvent botMsgEvent)
            {
                var botMsg = botMsgEvent.BotMessage;
                var qaRes = await botMsg.GetQaAnswerAsync(cmd);
                if (!string.IsNullOrEmpty(qaRes))
                {
                    return qaRes;
                }
            }

            // 兜底硬编码逻辑
            if (cmd.Contains("抢楼")) return _gameService.RobBuilding(userId);
            if (cmd.Contains("打飞机")) return _gameService.DaFeiji(userId);
            if (cmd.Contains("打地鼠")) return _gameService.DaDishu(userId);
            if (cmd.Contains("打群主")) return _gameService.DaQunzhu(userId);
            if (cmd.Contains("抢救群主")) return _gameService.QiangjiuQunzhu(userId);
            if (cmd.Contains("爱群主") || cmd.Contains("群主最伟大") || cmd.Contains("我爱群主"))
                return _gameService.AiQunzhu(userId);

            return string.Empty;
        }
    }
}
