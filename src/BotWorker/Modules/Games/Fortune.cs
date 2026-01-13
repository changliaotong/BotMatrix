using BotWorker.Domain.Interfaces;
using System.Text;

namespace BotWorker.Modules.Games
{
    [BotPlugin(
        Id = "game.fortune",
        Name = "今日运势",
        Version = "1.1.0",
        Author = "Matrix",
        Description = "查看今日运势、幸运色和幸运数字",
        Category = "Games"
    )]
    public class FortunePlugin : IPlugin
    {
        public async Task InitAsync(IRobot robot)
        {
            // 注册技能
            await robot.RegisterSkillAsync(new SkillCapability
            {
                Name = "今日运势",
                Commands = ["运势", "今日运势", "fortune"],
                Description = "查看今日运势、幸运色和幸运数字"
            }, HandleFortuneAsync);
        }

        public Task StopAsync() => Task.CompletedTask;

        private async Task<string> HandleFortuneAsync(IPluginContext ctx, string[] args)
        {
            // 优先尝试从问答库获取“抽签”或“运势”相关回复
            if (ctx is PluginContext pluginCtx && pluginCtx.Event is Infrastructure.Communication.OneBot.BotMessageEvent botMsgEvent)
            {
                var botMsg = botMsgEvent.BotMessage;
                var cmd = ctx.RawMessage.Trim();
                var qaRes = await botMsg.GetQaAnswerAsync(cmd);
                if (!string.IsNullOrEmpty(qaRes))
                {
                    return qaRes;
                }
            }

            var fortune = Fortune.GenerateFortune(ctx.UserId);
            return await Task.FromResult(Fortune.Format(fortune));
        }
    }

    public class Fortune
    {
        private static readonly string[] Colors = { "珊瑚红", "天空蓝", "墨绿色", "靛青", "浅紫", "鹅黄", "藏青", "象牙白", "奶油色", "玫瑰金" };
        private static readonly int[] LuckyNumbers = { 1, 3, 5, 6, 7, 8, 9 };
        private static readonly string[] Directions = { "正东", "正西", "正南", "正北", "东南", "西北", "东北", "西南" };
        private static readonly string[] Taboos = {"避免与上级争论", "避免久坐久看手机", "切忌冲动消费", "勿轻信他人承诺", "忌讳外出远行", "今日不宜开始新计划", 
                                                   "避免熬夜", "小心交通安全", "远离是非之地", "少说多做"};

        public static async Task<DailyFortune> GenerateFortuneAsync(string qq)
        {
            return await Task.Run(() => GenerateFortune(qq));
        }

        public static DailyFortune GenerateFortune(string qq)
        {
            int seed = (qq + DateTime.Today.ToString("yyyyMMdd")).GetHashCode();
            Random rng = new(seed);

            var fortune = new DailyFortune
            {
                Date = DateTime.Today,
                Love = rng.Next(44, 100),
                Wealth = rng.Next(44, 100),
                Career = rng.Next(44, 100),
                Health = rng.Next(44, 100),
                Color = Colors[rng.Next(Colors.Length)],
                LuckyNumber = LuckyNumbers[rng.Next(LuckyNumbers.Length)],
                Direction = Directions[rng.Next(Directions.Length)],
                Taboo = Taboos[rng.Next(Taboos.Length)]
            };

            fortune.Overall = (fortune.Love + fortune.Wealth + fortune.Career + fortune.Health) / 4;
            fortune.Comment = GetComment(fortune.Overall);

            return fortune;
        }

        private static string GetComment(int score)
        {
            if (score >= 90) return "鸿运当头，万事大吉";
            if (score >= 70) return "顺风顺水，小有收获";
            if (score >= 50) return "平平稳稳，按部就班";
            if (score >= 30) return "小心应对，略有波折";
            return "事与愿违，宜静不宜动";
        }

        public static string Format(DailyFortune fortune)
        {
            return $"🔮 今日运势（{fortune.Date:MM月dd日}）\n" +
                $"🌟 综合运势：{fortune.Overall} / 100\n" +
                $"✨ 福运评价：{fortune.Comment}\n" +
                $"❤️ 爱情运势：{fortune.Love}\n" +
                $"💰 财富运势：{fortune.Wealth}\n" +
                $"📚 事业运势：{fortune.Career}\n" +
                $"💪 健康运势：{fortune.Health}\n" +
                $"🎨 幸运颜色：{fortune.Color}\n" +
                $"🔢 幸运数字：{fortune.LuckyNumber}\n" +
                $"🧭 幸运方向：{fortune.Direction}\n" +
                $"🙅‍♂️ 禁忌事项：{fortune.Taboo}\n";
        }
    }

    public class DailyFortune
    {
        public DateTime Date { get; set; }
        public int Overall { get; set; }
        public int Love { get; set; }
        public int Wealth { get; set; }
        public int Career { get; set; }
        public int Health { get; set; }
        public string Color { get; set; } = string.Empty;
        public int LuckyNumber { get; set; }
        public string Direction { get; set; } = string.Empty;
        public string Taboo { get; set; } = string.Empty;
        public string Comment { get; set; } = string.Empty;
    }
}
