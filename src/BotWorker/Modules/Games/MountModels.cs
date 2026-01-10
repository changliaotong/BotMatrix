using BotWorker.Infrastructure.Persistence.ORM;
using System.Text;

namespace BotWorker.Modules.Games
{
    #region Enums & Config

    public enum MountRarity
    {
        Common = 0,    // 普通
        Rare = 1,      // 优秀
        Epic = 2,      // 史诗
        Legendary = 3, // 传说
        Mythic = 4     // 神话
    }

    public enum MountStatus
    {
        Idle,       // 休息中
        Riding,     // 骑乘中
        Training,   // 训练中
        Exploring   // 寻宝中
    }

    public class MountConfig
    {
        public double BaseExpRate { get; set; } = 1.0;
        public int MaxMountCount { get; set; } = 3;
        public double EvolutionSuccessRate { get; set; } = 0.8;
    }

    #endregion

    #region Domain Model

    public class Mount : MetaDataGuid<Mount>
    {
        public string UserId { get; set; } = string.Empty;
        public string Name { get; set; } = string.Empty;
        public string TemplateId { get; set; } = string.Empty;
        public MountRarity Rarity { get; set; } = MountRarity.Common;
        public MountStatus Status { get; set; } = MountStatus.Idle;

        // 基础属性
        public int Level { get; set; } = 1;
        public double Experience { get; set; } = 0;
        public int StarLevel { get; set; } = 0; // 星级（用于进阶）

        // 核心数值
        public double Speed { get; set; } = 10;     // 移动速度（影响冷却缩减）
        public double Power { get; set; } = 10;     // 力量（影响战斗/打工收益）
        public double Luck { get; set; } = 5;       // 幸运（影响掉落率）
        public double Stamina { get; set; } = 100;  // 耐力（骑乘消耗）

        public DateTime LastActionTime { get; set; } = DateTime.Now;
        public DateTime CreateTime { get; set; } = DateTime.Now;

        [DbIgnore] public double ExpToNextLevel => 50 * Math.Pow(Level, 1.5) * ((int)Rarity + 1);
        
        [DbIgnore] public string RarityName => Rarity switch
        {
            MountRarity.Common => "⚪ 普通",
            MountRarity.Rare => "🔵 优秀",
            MountRarity.Epic => "🟣 史诗",
            MountRarity.Legendary => "🟠 传说",
            MountRarity.Mythic => "🔴 神话",
            _ => "未知"
        };

        public override string TableName => "UserMounts";
        public override string KeyField => "Id";

        public static async Task<Mount?> GetActiveMountAsync(string userId)
        {
            return (await QueryWhere("UserId = @p1 AND Status = @p2", SqlParams(("@p1", userId), ("@p2", (int)MountStatus.Riding)))).FirstOrDefault();
        }

        public static async Task<List<Mount>> GetUserMountsAsync(string userId)
        {
            return await QueryWhere("UserId = @p1", SqlParams(("@p1", userId)));
        }

        public void GainExp(double exp)
        {
            Experience += exp;
            while (Experience >= ExpToNextLevel)
            {
                Experience -= ExpToNextLevel;
                Level++;
                // 升级提升属性
                Speed += 1 + (int)Rarity * 0.5;
                Power += 1 + (int)Rarity * 0.5;
                Luck += 0.5 + (int)Rarity * 0.2;
            }
        }
    }

    #endregion

    #region Templates

    public class MountTemplate
    {
        public string Id { get; set; } = string.Empty;
        public string Name { get; set; } = string.Empty;
        public MountRarity Rarity { get; set; }
        public string AsciiArt { get; set; } = string.Empty;
        public string Description { get; set; } = string.Empty;
        public double BaseSpeed { get; set; }
        public double BasePower { get; set; }
        public double BaseLuck { get; set; }

        [DbIgnore] public string RarityName => Rarity switch
        {
            MountRarity.Common => "⚪ 普通",
            MountRarity.Rare => "🔵 优秀",
            MountRarity.Epic => "🟣 史诗",
            MountRarity.Legendary => "🟠 传说",
            MountRarity.Mythic => "🔴 神话",
            _ => "未知"
        };

        public static readonly Dictionary<string, MountTemplate> All = new()
        {
            ["m_horse"] = new MountTemplate 
            { 
                Id = "m_horse", Name = "汗血宝马", Rarity = MountRarity.Common, 
                BaseSpeed = 15, BasePower = 10, BaseLuck = 2,
                AsciiArt = "  _\\ \\ \n ( - )_ \n  | |  \\ \n  |_|  |_|",
                Description = "一匹普普通通但忠诚稳健的马。" 
            },
            ["m_wolf"] = new MountTemplate 
            { 
                Id = "m_wolf", Name = "疾风苍狼", Rarity = MountRarity.Rare, 
                BaseSpeed = 25, BasePower = 15, BaseLuck = 5,
                AsciiArt = " /\\__/\\ \n( >.< ) \n )   (  \n( /  \\ )",
                Description = "穿梭在森林中的掠食者，速度极快。" 
            },
            ["m_dragon"] = new MountTemplate 
            { 
                Id = "m_dragon", Name = "裂空座", Rarity = MountRarity.Legendary, 
                BaseSpeed = 50, BasePower = 100, BaseLuck = 20,
                AsciiArt = "  <>_  \n <___> \n  | |  \n  ^ ^  ",
                Description = "传说中能撕裂天空的神龙。" 
            }
        };
    }

    #endregion
}
