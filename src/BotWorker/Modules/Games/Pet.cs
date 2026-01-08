using BotWorker.Infrastructure.Persistence.ORM;

namespace BotWorker.Modules.Games
{
    // 基础宠物类
    public class Pet : MetaData<Pet>
    {
        public Guid Id { get; set; } = Guid.NewGuid();   // 宠物唯一ID
        public string Name { get; set; } = string.Empty; // 宠物名字
        public PetType Type { get; set; }                // 宠物类型
        public int Age { get; set; }                     // 宠物年龄（天）

        // 基础状态
        public double Health { get; private set; } = 100;    // 健康值0~100
        public double Hunger { get; private set; } = 0;      // 饥饿值0~100，越高越饿
        public double Happiness { get; private set; } = 100; // 快乐值0~100
        public double Energy { get; private set; } = 100;    // 精力值0~100
        public int Level { get; private set; } = 1;          // 等级
        public double Experience { get; private set; } = 0;  // 当前经验值
        public double ExperienceToNextLevel => 100 * Level;  // 升级所需经验

        public List<Skill> Skills { get; private set; } = new List<Skill>(); // 技能列表

        public override string TableName => throw new NotImplementedException();

        public override string KeyField => throw new NotImplementedException();

        public Pet()
        { 

        }


        // 构造函数
        public Pet(string name, PetType type, int age)
        {
            Name = name;
            Type = type;
            Age = age;
        }

        // 喂食
        public void Feed(double foodValue)
        {
            Hunger = Math.Max(Hunger - foodValue, 0);
            Health = Math.Min(Health + foodValue * 0.5, 100);
            Happiness = Math.Min(Happiness + foodValue * 0.3, 100);
            Console.WriteLine($"{Name} 被喂食了。健康值：{Health:F1}, 快乐值：{Happiness:F1}, 饥饿值：{Hunger:F1}");
        }

        // 玩耍
        public void Play(double funValue)
        {
            if (Energy <= 0)
            {
                Console.WriteLine($"{Name} 太累了，无法玩耍。");
                return;
            }

            Happiness = Math.Min(Happiness + funValue, 100);
            Energy = Math.Max(Energy - funValue * 0.5, 0);
            Hunger = Math.Min(Hunger + funValue * 0.3, 100);
            GainExperience(funValue * 2); // 玩耍增加经验
            Console.WriteLine($"{Name} 玩耍了。快乐值：{Happiness:F1}, 精力值：{Energy:F1}, 饥饿值：{Hunger:F1}");
        }

        // 宠物休息
        public void Rest(double hours)
        {
            Energy = Math.Min(Energy + hours * 10, 100);
            Health = Math.Min(Health + hours * 5, 100);
            Hunger = Math.Min(Hunger + hours * 2, 100);
            Console.WriteLine($"{Name} 休息了 {hours} 小时。精力值：{Energy:F1}, 健康值：{Health:F1}");
        }

        // 学习技能
        public void LearnSkill(Skill skill)
        {
            if (Skills.Any(s => s.Name == skill.Name))
            {
                Console.WriteLine($"{Name} 已经掌握技能 {skill.Name}。");
                return;
            }
            Skills.Add(skill);
            Console.WriteLine($"{Name} 学会了技能：{skill.Name}");
        }

        // 使用技能
        public void UseSkill(string skillName)
        {
            var skill = Skills.FirstOrDefault(s => s.Name == skillName);
            if (skill == null)
            {
                Console.WriteLine($"{Name} 不会技能 {skillName}。");
                return;
            }

            Happiness = Math.Min(Happiness + skill.Fun, 100);
            Energy = Math.Max(Energy - skill.EnergyCost, 0);
            Hunger = Math.Min(Hunger + skill.HungerCost, 100);
            GainExperience(skill.ExpGain);
            Console.WriteLine($"{Name} 使用技能 {skill.Name}。快乐值：{Happiness:F1}, 精力值：{Energy:F1}");
        }

        // 增加经验并升级
        private void GainExperience(double exp)
        {
            Experience += exp;
            while (Experience >= ExperienceToNextLevel)
            {
                Experience -= ExperienceToNextLevel;
                Level++;
                Health = 100;
                Energy = 100;
                Happiness = 100;
                Console.WriteLine($"{Name} 升级了！当前等级：{Level}");
            }
        }

        // 打印状态面板
        public void ShowStatus()
        {
            Console.WriteLine("===== 宠物状态面板 =====");
            Console.WriteLine($"名字: {Name}  类型: {Type}  等级: {Level}  经验: {Experience:F1}/{ExperienceToNextLevel}");
            Console.WriteLine($"年龄: {Age} 天");
            Console.WriteLine($"健康: {Health:F1}  饥饿: {Hunger:F1}  快乐: {Happiness:F1}  精力: {Energy:F1}");
            Console.WriteLine("技能列表:");
            if (Skills.Count == 0)
                Console.WriteLine("无技能");
            else
                foreach (var s in Skills)
                    Console.WriteLine($"- {s.Name} (Lv{s.Level})");
            Console.WriteLine("========================");
        }
    }

    // 宠物类型
    public enum PetType
    {
        Cat,
        Dog,
        Bird,
        Fish,
        Hamster,
        Dragon
    }

    // 宠物技能
    public class Skill
    {
        public string Name { get; set; }
        public string Description { get; set; }
        public int Level { get; set; } = 1;
        public double Fun { get; set; }         // 快乐增益
        public double EnergyCost { get; set; }  // 精力消耗
        public double HungerCost { get; set; }  // 玩技能饥饿增加
        public double ExpGain { get; set; }     // 使用技能获得经验

        public Skill(string name, string description, double fun = 10, double energyCost = 5, double hungerCost = 3, double expGain = 15)
        {
            Name = name;
            Description = description;
            Fun = fun;
            EnergyCost = energyCost;
            HungerCost = hungerCost;
            ExpGain = expGain;
        }
    }
}
