using BotWorker.Infrastructure.Persistence.ORM;
using System.Reflection;

namespace BotWorker.Modules.Games
{
    #region Attributes & Config

    [AttributeUsage(AttributeTargets.Method)]
    public class PetCommandAttribute : Attribute
    {
        public string[] Aliases { get; }
        public string Description { get; }
        public int Order { get; }
        public PetCommandAttribute(string[] aliases, string description, int order = 0)
        {
            Aliases = aliases;
            Description = description;
            Order = order;
        }
    }

    public class PetConfig
    {
        public double ExpMultiplier { get; set; } = 1.0;
        public double HungerRate { get; set; } = 2.0;
        public double EnergyRecoveryRate { get; set; } = 5.0;
        public int MaxPetCount { get; set; } = 1;
        public string DefaultPetName { get; set; } = "小萌新";
        public double IntimacyGainRate { get; set; } = 1.0;
    }

    #endregion

    #region Domain Model

    public class Pet : MetaDataGuid<Pet>
    {
        public string UserId { get; set; } = string.Empty;
        public string Name { get; set; } = string.Empty;
        public PetType Type { get; set; } = PetType.Cat;
        public PetPersonality Personality { get; set; } = PetPersonality.Ordinary;
        public PetState CurrentState { get; set; } = PetState.Idle;
        public DateTime StateEndTime { get; set; } = DateTime.MinValue;

        public DateTime AdoptTime { get; set; }
        public DateTime LastUpdateTime { get; set; }

        public double Health { get; set; } = 100;
        public double Hunger { get; set; } = 0;
        public double Happiness { get; set; } = 100;
        public double Energy { get; set; } = 100;
        public double Intimacy { get; set; } = 0; // 亲密度
        public int Gold { get; set; } = 100; // 金币
        public int Level { get; set; } = 1;
        public double Experience { get; set; } = 0;

        [DbIgnore] public double ExperienceToNextLevel => 100 * Math.Pow(Level, 1.2);
        [DbIgnore] public int Age => (DateTime.Now - AdoptTime).Days;
        [DbIgnore] public string PersonalityName => Personality switch
        {
            PetPersonality.Ordinary => "平凡",
            PetPersonality.Energetic => "精力充沛",
            PetPersonality.Lazy => "懒散",
            PetPersonality.Aggressive => "好斗",
            PetPersonality.Gentle => "温柔",
            _ => "未知"
        };

        [DbIgnore] public List<string> Events { get; } = new();

        public override string TableName => "UserPets";
        public override string KeyField => "Id";

        public static async Task<Pet?> GetByUserIdAsync(string userId)
        {
            return (await QueryWhere("UserId = @p1", SqlParams(("@p1", userId)))).FirstOrDefault();
        }

        public async Task UpdateStateByTimeAsync(PetConfig config)
        {
            var now = DateTime.Now;
            var hours = (now - LastUpdateTime).TotalHours;
            if (hours < 0.01) return;

            // 检查状态是否结束
            if (CurrentState != PetState.Idle && now >= StateEndTime)
            {
                if (CurrentState == PetState.Adventuring)
                {
                    Events.Add("🌟 冒险归来：你的宠物在野外发现了一些好东西！");
                    await PetInventory.AddItemAsync(UserId, "food_meat", 1);
                }
                else if (CurrentState == PetState.Working)
                {
                    Events.Add("💰 打工结束：你的宠物辛勤劳动，带回了酬劳。");
                }
                CurrentState = PetState.Idle;
                StateEndTime = DateTime.MinValue;
            }

            // 性格对衰减的影响
            double hungerMod = 1.0, energyMod = 1.0, happinessMod = 1.0;
            switch (Personality)
            {
                case PetPersonality.Energetic: energyMod = 0.8; hungerMod = 1.2; break;
                case PetPersonality.Lazy: energyMod = 1.2; hungerMod = 0.8; happinessMod = 0.5; break;
                case PetPersonality.Aggressive: happinessMod = 1.5; break;
                case PetPersonality.Gentle: happinessMod = 0.8; break;
            }

            Hunger = Math.Min(Hunger + hours * config.HungerRate * hungerMod, 100);
            
            if (CurrentState == PetState.Resting)
                Energy = Math.Min(Energy + hours * config.EnergyRecoveryRate * 2 * energyMod, 100);
            else
                Energy = Math.Max(Energy - hours * 2 * energyMod, 0);

            Happiness = Math.Max(Happiness - hours * 1.5 * happinessMod, 0);

            if (Hunger > 80) Health = Math.Max(Health - (Hunger - 80) * 0.1 * hours, 0);
            if (Energy < 10) Health = Math.Max(Health - (10 - Energy) * 0.05 * hours, 0);

            // 随机事件触发 (每小时约10%概率)
            if (hours > 0.5 && Random.Shared.NextDouble() < 0.1 * hours)
            {
                TriggerRandomEvent();
            }

            LastUpdateTime = now;
            await UpdateAsync();
        }

        private void TriggerRandomEvent()
        {
            var events = new[]
            {
                "🌈 你的宠物在草地上发现了一枚闪亮的硬币！",
                "🦋 你的宠物追逐蝴蝶时摔了一跤，但看起来很高兴。",
                "📦 你的宠物在门口捡到了一个包裹，里面居然有食物！",
                "🐱 你的宠物和邻居家的猫打了一架，受了点轻伤。"
            };
            var evt = events[Random.Shared.Next(events.Length)];
            Events.Add(evt);
            if (evt.Contains("硬币")) Gold += 10;
            if (evt.Contains("包裹")) Events.Add("(自动获得：小面包)"); // 逻辑简化，实际可加物品
            if (evt.Contains("轻伤")) Health -= 5;
        }

        public void Feed(double value)
        {
            Hunger = Math.Max(Hunger - value, 0);
            Health = Math.Min(Health + 2, 100);
            Intimacy = Math.Min(Intimacy + 1, 1000);
        }

        public void Play(double fun, double expMul)
        {
            Happiness = Math.Min(Happiness + fun, 100);
            Energy = Math.Max(Energy - 15, 0);
            Intimacy = Math.Min(Intimacy + 2, 1000);
            GainExp(fun * 2 * expMul);
        }

        public void GainExp(double exp)
        {
            Experience += exp;
            while (Experience >= ExperienceToNextLevel)
            {
                Experience -= ExperienceToNextLevel;
                Level++;
            }
        }
    }

    public class PetInventory : MetaDataGuid<PetInventory>
    {
        public string UserId { get; set; } = string.Empty;
        public string ItemId { get; set; } = string.Empty;
        public new int Count { get; set; } = 0;

        public override string TableName => "UserPetInventory";
        public override string KeyField => "Id";

        public static async Task<List<PetInventory>> GetByUserAsync(string userId)
        {
            return await QueryWhere("UserId = @p1 AND Count > 0", SqlParams(("@p1", userId)));
        }

        public static async Task AddItemAsync(string userId, string itemId, int count)
        {
            var item = (await QueryWhere("UserId = @p1 AND ItemId = @p2", SqlParams(("@p1", userId), ("@p2", itemId)))).FirstOrDefault();
            if (item == null)
            {
                item = new PetInventory { UserId = userId, ItemId = itemId, Count = count };
                await item.InsertAsync();
            }
            else
            {
                item.Count += count;
                await item.UpdateAsync();
            }
        }
    }

    public class PetItem
    {
        public string Id { get; set; } = string.Empty;
        public string Name { get; set; } = string.Empty;
        public string Description { get; set; } = string.Empty;
        public int Price { get; set; }
        public Action<Pet>? Effect { get; set; }

        public static readonly Dictionary<string, PetItem> All = new()
        {
            ["food_bread"] = new PetItem { Id = "food_bread", Name = "小面包", Description = "恢复20点饱食度", Price = 10, Effect = p => p.Feed(20) },
            ["food_meat"] = new PetItem { Id = "food_meat", Name = "美味大肉块", Description = "恢复50点饱食度", Price = 30, Effect = p => p.Feed(50) },
            ["toy_ball"] = new PetItem { Id = "toy_ball", Name = "逗猫球", Description = "恢复30点快乐度", Price = 20, Effect = p => p.Play(30, 1.0) },
            ["med_bandage"] = new PetItem { Id = "med_bandage", Name = "绷带", Description = "恢复20点健康值", Price = 25, Effect = p => p.Health = Math.Min(100, p.Health + 20) },
            ["exp_book"] = new PetItem { Id = "exp_book", Name = "宠物百科", Description = "增加100点经验值", Price = 50, Effect = p => p.GainExp(100) }
        };
    }

    public enum PetType { Cat, Dog, Bird, Slime, Dragon }
    public enum PetPersonality { Ordinary, Energetic, Lazy, Aggressive, Gentle }
    public enum PetState { Idle, Resting, Working, Adventuring }

    #endregion
}