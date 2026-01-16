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
            if (cmd.Contains("抢楼")) return SimpleGame.RobBuilding(userId);
            if (cmd.Contains("打飞机")) return SimpleGame.DaFeiji(userId);
            if (cmd.Contains("打地鼠")) return SimpleGame.DaDishu(userId);
            if (cmd.Contains("打群主")) return SimpleGame.DaQunzhu(userId);
            if (cmd.Contains("抢救群主")) return SimpleGame.QiangjiuQunzhu(userId);
            if (cmd.Contains("爱群主") || cmd.Contains("群主最伟大") || cmd.Contains("我爱群主"))
                return SimpleGame.AiQunzhu(userId);

            return string.Empty;
        }
    }

    internal static class SimpleGame
    {
        private static readonly Random _random = new();

        private static int RandomInt(int min, int max) => _random.Next(min, max + 1);

        /// <summary>
        /// 抢楼游戏      
        /// </summary>
        /// <param name="qq"></param>
        /// <returns></returns>
        public static string RobBuilding(long qq)
        {
            int i = RandomInt(1, 19);
            return i switch
            {
                1 => "恭喜你，抽到了xxx积分",
                2 => "恭喜你，抽到了禁言卡",
                4 => "恭喜你，抽到了xxx经验",
                5 => "恭喜你，抽到了xxx经验",
                16 => "恭喜你，抽到解禁卡",
                18 => "恭喜你，抽到紫币卡",
                19 => "恭喜你，抽到xxx金币",
                _ => "很遗憾，抢楼失败"
            };
        }

        public static string DaFeiji(long qq)
        {
            int i = RandomInt(1, 6);
            return i switch
            {
                1 => "拿着大炮打飞机，飞机跑掉了，损失xxx金币",
                2 => "拿着连环炮打飞机，一连打了好几架，获得xxx金币",
                3 => "左手只是辅助，右手才是关键，打飞机成功达到最高境界，舒服的享受，扣除纸巾费用xxx金币，获得社会威望50，享受期间不得离开，禁止游戏1分钟。",
                4 => "拿着手枪打飞机，飞机跑掉了，损失xxx金币",
                5 => "拿着火箭打飞机，一连下了好几架，获得金币，积分",
                _ => "打飞机失败"
            };
        }

        public static string DaDishu(long qq)
        {
            int i = RandomInt(1, 4);
            return i switch
            {
                1 => "打地鼠成功，获得xxx金币",
                2 => "打地鼠失败，损失xxx金币",
                3 => "打地鼠平手",
                _ => "打地鼠失败"
            };
        }

        public static string DaQunzhu(long qq)
        {
            int i = RandomInt(1, 4);
            return i switch
            {
                1 => "打群主成功，获得xxx金币",
                2 => "打群主失败，被群主反杀，损失xxx金币",
                3 => "打群主平手",
                _ => "打群主失败"
            };
        }

        public static string QiangjiuQunzhu(long qq)
        {
            int i = RandomInt(1, 4);
            return i switch
            {
                1 => "抢救群主成功，获得xxx金币",
                2 => "抢救群主失败，群主驾鹤西去，损失xxx金币",
                3 => "抢救群主平手",
                _ => "抢救群主失败"
            };
        }

        public static string AiQunzhu(long qq)
        {
            int i = RandomInt(1, 4);
            return i switch
            {
                1 => "爱群主成功，获得xxx金币",
                2 => "爱群主失败，群主不爱你，损失xxx金币",
                3 => "爱群主平手",
                _ => "爱群主失败"
            };
        }
    }
}
