using BotWorker.Domain.Entities;
using BotWorker.Common.Extensions;
using BotWorker.Domain.Interfaces;
using BotWorker.Domain.Models.BotMessages;
using System.Threading.Tasks;
using System;
using System.Collections.Generic;
using System.Linq;
using Microsoft.Extensions.DependencyInjection;
using BotWorker.Domain.Repositories;
using Dapper.Contrib.Extensions;

namespace BotWorker.Modules.Games
{
    [BotPlugin(
        Id = "game.fishing.v2",
        Name = "新版钓鱼王",
        Version = "2.0.0",
        Author = "Matrix",
        Description = "深度钓鱼模拟：多场景探索、装备强化、鱼种图鉴、实时交易",
        Category = "Games"
    )]
    public class FishingPlugin : IPlugin
    {
        public List<Intent> Intents => [
            new() { Name = "钓鱼", Keywords = ["钓鱼", "钓鱼状态"] },
            new() { Name = "抛竿", Keywords = ["抛竿"] },
            new() { Name = "收竿", Keywords = ["收竿"] },
            new() { Name = "鱼篓", Keywords = ["鱼篓"] },
            new() { Name = "卖鱼", Keywords = ["卖鱼"] },
            new() { Name = "钓鱼商店", Keywords = ["钓鱼商店"] },
            new() { Name = "升级鱼竿", Keywords = ["升级鱼竿"] }
        ];

        private IRobot? _robot;
        public async Task InitAsync(IRobot robot)
        {
            _robot = robot;
            // 确保数据库表已创建
            await Fishing.EnsureTablesCreatedAsync();

            await robot.RegisterSkillAsync(new SkillCapability
            {
                Name = "新版钓鱼",
                Commands = ["钓鱼", "抛竿", "收竿", "鱼篓", "卖鱼", "钓鱼商店", "升级鱼竿", "钓鱼状态"],
                Description = "【钓鱼】查看当前状态；【抛竿】开始钓鱼；【收竿】看看收获；【鱼篓】查看战利品；【卖鱼】换取金币"
            }, HandleFishingAsync);
        }

        public async Task StopAsync() => await Task.CompletedTask;

        private async Task<string> HandleFishingAsync(IPluginContext ctx, string[] args)
        {
            var userId = long.Parse(ctx.UserId);
            var cmd = ctx.RawMessage.Trim().Split(' ')[0];

            try
            {
                return cmd switch
                {
                    "钓鱼" or "钓鱼状态" => await Fishing.GetStatusAsync(userId, ctx.User?.Name ?? "钓鱼佬"),
                    "抛竿" => await Fishing.CastAsync(userId),
                    "收竿" => await Fishing.ReelInAsync(userId),
                    "鱼篓" => await Fishing.GetBagAsync(userId),
                    "卖鱼" => await Fishing.SellFishAsync(userId),
                    "钓鱼商店" => await Fishing.GetShopAsync(userId),
                    "升级鱼竿" => await Fishing.UpgradeRodAsync(userId),
                    _ => "未知钓鱼指令"
                };
            }
            catch (Exception ex)
            {
                return $"❌ 钓鱼组件故障：{ex.Message}";
            }
        }
    }

    #region 数据实体

    [Table("fishing_user")]
    public class FishingUser
    {
        private static IFishingUserRepository Repository => 
            BotMessage.ServiceProvider?.GetRequiredService<IFishingUserRepository>() 
            ?? throw new InvalidOperationException("IFishingUserRepository not registered");

        [ExplicitKey]
        public long UserId { get; set; }
        public int Level { get; set; } = 1;
        public long Exp { get; set; } = 0;
        public long Gold { get; set; } = 0;
        public int RodLevel { get; set; } = 1;
        public int CurrentLocation { get; set; } = 0; // 0:淡水湖, 1:近海, 2:珊瑚礁, 3:深海
        public int State { get; set; } = 0; // 0:空闲, 1:钓鱼中
        public DateTime LastActionTime { get; set; } = DateTime.Now;
        public int WaitMinutes { get; set; } = 0;

        public static async Task<FishingUser> GetOrCreateAsync(long userId)
        {
            var user = await Repository.GetByIdAsync(userId);
            if (user == null)
            {
                user = new FishingUser { UserId = userId, Gold = 500, Level = 1, RodLevel = 1, LastActionTime = DateTime.Now };
                await Repository.AddAsync(user);
            }
            return user;
        }

        public static async Task UpdateStateAsync(long userId, int state, int waitMinutes)
        {
            await Repository.UpdateStateAsync(userId, state, waitMinutes);
        }

        public static async Task UpdateStateAsync(long userId, int state)
        {
            await Repository.UpdateStateAsync(userId, state);
        }

        public static async Task AddExpAndResetStateAsync(long userId, int exp)
        {
            await Repository.AddExpAndResetStateAsync(userId, exp);
        }

        public static async Task UpgradeRodAsync(long userId, long cost)
        {
            await Repository.UpgradeRodAsync(userId, cost);
        }

        public static async Task SellFishAsync(long userId, long totalGold)
        {
            await Repository.SellFishAsync(userId, totalGold);
        }
    }

    [Table("fishing_bag")]
    public class FishingBag
    {
        private static IFishingBagRepository Repository => 
            BotMessage.ServiceProvider?.GetRequiredService<IFishingBagRepository>() 
            ?? throw new InvalidOperationException("IFishingBagRepository not registered");

        [Key]
        public long Id { get; set; }
        public long UserId { get; set; }
        public string FishName { get; set; } = "";
        public double Weight { get; set; }
        public int Quality { get; set; } // 0:普通, 1:稀有, 2:史诗, 3:传说
        public long Value { get; set; }
        public DateTime CatchTime { get; set; } = DateTime.Now;

        public static async Task AddFishAsync(long userId, FishDef fish, double weight, long value)
        {
            await Repository.AddAsync(new FishingBag
            {
                UserId = userId,
                FishName = fish.Name,
                Weight = weight,
                Quality = (int)fish.Quality,
                Value = value,
                CatchTime = DateTime.Now
            });
        }

        public static async Task<IEnumerable<FishingBag>> GetByUserIdAsync(long userId, int limit)
        {
            return await Repository.GetByUserIdAsync(userId, limit);
        }

        public static async Task<IEnumerable<FishingBag>> GetAllByUserIdAsync(long userId)
        {
            return await Repository.GetAllByUserIdAsync(userId);
        }
    }

    #endregion

    #region 游戏逻辑引擎

    public enum FishQuality { Common = 0, Rare = 1, Epic = 2, Legendary = 3 }

    public class FishDef
    {
        public string Name { get; set; } = "";
        public FishQuality Quality { get; set; }
        public double MinWeight { get; set; }
        public double MaxWeight { get; set; }
        public long BaseValue { get; set; }
    }

    public class LocationDef
    {
        public string Name { get; set; } = "";
        public int MinLevel { get; set; }
        public List<FishDef> FishPool { get; set; } = new();
    }

    public static class Fishing
    {
        private static readonly List<LocationDef> Locations = new()
        {
            new LocationDef { Name = "淡水湖", MinLevel = 1, FishPool = new() {
                new FishDef { Name = "草鱼", Quality = FishQuality.Common, MinWeight = 0.5, MaxWeight = 5.0, BaseValue = 10 },
                new FishDef { Name = "鲤鱼", Quality = FishQuality.Common, MinWeight = 1.0, MaxWeight = 8.0, BaseValue = 15 },
                new FishDef { Name = "金色锦鲤", Quality = FishQuality.Rare, MinWeight = 2.0, MaxWeight = 10.0, BaseValue = 100 },
                new FishDef { Name = "湖中剑", Quality = FishQuality.Epic, MinWeight = 50.0, MaxWeight = 50.0, BaseValue = 1000 }
            }},
            new LocationDef { Name = "近海", MinLevel = 5, FishPool = new() {
                new FishDef { Name = "黄鱼", Quality = FishQuality.Common, MinWeight = 0.3, MaxWeight = 2.0, BaseValue = 30 },
                new FishDef { Name = "带鱼", Quality = FishQuality.Common, MinWeight = 0.5, MaxWeight = 3.0, BaseValue = 45 },
                new FishDef { Name = "真鲷", Quality = FishQuality.Rare, MinWeight = 1.0, MaxWeight = 15.0, BaseValue = 200 },
                new FishDef { Name = "大白鲨", Quality = FishQuality.Legendary, MinWeight = 500.0, MaxWeight = 2000.0, BaseValue = 5000 }
            }},
            new LocationDef { Name = "深海", MinLevel = 15, FishPool = new() {
                new FishDef { Name = "金枪鱼", Quality = FishQuality.Rare, MinWeight = 20.0, MaxWeight = 200.0, BaseValue = 800 },
                new FishDef { Name = "旗鱼", Quality = FishQuality.Epic, MinWeight = 100.0, MaxWeight = 500.0, BaseValue = 2500 },
                new FishDef { Name = "克苏鲁之眼", Quality = FishQuality.Legendary, MinWeight = 1000.0, MaxWeight = 1000.0, BaseValue = 50000 }
            }}
        };

        // 兼容旧版 HotCmdMessage 调用
        public static async Task<string> GetFishing(long groupId, string groupName, long userId, string name, string cmdName, string cmdPara)
        {
            return await GetStatusAsync(userId, name);
        }

        public static string GetBuyTools(long selfId, long groupId, string groupName, long userId, string name, string cmdName, string cmdPara, string cmdPara2)
        {
            return GetShopAsync(userId).GetAwaiter().GetResult();
        }

        public static async Task EnsureTablesCreatedAsync()
        {
            // await FishingUser.EnsureTableCreatedAsync();
            // await FishingBag.EnsureTableCreatedAsync();
            await Task.CompletedTask;
        }

        public static async Task<string> GetStatusAsync(long userId, string nickname)
        {
            var user = await FishingUser.GetOrCreateAsync(userId);
            var loc = Locations[user.CurrentLocation];
            var stateStr = user.State == 1 ? "🎣 正在垂钓中... (输入 收竿/收杆 看看收获)" : "💤 闲逛中 (输入 抛竿 开始钓鱼)";
            
            return $"【钓鱼执照】\n" +
                   $"等级：Lv.{user.Level} (XP: {user.Exp})\n" +
                   $"金币：{user.Gold} 💰\n" +
                   $"鱼竿：{user.RodLevel}级 (最大承重: {user.RodLevel * 10}kg)\n" +
                   $"当前位置：{loc.Name}\n" +
                   $"当前状态：{stateStr}";
        }

        public static async Task<string> CastAsync(long userId)
        {
            var user = await FishingUser.GetOrCreateAsync(userId);
            if (user.State == 1) return "你已经在钓鱼了，耐心一点！";

            int wait = new Random().Next(1, 4); // 1-3分钟
            await FishingUser.UpdateStateAsync(userId, 1, wait);
            
            return $"✅ 成功抛竿到 {Locations[user.CurrentLocation].Name}！\n静静等待鱼儿上钩吧...";
        }

        public static async Task<string> ReelInAsync(long userId)
        {
            var user = await FishingUser.GetOrCreateAsync(userId);
            if (user.State == 0) return "你还没抛竿呢，收什么竿？";

            var diff = (DateTime.Now - user.LastActionTime).TotalMinutes;
            if (diff < user.WaitMinutes)
            {
                await FishingUser.UpdateStateAsync(userId, 0);
                return "💨 哎呀，收竿太快，鱼被惊走了！";
            }

            // 成功捕获逻辑
            var loc = Locations[user.CurrentLocation];
            var random = new Random();
            var fish = loc.FishPool[random.Next(loc.FishPool.Count)];
            
            // 随机重量
            double weight = Math.Round(random.NextDouble() * (fish.MaxWeight - fish.MinWeight) + fish.MinWeight, 2);
            long value = (long)(fish.BaseValue * (weight / fish.MinWeight));

            // 检查鱼竿承重
            double maxWeight = user.RodLevel * 10.0;
            if (weight > maxWeight)
            {
                await FishingUser.UpdateStateAsync(userId, 0);
                return $"💔 糟糕！钓到了一头巨物({fish.Name} {weight}kg)，但是鱼竿承受不住，断线了！建议升级鱼竿。";
            }

            // 保存到鱼篓
            await FishingBag.AddFishAsync(userId, fish, weight, value);
            
            // 上报成就指标
            _ = AchievementPlugin.ReportMetricAsync(userId.ToString(), "fishing.catch_count", 1);

            // 增加经验
            int expGained = (int)fish.Quality * 10 + 5;
            await FishingUser.AddExpAndResetStateAsync(userId, expGained);

            string qualityStar = new string('⭐', (int)fish.Quality + 1);
            return $"🎊 恭喜！你收竿成功，钓到了：\n" +
                   $"🐟 品种：{fish.Name} {qualityStar}\n" +
                   $"⚖️ 重量：{weight} kg\n" +
                   $"💰 估值：{value} 金币\n" +
                   $"已放入鱼篓。经验 +{expGained}";
        }

        public static async Task<string> GetBagAsync(long userId)
        {
            var fishList = (await FishingBag.GetByUserIdAsync(userId, 1000)).ToList();
            if (fishList.Count == 0) return "你的鱼篓空空如也。";

            var sb = new System.Text.StringBuilder();
            sb.AppendLine($"🎒 {userId} 的鱼篓 ({fishList.Count} 条鱼)：");
            foreach (var f in fishList.Take(15))
            {
                string qualityIcon = new string('⭐', f.Quality + 1);
                sb.AppendLine($"{qualityIcon} {f.FishName} ({f.Weight:F1}kg) - {f.Value}金币 [{f.CatchTime:HH:mm}]");
            }
            if (fishList.Count > 15) sb.AppendLine($"... 还有 {fishList.Count - 15} 条鱼");

            long totalValue = fishList.Sum(f => f.Value);
            sb.AppendLine($"\n💰 总价值：{totalValue} 金币");
            return sb.ToString();
        }

        public static async Task<string> SellFishAsync(long userId)
        {
            var fishList = (await FishingBag.GetAllByUserIdAsync(userId)).ToList();
            if (fishList.Count == 0) return "没什么好卖的。";

            long totalGold = fishList.Sum(f => f.Value);
            
            try {
                await FishingUser.SellFishAsync(userId, totalGold);

                // 上报金币成就指标
                _ = AchievementPlugin.ReportMetricAsync(userId.ToString(), "fishing.total_gold", totalGold);

                return $"💰 所有的鱼已售出，获得 {totalGold} 金币！";
            } catch {
                return "交易失败，请稍后再试。";
            }
        }

        public static async Task<string> GetShopAsync(long userId)
        {
            var user = await FishingUser.GetOrCreateAsync(userId);
            long upgradeCost = user.RodLevel * 1000;
            return $"【钓鱼商店】\n" +
                   $"1. 升级鱼竿 (当前Lv.{user.RodLevel} -> Lv.{user.RodLevel + 1})\n" +
                   $"   效果：最大承重增加 10kg\n" +
                   $"   价格：{upgradeCost} 💰\n" +
                   $"发送【升级鱼竿】进行购买。";
        }

        public static async Task<string> UpgradeRodAsync(long userId)
        {
            var user = await FishingUser.GetOrCreateAsync(userId);
            long upgradeCost = user.RodLevel * 1000;
            if (user.Gold < upgradeCost) return $"你的金币不足！需要 {upgradeCost} 💰";

            await FishingUser.UpgradeRodAsync(userId, upgradeCost);
            return $"✅ 升级成功！当前鱼竿等级：Lv.{user.RodLevel + 1}，最大承重：{(user.RodLevel + 1) * 10}kg";
        }
    }

    #endregion
}
