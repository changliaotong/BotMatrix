using BotWorker.Domain.Interfaces;
using BotWorker.Domain.Repositories;
using System.Text;

namespace BotWorker.Modules.Games
{
    [BotPlugin(
        Id = "game.baby.v2",
        Name = "育儿系统",
        Version = "1.0.0",
        Author = "Matrix",
        Description = "开启育儿之旅：领养、培养、互动、成长",
        Category = "Games"
    )]
    public class BabyService : IPlugin
    {
        private readonly IBabyRepository _babyRepo;
        private readonly IBabyEventRepository _eventRepo;
        private readonly IBabyConfigRepository _configRepo;
        private readonly IAchievementService _achievementService;

        public BabyService(
            IBabyRepository babyRepo, 
            IBabyEventRepository eventRepo, 
            IBabyConfigRepository configRepo,
            IAchievementService achievementService)
        {
            _babyRepo = babyRepo;
            _eventRepo = eventRepo;
            _configRepo = configRepo;
            _achievementService = achievementService;
        }

        public List<Intent> Intents => [
            new() { Name = "宝宝系统", Keywords = ["我的宝宝", "领养宝宝", "宝宝学习", "宝宝打工", "宝宝互动", "宝宝改名"] }
        ];

        public async Task StopAsync() => await Task.CompletedTask;

        public async Task InitAsync(IRobot robot)
        {
            await EnsureTablesCreatedAsync();
            await robot.RegisterSkillAsync(new SkillCapability
            {
                Name = "育儿系统",
                Commands = ["领养宝宝", "我的宝宝", "宝宝改名", "宝宝学习", "宝宝打工", "宝宝互动", "宝宝商城", "购买", "拐卖宝宝说明", "开启宝宝系统", "关闭宝宝系统", "抛弃宝宝"],
                Description = "【领养宝宝】开始育儿；【我的宝宝】查看状态；【宝宝商城】购买用品；【宝宝学习】增加成长"
            }, HandleCommandAsync);
        }

        private async Task EnsureTablesCreatedAsync()
        {
            await _babyRepo.EnsureTableCreatedAsync();
            await _eventRepo.EnsureTableCreatedAsync();
            await _configRepo.EnsureTableCreatedAsync();
        }

        private async Task<string> HandleCommandAsync(IPluginContext ctx, string[] args)
        {
            var config = await _configRepo.GetAsync();
            var cmd = ctx.RawMessage.Trim().Split(' ')[0];

            // 管理员指令不受系统开关限制
            if (cmd == "开启宝宝系统" || cmd == "关闭宝宝系统" || cmd == "抛弃宝宝")
            {
                return await HandleAdminCommandAsync(ctx, cmd, args);
            }

            if (!config.IsEnabled) return "⚠️ 育儿系统当前已关闭。";

            // 自动检查每日成长和生日
            var updateMsg = await CheckDailyUpdateAsync(ctx);

            var res = cmd switch
            {
                "领养宝宝" => await AdoptBabyAsync(ctx, args),
                "我的宝宝" => await GetBabyStatusAsync(ctx),
                "宝宝改名" => await RenameBabyAsync(ctx, args),
                "宝宝学习" => await BabyLearnAsync(ctx),
                "宝宝打工" => await BabyWorkAsync(ctx),
                "宝宝互动" => await BabyInteractAsync(ctx),
                "宝宝商城" => await GetBabyMallAsync(ctx),
                "购买" => await BuyBabyItemAsync(ctx, args),
                "拐卖宝宝说明" => GetBabyHelp(ctx),
                _ => "未知育儿指令"
            };

            return string.IsNullOrEmpty(updateMsg) ? res : $"{updateMsg}\n\n{res}";
        }

        private async Task<string> CheckDailyUpdateAsync(IPluginContext ctx)
        {
            var baby = await _babyRepo.GetByUserIdAsync(ctx.UserId);
            if (baby == null) return string.Empty;

            var now = DateTime.Now;
            var sb = new StringBuilder();

            // 1. 每日自动成长
            if (baby.LastDailyUpdate.Date < now.Date)
            {
                baby.GrowthValue += 50;
                baby.LastDailyUpdate = now;
                await UpdateBabyGrowthAsync(baby);
                await _babyRepo.UpdateEntityAsync(baby);
                sb.AppendLine($"☀️ 新的一天，【{baby.Name}】 自动成长了！(成长值+50)");
            }

            // 2. 生日系统 (周年生日)
            if (baby.Birthday.Month == now.Month && baby.Birthday.Day == now.Day && baby.LastDailyUpdate.Year < now.Year)
            {
                var age = now.Year - baby.Birthday.Year;
                if (age > 0)
                {
                    baby.Points += 500; // 生日奖励 500 积分
                    baby.GrowthValue += 200; // 生日奖励 200 成长值
                    await UpdateBabyGrowthAsync(baby);
                    await _babyRepo.UpdateEntityAsync(baby);
                    sb.AppendLine($"🎂 哇！今天是 【{baby.Name}】 的 {age} 岁生日！");
                    sb.AppendLine($"🎁 收到系统赠送的生日大礼包：积分+500，成长值+200！");
                }
            }

            return sb.ToString().Trim();
        }

        private async Task<string> HandleAdminCommandAsync(IPluginContext ctx, string cmd, string[] args)
        {
            var botId = long.Parse(ctx.BotId);
            var userId = long.Parse(ctx.UserId);
            if (botId != userId && !BotWorker.Domain.Entities.BotInfo.IsAdmin(botId, userId))
            {
                return "❌ 只有机器人主人或系统管理员可以执行此操作。";
            }

            var config = await _configRepo.GetAsync();
            switch (cmd)
            {
                case "开启宝宝系统":
                    config.IsEnabled = true;
                    config.UpdatedAt = DateTime.Now;
                    await _configRepo.UpdateEntityAsync(config);
                    return "✅ 育儿系统已开启。";
                case "关闭宝宝系统":
                    config.IsEnabled = false;
                    config.UpdatedAt = DateTime.Now;
                    await _configRepo.UpdateEntityAsync(config);
                    return "📴 育儿系统已关闭。";
                case "抛弃宝宝":
                    if (args.Length == 0) return "请输入要抛弃宝宝的用户QQ。";
                    var targetId = args[0].Replace("@", "").Trim();
                    var baby = await _babyRepo.GetByUserIdAsync(targetId);
                    if (baby == null) return "该用户没有宝宝。";
                    baby.Status = "abandoned";
                    baby.UpdatedAt = DateTime.Now;
                    await _babyRepo.UpdateEntityAsync(baby);
                    return $"🚮 已强制抛弃用户 【{targetId}】 的宝宝 【{baby.Name}】。";
                default:
                    return "未知管理指令";
            }
        }

        private async Task<string> AdoptBabyAsync(IPluginContext ctx, string[] args)
        {
            var existing = await _babyRepo.GetByUserIdAsync(ctx.UserId);
            if (existing != null) return $"你已经有一个名为 {existing.Name} 的宝宝了。";

            var name = args.Length > 0 ? args[0] : "小宝贝";
            var baby = new Baby { UserId = ctx.UserId, Name = name };
            await _babyRepo.InsertAsync(baby);

            await _eventRepo.InsertAsync(new BabyEvent { BabyId = baby.Id, EventType = "adopt", Content = "降临到这个世界" });

            // 上报成就
            _ = AchievementPlugin.ReportMetricAsync(ctx.UserId, "baby.adopt_count", 1);

            return $"👶 恭喜！你的宝宝 【{name}】 降临了！快去照顾TA吧。";
        }

        private async Task<string> GetBabyStatusAsync(IPluginContext ctx)
        {
            var baby = await _babyRepo.GetByUserIdAsync(ctx.UserId);
            if (baby == null) return "你还没有宝宝，发送【领养宝宝】来获得一个吧。";

            var sb = new StringBuilder();
            sb.AppendLine($"👶 【{baby.Name}】 的成长记录");
            sb.AppendLine("━━━━━━━━━━━━━━");
            sb.AppendLine($"🎂 生日: {baby.Birthday:yyyy-MM-dd}");
            sb.AppendLine($"🌟 等级: Lv.{baby.Level}");
            sb.AppendLine($"📈 成长值: {baby.GrowthValue}");
            sb.AppendLine($"🕒 成长天数: {baby.DaysOld}天");
            sb.AppendLine($"💰 积分: {baby.Points}");
            sb.AppendLine("━━━━━━━━━━━━━━");
            return sb.ToString();
        }

        private async Task<string> RenameBabyAsync(IPluginContext ctx, string[] args)
        {
            if (args.Length == 0) return "宝宝要叫什么名字呢？";
            var baby = await _babyRepo.GetByUserIdAsync(ctx.UserId);
            if (baby == null) return "你还没有宝宝。";

            baby.Name = args[0];
            await _babyRepo.UpdateEntityAsync(baby);
            return $"📝 好的，以后宝宝就叫 【{baby.Name}】 啦。";
        }

        private async Task<string> BabyLearnAsync(IPluginContext ctx)
        {
            var baby = await _babyRepo.GetByUserIdAsync(ctx.UserId);
            if (baby == null) return "你还没有宝宝。";

            baby.GrowthValue += 100;
            await UpdateBabyGrowthAsync(baby);
            await _babyRepo.UpdateEntityAsync(baby);

            await _eventRepo.InsertAsync(new BabyEvent { BabyId = baby.Id, EventType = "learn", Content = "学习了新知识，成长值+100" });
            return $"📚 【{baby.Name}】 正在认真学习，看起来变聪明了！(成长+100)";
        }

        private async Task<string> BabyWorkAsync(IPluginContext ctx)
        {
            var baby = await _babyRepo.GetByUserIdAsync(ctx.UserId);
            if (baby == null) return "你还没有宝宝。";

            if (baby.DaysOld < 30) return $"⚠️ 【{baby.Name}】 还太小了，需要满 30 天（当前 {baby.DaysOld} 天）才能出去打工哦。";

            baby.GrowthValue += 150;
            baby.Points += 50;
            await UpdateBabyGrowthAsync(baby);
            await _babyRepo.UpdateEntityAsync(baby);
            await _eventRepo.InsertAsync(new BabyEvent { BabyId = baby.Id, EventType = "work", Content = "帮爸爸妈妈干活，成长值+150，获得50积分" });
            return $"💪 【{baby.Name}】 真懂事，在帮爸爸妈妈干活呢！(成长+150, 积分+50)";
        }

        private async Task<string> BabyInteractAsync(IPluginContext ctx)
        {
            var baby = await _babyRepo.GetByUserIdAsync(ctx.UserId);
            if (baby == null) return "你还没有宝宝。";

            baby.GrowthValue += 50;
            await UpdateBabyGrowthAsync(baby);
            await _babyRepo.UpdateEntityAsync(baby);
            return $"🥰 你抱了抱 【{baby.Name}】，宝宝开心地笑了。(成长+50)";
        }

        private async Task UpdateBabyGrowthAsync(Baby baby)
        {
            var config = await _configRepo.GetAsync();
            // 1000成长值增加1天年龄
            if (baby.GrowthValue >= config.GrowthRate)
            {
                var days = baby.GrowthValue / config.GrowthRate;
                baby.DaysOld += days;
                baby.GrowthValue %= config.GrowthRate;
            }

            // 每30天年龄提升1级
            baby.Level = 1 + (baby.DaysOld / 30);
            baby.UpdatedAt = DateTime.Now;
        }

        private async Task<string> GetBabyMallAsync(IPluginContext ctx)
        {
            var baby = await _babyRepo.GetByUserIdAsync(ctx.UserId);
            if (baby == null) return "你还没有宝宝，无法进入商城。";

            var sb = new StringBuilder();
            sb.AppendLine("🏪 【宝宝商城】");
            sb.AppendLine("━━━━━━━━━━━━━━");
            sb.AppendLine("1. 奶瓶 (50积分) - 增加100成长值");
            sb.AppendLine("2. 玩具车 (100积分) - 增加200成长值");
            sb.AppendLine("3. 故事书 (150积分) - 增加300成长值");
            sb.AppendLine("4. 新衣服 (200积分) - 增加400成长值");
            sb.AppendLine("━━━━━━━━━━━━━━");
            sb.AppendLine($"💰 当前积分: {baby.Points}");
            sb.AppendLine("💡 发送【购买+编号】即可购买。");
            return sb.ToString();
        }

        private async Task<string> BuyBabyItemAsync(IPluginContext ctx, string[] args)
        {
            if (args.Length == 0) return "请输入要购买的商品编号。";
            var baby = await _babyRepo.GetByUserIdAsync(ctx.UserId);
            if (baby == null) return "你还没有宝宝。";

            var itemNo = args[0];
            var (cost, growth, name) = itemNo switch
            {
                "1" => (50, 100, "奶瓶"),
                "2" => (100, 200, "玩具车"),
                "3" => (150, 300, "故事书"),
                "4" => (200, 400, "新衣服"),
                _ => (0, 0, "")
            };

            if (cost == 0) return "❌ 商品编号不存在。";
            if (baby.Points < cost) return $"❌ 积分不足，购买 {name} 需要 {cost} 积分，你当前只有 {baby.Points} 积分。";

            baby.Points -= cost;
            baby.GrowthValue += growth;
            await UpdateBabyGrowthAsync(baby);
            await _babyRepo.UpdateEntityAsync(baby);

            await _eventRepo.InsertAsync(new BabyEvent { BabyId = baby.Id, EventType = "buy", Content = $"购买了 {name}，成长值+{growth}" });
            return $"🛍️ 购买成功！宝宝使用了 【{name}】，(成长+{growth}，积分-{cost})。";
        }

        private string GetBabyHelp(IPluginContext ctx)
        {
            var sb = new StringBuilder();
            sb.AppendLine("📖 【育儿系统使用规范】");
            sb.AppendLine("━━━━━━━━━━━━━━");
            sb.AppendLine("1. 每位用户只能领养一个宝宝。");
            sb.AppendLine("2. 通过学习、打工、互动可获得成长值。");
            sb.AppendLine("3. 严禁通过作弊手段刷成长值，一经发现将由超级管理员【抛弃宝宝】。");
            sb.AppendLine("4. 宝宝打工可以获得积分，积分可在商城购买用品。");
            sb.AppendLine("5. 名字长度需在2-10个字符之间。");
            sb.AppendLine("━━━━━━━━━━━━━━");
            sb.AppendLine("💡 提示：合理安排宝宝的成长计划，TA会带给你更多惊喜。");
            return sb.ToString();
        }
    }
}
