using BotWorker.Domain.Entities;
using BotWorker.Domain.Interfaces;
using BotWorker.Domain.Repositories;
using Microsoft.Extensions.Logging;
using System;
using System.Collections.Generic;
using System.Linq;
using System.Threading.Tasks;

namespace BotWorker.Modules.Games
{
    public class FishingService : IFishingService
    {
        private readonly IFishingUserRepository _userRepo;
        private readonly IFishingBagRepository _bagRepo;
        private readonly IAchievementService _achievementService;
        private readonly ILogger<FishingService> _logger;

        public FishingService(
            IFishingUserRepository userRepo,
            IFishingBagRepository bagRepo,
            IAchievementService achievementService,
            ILogger<FishingService> logger)
        {
            _userRepo = userRepo;
            _bagRepo = bagRepo;
            _achievementService = achievementService;
            _logger = logger;
        }

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

        private async Task<FishingUser> GetOrCreateUserAsync(long userId)
        {
            var user = await _userRepo.GetByIdAsync(userId);
            if (user == null)
            {
                user = new FishingUser { UserId = userId, Gold = 500, Level = 1, RodLevel = 1, LastActionTime = DateTime.Now };
                await _userRepo.AddAsync(user);
            }
            return user;
        }

        public async Task<string> GetStatusAsync(long userId, string nickname)
        {
            var user = await GetOrCreateUserAsync(userId);
            var loc = Locations[user.CurrentLocation];
            var stateStr = user.State == 1 ? "🎣 正在垂钓中... (输入 收竿/收杆 看看收获)" : "💤 闲逛中 (输入 抛竿 开始钓鱼)";
            
            return $"【钓鱼执照】\n" +
                   $"等级：Lv.{user.Level} (XP: {user.Exp})\n" +
                   $"金币：{user.Gold} 💰\n" +
                   $"鱼竿：{user.RodLevel}级 (最大承重: {user.RodLevel * 10}kg)\n" +
                   $"当前位置：{loc.Name}\n" +
                   $"当前状态：{stateStr}";
        }

        public async Task<string> CastAsync(long userId)
        {
            var user = await GetOrCreateUserAsync(userId);
            if (user.State == 1) return "你已经在钓鱼了，耐心一点！";

            int wait = Random.Shared.Next(1, 4); // 1-3分钟
            await _userRepo.UpdateStateAsync(userId, 1, wait);
            
            return $"✅ 成功抛竿到 {Locations[user.CurrentLocation].Name}！\n静静等待鱼儿上钩吧...";
        }

        public async Task<string> ReelInAsync(long userId)
        {
            var user = await GetOrCreateUserAsync(userId);
            if (user.State == 0) return "你还没抛竿呢，收什么竿？";

            var diff = (DateTime.Now - user.LastActionTime).TotalMinutes;
            if (diff < user.WaitMinutes)
            {
                await _userRepo.UpdateStateAsync(userId, 0);
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
                await _userRepo.UpdateStateAsync(userId, 0);
                return $"💔 糟糕！钓到了一头巨物({fish.Name} {weight}kg)，但是鱼竿承受不住，断线了！建议升级鱼竿。";
            }

            // 保存到鱼篓
            await _bagRepo.AddAsync(new FishingBag
            {
                UserId = userId,
                FishName = fish.Name,
                Weight = weight,
                Quality = (int)fish.Quality,
                Value = value,
                CatchTime = DateTime.Now
            });
            
            // 上报成就指标
            _ = _achievementService.ReportMetricAsync(userId.ToString(), "fishing.catch_count", 1);

            // 增加经验
            int expGained = (int)fish.Quality * 10 + 5;
            await _userRepo.AddExpAndResetStateAsync(userId, expGained);

            string qualityStar = new string('⭐', (int)fish.Quality + 1);
            return $"🎊 恭喜！你收竿成功，钓到了：\n" +
                   $"🐟 品种：{fish.Name} {qualityStar}\n" +
                   $"⚖️ 重量：{weight} kg\n" +
                   $"💰 估值：{value} 金币\n" +
                   $"已放入鱼篓。经验 +{expGained}";
        }

        public async Task<string> GetBagAsync(long userId)
        {
            var fishList = (await _bagRepo.GetByUserIdAsync(userId, 1000)).ToList();
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

        public async Task<string> SellFishAsync(long userId)
        {
            var fishList = (await _bagRepo.GetAllByUserIdAsync(userId)).ToList();
            if (fishList.Count == 0) return "没什么好卖的。";

            long totalGold = fishList.Sum(f => f.Value);
            
            try {
                await _userRepo.SellFishAsync(userId, totalGold);

                // 上报金币成就指标
                _ = _achievementService.ReportMetricAsync(userId.ToString(), "fishing.total_gold", totalGold);

                return $"💰 所有的鱼已售出，获得 {totalGold} 金币！";
            } catch {
                return "交易失败，请稍后再试。";
            }
        }

        public async Task<string> GetShopAsync(long userId)
        {
            var user = await GetOrCreateUserAsync(userId);
            long upgradeCost = user.RodLevel * 1000;
            return $"【钓鱼商店】\n" +
                   $"1. 升级鱼竿 (当前Lv.{user.RodLevel} -> Lv.{user.RodLevel + 1})\n" +
                   $"   效果：最大承重增加 10kg\n" +
                   $"   价格：{upgradeCost} 💰\n" +
                   $"发送【升级鱼竿】进行购买。";
        }

        public async Task<string> UpgradeRodAsync(long userId)
        {
            var user = await GetOrCreateUserAsync(userId);
            long upgradeCost = user.RodLevel * 1000;
            if (user.Gold < upgradeCost) return $"你的金币不足！需要 {upgradeCost} 💰";

            await _userRepo.UpgradeRodAsync(userId, upgradeCost);
            return $"✅ 升级成功！当前鱼竿等级：Lv.{user.RodLevel + 1}，最大承重：{(user.RodLevel + 1) * 10}kg";
        }

        public async Task<string> HandleFishingAsync(long userId, string userName, string cmd)
        {
            try
            {
                return cmd switch
                {
                    "钓鱼" or "钓鱼状态" => await GetStatusAsync(userId, userName),
                    "抛竿" => await CastAsync(userId),
                    "收竿" => await ReelInAsync(userId),
                    "鱼篓" => await GetBagAsync(userId),
                    "卖鱼" => await SellFishAsync(userId),
                    "钓鱼商店" => await GetShopAsync(userId),
                    "升级鱼竿" => await UpgradeRodAsync(userId),
                    _ => "未知钓鱼指令"
                };
            }
            catch (Exception ex)
            {
                _logger.LogError(ex, "钓鱼指令处理失败: {Command}", cmd);
                return $"❌ 钓鱼组件故障：{ex.Message}";
            }
        }

        #region Helper Classes
        private class LocationDef
        {
            public string Name { get; set; } = string.Empty;
            public int MinLevel { get; set; }
            public List<FishDef> FishPool { get; set; } = new();
        }

        private class FishDef
        {
            public string Name { get; set; } = string.Empty;
            public FishQuality Quality { get; set; }
            public double MinWeight { get; set; }
            public double MaxWeight { get; set; }
            public int BaseValue { get; set; }
        }

        private enum FishQuality
        {
            Common = 0,
            Rare = 1,
            Epic = 2,
            Legendary = 3
        }
        #endregion
    }
}
