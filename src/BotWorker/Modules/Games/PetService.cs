using BotWorker.Domain.Interfaces;
using Microsoft.Extensions.Logging;
using System.Text;

namespace BotWorker.Modules.Games
{
    [BotPlugin(
        Id = "game.pet.v2",
        Name = "宠物养成",
        Version = "2.0.0",
        Author = "Matrix",
        Description = "深度宠物养成系统：领养、打工、冒险、进化、多样互动",
        Category = "Games"
    )]
    public class PetService : IPlugin
    {
        private IRobot? _robot;
        private ILogger? _logger;
        private readonly PetConfig _config;

        public PetService() 
        {
            _config = new PetConfig();
        }

        public PetService(IRobot robot, ILogger logger, PetConfig config)
        {
            _robot = robot;
            _logger = logger;
            _config = config;
        }

        public List<Intent> Intents => [
            new() { Name = "领养宠物", Keywords = ["领养宠物", "adopt"] },
            new() { Name = "我的宠物", Keywords = ["我的宠物", "status", "pet"] },
            new() { Name = "喂食", Keywords = ["喂食", "feed"] }
        ];

        public async Task StopAsync() => await Task.CompletedTask;

        public async Task InitAsync(IRobot robot)
        {
            await EnsureTablesCreatedAsync();
            await robot.RegisterSkillAsync(new SkillCapability
            {
                Name = "宠物养成",
                Commands = ["领养宠物", "我的宠物", "喂食", "宠物状态", "宠物商店", "宠物互动", "宠物打工", "宠物冒险", "宠物休息", "宠物改名", "宠物排行榜"],
                Description = "【领养宠物】开启养宠之旅；【我的宠物】查看面板；【喂食】补充体力；【宠物打工/冒险】获取收益"
            }, HandlePetCommandAsync);
        }

        private async Task EnsureTablesCreatedAsync()
        {
            try
            {
                // 检查 UserPets 表
                var checkPet = await Pet.QueryScalarAsync<int>("SELECT COUNT(*) FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_NAME = 'UserPets'");
                if (checkPet == 0)
                {
                    var sql = BotWorker.Infrastructure.Utils.Schema.SchemaSynchronizer.GenerateCreateTableSql<Pet>();
                    await Pet.ExecAsync(sql);
                    Console.WriteLine("[Pet] Created table UserPets");
                }

                // 检查 PetInventory 表
                var checkInv = await PetInventory.QueryScalarAsync<int>("SELECT COUNT(*) FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_NAME = 'PetInventory'");
                if (checkInv == 0)
                {
                    var sql = BotWorker.Infrastructure.Utils.Schema.SchemaSynchronizer.GenerateCreateTableSql<PetInventory>();
                    await PetInventory.ExecAsync(sql);
                    Console.WriteLine("[Pet] Created table PetInventory");
                }
            }
            catch (Exception ex)
            {
                Console.WriteLine($"[PetService] 数据库表初始化失败: {ex.Message}");
            }
        }

        private async Task<string> HandlePetCommandAsync(IPluginContext ctx, string[] args)
        {
            var cmd = ctx.RawMessage.Trim().Split(' ')[0];
            return cmd switch
            {
                "领养宠物" or "adopt" => await AdoptAsync(ctx, args),
                "我的宠物" or "status" or "pet" or "宠物状态" => await GetStatusAsync(ctx, args),
                "喂食" or "feed" => await FeedAsync(ctx, args),
                "宠物商店" or "shop" => await ShopAsync(ctx, args),
                "购买" or "buy" => await BuyAsync(ctx, args),
                "宠物背包" or "bag" => await BagAsync(ctx, args),
                "打工" or "work" or "宠物打工" => await WorkAsync(ctx, args),
                "冒险" or "adventure" or "宠物冒险" => await AdventureAsync(ctx, args),
                "休息" or "rest" or "宠物休息" => await RestAsync(ctx, args),
                "互动" or "play" or "宠物互动" => await InteractAsync(ctx, args),
                "改名" or "rename" or "宠物改名" => await RenameAsync(ctx, args),
                "宠物排行" or "top" or "宠物排行榜" => await GetTopAsync(ctx, args),
                _ => "未知宠物指令"
            };
        }

        [PetCommand(["领养宠物", "adopt"], "开始领养你的第一个伙伴", 1)]
        public async Task<string> AdoptAsync(IPluginContext ctx, string[] args)
        {
            var existing = await Pet.GetByUserIdAsync(ctx.UserId);
            if (existing != null) return $"你已经有一只名为 {existing.Name} 的宠物了！";

            var name = args.Length > 0 ? args[0] : _config.DefaultPetName;
            var pet = new Pet
            {
                UserId = ctx.UserId,
                Name = name,
                Personality = (PetPersonality)Random.Shared.Next(0, 5),
                AdoptTime = DateTime.Now,
                LastUpdateTime = DateTime.Now
            };

            await pet.InsertAsync();

            // 上报成就
            _ = AchievementPlugin.ReportMetricAsync(ctx.UserId, "pet.adopt_count", 1);

            return $"🎊 领养成功！欢迎新成员 {name}！";
        }

        [PetCommand(["我的宠物", "status", "pet"], "查看宠物的详细状态面板", 2)]
        public async Task<string> GetStatusAsync(IPluginContext ctx, string[] args)
        {
            var pet = await Pet.GetByUserIdAsync(ctx.UserId);
            if (pet == null) return "你还没有宠物，快去【领养宠物】吧！";

            await pet.UpdateStateByTimeAsync(_config);

            var sb = new StringBuilder();
            sb.AppendLine(GetPetAscii(pet.Type));
            sb.AppendLine($"🐾 【{pet.Name}】的状态面板 ({pet.PersonalityName})");
            sb.AppendLine($"━━━━━━━━━━━━━━");
            sb.AppendLine($"⭐ 等级: {pet.Level} (EXP: {pet.Experience:F0}/{pet.ExperienceToNextLevel})");
            sb.AppendLine($"❤️ 健康: {RenderBar(pet.Health)} {pet.Health:F0}%");
            sb.AppendLine($"🍕 饱食: {RenderBar(100 - pet.Hunger)} {100 - pet.Hunger:F0}%");
            sb.AppendLine($"🎮 快乐: {RenderBar(pet.Happiness)} {pet.Happiness:F0}%");
            sb.AppendLine($"⚡ 精力: {RenderBar(pet.Energy)} {pet.Energy:F0}%");
            sb.AppendLine($"💞 亲密: {pet.Intimacy:F0} | 💰 金币: {pet.Gold}");
            sb.AppendLine($"🕒 状态: {RenderState(pet)}");
            sb.AppendLine($"📅 陪伴天数: {pet.Age}天");
            sb.AppendLine($"━━━━━━━━━━━━━━");

            if (pet.Events.Count > 0)
            {
                sb.AppendLine("📢 最近动态：");
                foreach (var evt in pet.Events) sb.AppendLine($"• {evt}");
                sb.AppendLine($"━━━━━━━━━━━━━━");
            }

            sb.Append($"💡 提示：{GetTip(pet)}");

            return sb.ToString();
        }

        private string GetPetAscii(PetType type)
        {
            return type switch
            {
                PetType.Cat => "  /\\_/\\\n ( o.o )\n  > ^ <",
                PetType.Dog => "  __      _\n /  \\____/ |\n <_  ____  |\n   \\/    \\/ ",
                PetType.Slime => "  _____\n /     \\\n(  o o  )\n \\_____/",
                PetType.Dragon => "  ^__^\n  (oo)\\_______\n  (__)\\       )\\/\\\n      ||----w |\n      ||     ||",
                _ => " (•‿•) "
            };
        }

        [PetCommand(["喂食", "feed"], "给宠物喂食（需消耗小面包或肉块）", 3)]
        public async Task<string> FeedAsync(IPluginContext ctx, string[] args)
        {
            var pet = await Pet.GetByUserIdAsync(ctx.UserId);
            if (pet == null) return "你还没有宠物。";

            var inv = await PetInventory.GetByUserAsync(ctx.UserId);
            var food = inv.FirstOrDefault(i => i.ItemId.StartsWith("food_"));
            if (food == null) return "你的背包里没有食物了，快去【宠物商店】看看吧！";

            if (!PetItem.All.TryGetValue(food.ItemId, out var item) || item == null) return "该食物项已失效。";
            if (pet == null) return "宠物不存在。";
            if (_config == null) return "宠物系统配置未加载。";
            await pet.UpdateStateByTimeAsync(_config);
            item.Effect?.Invoke(pet);
            food.Count--;
            await food.UpdateAsync();
            await pet.UpdateAsync();

            return $"🍖 你给 {pet.Name} 喂了 {item.Name}，{item.Description}。";
        }

        [PetCommand(["宠物商店", "shop"], "购买各种宠物道具", 6)]
        public async Task<string> ShopAsync(IPluginContext ctx, string[] args)
        {
            var sb = new StringBuilder();
            sb.AppendLine("🏪 【宠物商店】清单");
            sb.AppendLine("------------------");
            foreach (var item in PetItem.All.Values)
            {
                sb.AppendLine($"• {item.Name} ({item.Price}金币) - {item.Description}");
            }
            sb.AppendLine("------------------");
            sb.Append("使用【购买 [商品名]】进行购买");
            return sb.ToString();
        }

        [PetCommand(["购买", "buy"], "购买商店中的道具", 7)]
        public async Task<string> BuyAsync(IPluginContext ctx, string[] args)
        {
            if (args.Length == 0) return "请输入要购买的商品名称。";
            var pet = await Pet.GetByUserIdAsync(ctx.UserId);
            if (pet == null) return "你还没有宠物，买来也没法用。";

            var itemName = args[0];
            var item = PetItem.All.Values.FirstOrDefault(i => i.Name == itemName);
            if (item == null) return $"商店里没有名为 {itemName} 的商品。";

            if (pet.Gold < item.Price) return $"金币不足！你需要 {item.Price} 金币，但目前只有 {pet.Gold}。";

            pet.Gold -= item.Price;
            await pet.UpdateAsync();
            await PetInventory.AddItemAsync(ctx.UserId, item.Id, 1);

            return $"🛒 购买成功！获得了 {item.Name}，消耗了 {item.Price} 金币。";
        }

        [PetCommand(["宠物背包", "bag"], "查看你拥有的宠物道具", 10)]
        public async Task<string> BagAsync(IPluginContext ctx, string[] args)
        {
            var inv = await PetInventory.GetByUserAsync(ctx.UserId);
            if (inv.Count == 0) return "你的背包空空如也。";

            var sb = new StringBuilder();
            sb.AppendLine("🎒 【我的宠物背包】");
            sb.AppendLine("------------------");
            foreach (var pi in inv)
            {
                if (PetItem.All.TryGetValue(pi.ItemId, out var item))
                {
                    sb.AppendLine($"• {item.Name} x{pi.Count} - {item.Description}");
                }
            }
            sb.AppendLine("------------------");
            sb.Append("使用【喂食】会自动消耗食物类道具。");
            return sb.ToString();
        }

        [PetCommand(["打工", "work"], "派遣宠物打工赚取金币", 8)]
        public async Task<string> WorkAsync(IPluginContext ctx, string[] args)
        {
            return await ExecuteInteraction(ctx.UserId, p => {
                if (p.CurrentState != PetState.Idle) return $"{p.Name} 正在忙着呢，目前状态：{RenderState(p)}";
                if (p.Energy < 30) return $"{p.Name} 太累了，没法去打工。";
                
                p.CurrentState = PetState.Working;
                p.StateEndTime = DateTime.Now.AddHours(2);
                p.Energy -= 30;
                p.Gold += 50;
                p.GainExp(20);
                return $"💼 {p.Name} 去外面打工了，预计2小时后回来，将带回50金币。";
            });
        }

        [PetCommand(["宠物排行", "top"], "查看最强的宠物们", 11)]
        public async Task<string> GetTopAsync(IPluginContext ctx, string[] args)
        {
            var pets = (await Pet.QueryAsync("ORDER BY Level DESC, Experience DESC LIMIT 10", null)).ToList();
            if (pets.Count == 0) return "目前还没有宠物。";

            var sb = new StringBuilder();
            sb.AppendLine("🏆 【宠物等级排行榜】");
            sb.AppendLine("--------------------");
            for (int i = 0; i < pets.Count; i++)
            {
                var p = pets[i];
                sb.AppendLine($"{i + 1}. {p.Name} (Lv.{p.Level}) - 亲密:{p.Intimacy:F0}");
            }
            return sb.ToString();
        }

        [PetCommand(["冒险", "adventure"], "让宠物去冒险，可能带回稀有物品", 9)]
        public async Task<string> AdventureAsync(IPluginContext ctx, string[] args)
        {
            return await ExecuteInteraction(ctx.UserId, p => {
                if (p.CurrentState != PetState.Idle) return $"{p.Name} 正在忙着呢。";
                if (p.Energy < 50) return $"{p.Name} 精力不足，没法去冒险。";

                p.CurrentState = PetState.Adventuring;
                p.StateEndTime = DateTime.Now.AddHours(4);
                p.Energy -= 50;
                p.GainExp(100);
                return $"⚔️ {p.Name} 踏上了冒险之旅，预计4小时后归来。";
            });
        }

        [PetCommand(["休息", "rest"], "让宠物休息恢复精力", 5)]
        public async Task<string> RestAsync(IPluginContext ctx, string[] args)
        {
            return await ExecuteInteraction(ctx.UserId, p => {
                if (p.CurrentState != PetState.Idle) return $"{p.Name} 正在忙着呢。";
                p.CurrentState = PetState.Resting;
                p.StateEndTime = DateTime.Now.AddHours(1);
                return $"💤 {p.Name} 趴在垫子上睡着了，1小时后将恢复大量精力。";
            });
        }

        [PetCommand(["互动", "play"], "与宠物进行互动，增加亲密度和快乐", 4)]
        public async Task<string> InteractAsync(IPluginContext ctx, string[] args)
        {
            return await ExecuteInteraction(ctx.UserId, p => {
                if (p.CurrentState != PetState.Idle) return $"{p.Name} 正在忙着呢。";
                if (p.Energy < 10) return $"{p.Name} 太累了，不想理你。";
                
                p.Play(20, _config.ExpMultiplier);
                return $"✨ 你和 {p.Name} 玩了一会，它看起来开心多了！(亲密+2, 快乐+20)";
            });
        }

        [PetCommand(["改名", "rename"], "给宠物起一个新名字", 12)]
        public async Task<string> RenameAsync(IPluginContext ctx, string[] args)
        {
            if (args.Length == 0) return "请输入新的名字。";
            var pet = await Pet.GetByUserIdAsync(ctx.UserId);
            if (pet == null) return "你还没有宠物。";

            var oldName = pet.Name;
            pet.Name = args[0];
            await pet.UpdateAsync();
            return $"📝 改名成功！{oldName} 现在叫做 {pet.Name} 了。";
        }

        private string RenderState(Pet p)
        {
            if (p.CurrentState == PetState.Idle) return "闲逛中";
            var remaining = p.StateEndTime - DateTime.Now;
            var timeStr = remaining.TotalMinutes > 0 ? $" (剩余 {remaining.TotalMinutes:F0} 分钟)" : "";
            return p.CurrentState switch
            {
                PetState.Resting => "休息中" + timeStr,
                PetState.Working => "打工中" + timeStr,
                PetState.Adventuring => "冒险中" + timeStr,
                _ => "未知"
            };
        }

        private string GetTip(Pet p)
        {
            if (p.Hunger > 80) return "它看起来饿极了，快喂喂它吧！";
            if (p.Energy < 20) return "它看起来很疲惫，需要休息。";
            if (p.Happiness < 30) return "它看起来不太开心，陪它玩玩？";
            return "它今天心情不错！";
        }

        private async Task<string> ExecuteInteraction(string userId, Func<Pet, string> action)
        {
            var pet = await Pet.GetByUserIdAsync(userId);
            if (pet == null) return "你还没有宠物。";

            await pet.UpdateStateByTimeAsync(_config);
            var result = action(pet);
            await pet.UpdateAsync();

            // 统一上报宠物等级指标
            _ = AchievementPlugin.ReportMetricAsync(userId, "pet.max_level", pet.Level, true);

            return result;
        }

        private string RenderBar(double value)
        {
            const int length = 10;
            int filled = (int)Math.Clamp(value / 10, 0, length);
            return $"[{new string('■', filled).PadRight(length, '□')}]";
        }
    }
}