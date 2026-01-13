using BotWorker.Infrastructure.Persistence.ORM;
using System.Text;

namespace BotWorker.Modules.Games
{
    #region Enums & Config

    public enum VehicleRarity
    {
        Common = 0,    // 经济型
        Rare = 1,      // 舒适型
        Epic = 2,      // 豪华型
        Legendary = 3, // 超级跑车
        Mythic = 4     // 未来概念
    }

    public enum VehicleStatus
    {
        Idle,       // 停车中
        Driving,    // 驾驶中
        Repairing,  // 维修中
        Tuning      // 改装中
    }

    public class VehicleConfig
    {
        public double BaseFuelConsumption { get; set; } = 1.0;
        public int MaxVehicleCount { get; set; } = 3;
        public double TuningSuccessRate { get; set; } = 0.75;
    }

    #endregion

    #region Domain Model

    public class Vehicle : MetaDataGuid<Vehicle>
    {
        public string UserId { get; set; } = string.Empty;
        public string Name { get; set; } = string.Empty;
        public string TemplateId { get; set; } = string.Empty;
        public VehicleRarity Rarity { get; set; } = VehicleRarity.Common;
        public VehicleStatus Status { get; set; } = VehicleStatus.Idle;

        // 基础属性
        public int Level { get; set; } = 1;
        public double Experience { get; set; } = 0;
        public int ModificationLevel { get; set; } = 0; // 改装等级

        // 核心数值
        public double Speed { get; set; } = 20;      // 最高时速
        public double Handling { get; set; } = 10;   // 操控性（影响事件成功率）
        public double Tech { get; set; } = 5;        // 科技感（影响特殊奖励）
        public double Fuel { get; set; } = 100;      // 燃料/能量（消耗品）

        public DateTime LastActionTime { get; set; } = DateTime.Now;
        public DateTime CreateTime { get; set; } = DateTime.Now;

        [DbIgnore] public double ExpToNextLevel => 100 * Math.Pow(Level, 1.6) * ((int)Rarity + 1);
        
        [DbIgnore] public string RarityName => Rarity switch
        {
            VehicleRarity.Common => "⚪ 经济型",
            VehicleRarity.Rare => "🔵 舒适型",
            VehicleRarity.Epic => "🟣 豪华型",
            VehicleRarity.Legendary => "🟠 超级跑车",
            VehicleRarity.Mythic => "🔴 未来概念",
            _ => "未知"
        };

        public override string TableName => "UserVehicles";
        public override string KeyField => "Id";

        public static async Task<Vehicle?> GetActiveVehicleAsync(string userId)
        {
            return (await QueryWhere("UserId = @p1 AND Status = @p2", SqlParams(("@p1", userId), ("@p2", (int)VehicleStatus.Driving)))).FirstOrDefault();
        }

        public static async Task<List<Vehicle>> GetUserVehiclesAsync(string userId)
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
                Speed += 2 + (int)Rarity * 1.0;
                Handling += 1 + (int)Rarity * 0.5;
                Tech += 0.5 + (int)Rarity * 0.3;
            }
        }
    }

    #endregion

    #region Templates

    public class VehicleTemplate
    {
        public string Id { get; set; } = string.Empty;
        public string Name { get; set; } = string.Empty;
        public VehicleRarity Rarity { get; set; }
        public string AsciiArt { get; set; } = string.Empty;
        public string Description { get; set; } = string.Empty;
        public double BaseSpeed { get; set; }
        public double BaseHandling { get; set; }
        public double BaseTech { get; set; }

        [DbIgnore] public string RarityName => Rarity switch
        {
            VehicleRarity.Common => "⚪ 经济型",
            VehicleRarity.Rare => "🔵 舒适型",
            VehicleRarity.Epic => "🟣 豪华型",
            VehicleRarity.Legendary => "🟠 超级跑车",
            VehicleRarity.Mythic => "🔴 未来概念",
            _ => "未知"
        };

        public static readonly Dictionary<string, VehicleTemplate> All = new()
        {
            ["v_scooter"] = new VehicleTemplate 
            { 
                Id = "v_scooter", Name = "小电驴", Rarity = VehicleRarity.Common, 
                BaseSpeed = 30, BaseHandling = 15, BaseTech = 2,
                AsciiArt = "  __o \n _`\\<, \n(*)/(*)",
                Description = "穿梭在城市小巷的最佳选择，经济实惠。" 
            },
            ["v_suv"] = new VehicleTemplate 
            { 
                Id = "v_suv", Name = "越野悍马", Rarity = VehicleRarity.Rare, 
                BaseSpeed = 80, BaseHandling = 40, BaseTech = 10,
                AsciiArt = "  _______ \n /|_||_\\`.__ \n(   _    _ _\\ \n=`-(_)--(_)-' ",
                Description = "强悍的越野性能，无视任何地形。" 
            },
            ["v_supercar"] = new VehicleTemplate 
            { 
                Id = "v_supercar", Name = "幽灵之子", Rarity = VehicleRarity.Legendary, 
                BaseSpeed = 350, BaseHandling = 95, BaseTech = 50,
                AsciiArt = "     _______ \n  _ /_||_||_\\ _ \n [____________] \n  (_)      (_)  ",
                Description = "速度的极致，地表的飞行器。" 
            }
        };
    }

    #endregion
}
