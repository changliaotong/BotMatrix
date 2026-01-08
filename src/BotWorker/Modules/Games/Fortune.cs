namespace BotWorker.Modules.Games
{
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
            return $"🔮 今日运势（{{农历月}}月{{农历日}}）\n" +
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
