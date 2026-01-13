using BotWorker.Domain.Interfaces;
using Microsoft.Extensions.Logging;
using System.Text;

namespace BotWorker.Modules.Games
{
    [BotPlugin(
        Id = "game.vehicle",
        Name = "座驾系统",
        Version = "1.0.0",
        Author = "Matrix",
        Description = "现代座驾系统：购买、改装、竞速、驾驶巡逻",
        Category = "Games"
    )]
    public class VehicleService : IPlugin
    {
        private IRobot? _robot;
        private ILogger? _logger;
        private readonly VehicleConfig _config;

        public VehicleService()
        {
            _config = new VehicleConfig();
        }

        public VehicleService(IRobot robot, ILogger logger, VehicleConfig config)
        {
            _robot = robot;
            _logger = logger;
            _config = config;
        }

        public List<Intent> Intents => [
            new() { Name = "我的座驾", Keywords = ["我的座驾", "vehicles", "cars"] },
            new() { Name = "驾驶座驾", Keywords = ["驾驶座驾", "drive"] },
            new() { Name = "购买座驾", Keywords = ["购买座驾", "buy_vehicle"] },
            new() { Name = "维修座驾", Keywords = ["维修座驾", "repair"] },
            new() { Name = "改装座驾", Keywords = ["改装座驾", "tune"] }
        ];

        public async Task InitAsync(IRobot robot)
        {
            _robot = robot;
            await EnsureTablesCreatedAsync();
            await robot.RegisterSkillAsync(new SkillCapability
            {
                Name = "座驾系统",
                Commands = ["我的座驾", "驾驶座驾", "座驾状态", "购买座驾", "维修座驾", "改装座驾"],
                Description = "【购买座驾】获取新车；【我的座驾】查看列表；【驾驶座驾】开启巡逻模式"
            }, HandleVehicleCommandAsync);
        }

        public async Task StopAsync() => await Task.CompletedTask;

        private async Task EnsureTablesCreatedAsync()
        {
            await Vehicle.EnsureTableCreatedAsync();
        }

        private async Task<string> HandleVehicleCommandAsync(IPluginContext ctx, string[] args)
        {
            var cmd = ctx.RawMessage.Trim().Split(' ')[0];
            return cmd switch
            {
                "我的座驾" or "vehicles" or "座驾状态" => await GetMyVehiclesAsync(ctx),
                "驾驶座驾" or "drive" => await DriveVehicleAsync(ctx, args),
                "购买座驾" or "buy_vehicle" => await BuyVehicleAsync(ctx, args),
                "改装座驾" or "tune" => await TuneVehicleAsync(ctx, args),
                "维修座驾" or "repair" => await RepairVehicleAsync(ctx),
                _ => "🏎️ 现代座驾系统：使用【购买座驾】来挑选你的第一辆车吧！"
            };
        }

        private async Task<string> GetMyVehiclesAsync(IPluginContext ctx)
        {
            var vehicles = await Vehicle.GetUserVehiclesAsync(ctx.UserId);
            if (vehicles.Count == 0) return "你名下还没有任何座驾。输入【购买座驾】去车展看看吧！";

            var sb = new StringBuilder();
            sb.AppendLine($"📂 【{ctx.UserName} 的车库】");
            sb.AppendLine("--------------------");
            foreach (var v in vehicles)
            {
                var statusStr = v.Status switch
                {
                    VehicleStatus.Driving => "🟢 驾驶中",
                    VehicleStatus.Repairing => "🔧 维修中",
                    VehicleStatus.Tuning => "🛠️ 改装中",
                    _ => "⚪ 停车中"
                };
                sb.AppendLine($"{v.RarityName} {v.Name} (Lv.{v.Level}) [{statusStr}]");
                sb.AppendLine($"⛽ 燃料: {v.Fuel:F0}/100 | 🏎️ 时速: {v.Speed:F1}");
            }
            return sb.ToString();
        }

        private async Task<string> DriveVehicleAsync(IPluginContext ctx, string[] args)
        {
            var vehicles = await Vehicle.GetUserVehiclesAsync(ctx.UserId);
            if (vehicles.Count == 0) return "你还没有座驾，请先【购买座驾】！";

            // 逻辑简化：如果有驾驶中的，先停掉
            var active = vehicles.FirstOrDefault(v => v.Status == VehicleStatus.Driving);
            if (active != null && (args.Length == 0 || active.Name != args[0]))
            {
                active.Status = VehicleStatus.Idle;
                await active.UpdateAsync();
            }

            var target = args.Length > 0 
                ? vehicles.FirstOrDefault(v => v.Name == args[0]) 
                : vehicles.FirstOrDefault();

            if (target == null) return "未找到指定的座驾。";
            if (target.Fuel < 10) return "燃料不足，请先【维修座驾】（加油）！";

            target.Status = VehicleStatus.Driving;
            target.LastActionTime = DateTime.Now;
            await target.UpdateAsync();

            return $"🏎️ 引擎轰鸣！你发动了 {target.Name}，开始在城市中巡逻！\n{VehicleTemplate.All.GetValueOrDefault(target.TemplateId)?.AsciiArt}";
        }

        private async Task<string> BuyVehicleAsync(IPluginContext ctx, string[] args)
        {
            // 这里应该对接积分系统，目前简化为直接获取
            if (args.Length == 0)
            {
                var sb = new StringBuilder();
                sb.AppendLine("🏪 【Matrix 车展中心】");
                foreach (var t in VehicleTemplate.All.Values)
                {
                    sb.AppendLine($"{t.RarityName} {t.Name} - {t.Description}");
                }
                sb.AppendLine("\n用法：购买座驾 [名称]");
                return sb.ToString();
            }

            var template = VehicleTemplate.All.Values.FirstOrDefault(t => t.Name == args[0]);
            if (template == null) return "展厅里没有这辆车。";

            var myVehicles = await Vehicle.GetUserVehiclesAsync(ctx.UserId);
            if (myVehicles.Count >= _config.MaxVehicleCount) return $"你的车库已满（上限 {_config.MaxVehicleCount} 辆）！";

            var vehicle = new Vehicle
            {
                UserId = ctx.UserId,
                Name = template.Name,
                TemplateId = template.Id,
                Rarity = template.Rarity,
                Speed = template.BaseSpeed,
                Handling = template.BaseHandling,
                Tech = template.BaseTech,
                Status = VehicleStatus.Idle
            };
            await vehicle.InsertAsync();

            return $"🎊 恭喜！你成功购买了 {template.Name}，已送往你的车库！";
        }

        private async Task<string> TuneVehicleAsync(IPluginContext ctx, string[] args)
        {
            var active = await Vehicle.GetActiveVehicleAsync(ctx.UserId);
            if (active == null) return "你必须先【驾驶座驾】才能进行改装！";

            if (DateTime.Now - active.LastActionTime < TimeSpan.FromMinutes(5))
                return "零件还在冷却中，请稍后再试（改装冷却：5分钟）。";

            var success = Random.Shared.NextDouble() < _config.TuningSuccessRate;
            active.LastActionTime = DateTime.Now;
            
            if (success)
            {
                var expGain = 50 * (1 + (int)active.Rarity * 0.5);
                var oldLevel = active.Level;
                active.GainExp(expGain);
                active.ModificationLevel++;
                await active.UpdateAsync();

                var sb = new StringBuilder();
                sb.AppendLine($"🛠️ 改装成功！{active.Name} 的性能得到了提升！");
                if (active.Level > oldLevel)
                    sb.AppendLine($"🎊 等级提升至 Lv.{active.Level}！");
                return sb.ToString();
            }
            else
            {
                active.Fuel -= 10;
                await active.UpdateAsync();
                return $"💥 改装失败！虽然浪费了一些燃料，但你积累了宝贵的失败经验。";
            }
        }

        private async Task<string> RepairVehicleAsync(IPluginContext ctx)
        {
            var vehicles = await Vehicle.GetUserVehiclesAsync(ctx.UserId);
            var toRepair = vehicles.FirstOrDefault(v => v.Fuel < 100);
            if (toRepair == null) return "你的所有座驾都状态良好，无需维修或加油。";

            toRepair.Fuel = 100;
            toRepair.Status = VehicleStatus.Idle;
            await toRepair.UpdateAsync();

            return $"🔧 经过一番整备，{toRepair.Name} 已恢复至最佳状态！燃料已加满。";
        }
    }
}
