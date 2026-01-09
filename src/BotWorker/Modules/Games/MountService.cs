using BotWorker.Domain.Interfaces;
using Microsoft.Extensions.Logging;
using System.Text;

namespace BotWorker.Modules.Games
{
    [BotPlugin(
        Id = "game.mount",
        Name = "超级坐骑系统",
        Version = "1.0.0",
        Author = "Matrix",
        Description = "超级牛逼的坐骑系统：捕捉、培养、进化、骑乘战斗",
        Category = "Games"
    )]
    public class MountService : IPlugin
    {
        private IRobot? _robot;
        private ILogger? _logger;
        private readonly MountConfig _config;

        public MountService()
        {
            _config = new MountConfig();
        }

        public MountService(IRobot robot, ILogger logger, MountConfig config)
        {
            _robot = robot;
            _logger = logger;
            _config = config;
        }

        public List<Intent> Intents => [
            new() { Name = "我的坐骑", Keywords = ["我的坐骑", "mounts"] },
            new() { Name = "召唤坐骑", Keywords = ["召唤坐骑", "ride"] },
            new() { Name = "寻找坐骑", Keywords = ["寻找坐骑", "capture"] }
        ];

        public async Task InitAsync(IRobot robot)
        {
            _robot = robot;
            await EnsureTablesCreatedAsync();
            await robot.RegisterSkillAsync(new SkillCapability
            {
                Name = "坐骑系统",
                Commands = ["我的坐骑", "召唤坐骑", "坐骑状态", "寻找坐骑", "坐骑休息", "坐骑训练", "坐骑改名"],
                Description = "【寻找坐骑】开启冒险；【我的坐骑】查看列表；【召唤坐骑】获得神力加成"
            }, HandleMountCommandAsync);
        }

        public async Task StopAsync() => await Task.CompletedTask;

        private async Task EnsureTablesCreatedAsync()
        {
            await Mount.EnsureTableCreatedAsync();
        }

        private async Task<string> HandleMountCommandAsync(IPluginContext ctx, string[] args)
        {
            var cmd = ctx.RawMessage.Trim().Split(' ')[0];
            return cmd switch
            {
                "我的坐骑" or "mounts" or "坐骑状态" => await GetMyMountsAsync(ctx),
                "召唤坐骑" or "ride" => await RideMountAsync(ctx, args),
                "寻找坐骑" or "capture" => await CaptureMountAsync(ctx),
                "坐骑训练" or "train" => await TrainMountAsync(ctx, args),
                _ => "🔮 强大的坐骑系统：使用【寻找坐骑】来开始你的征程吧！"
            };
        }

        private async Task<string> GetMyMountsAsync(IPluginContext ctx)
        {
            var mounts = await Mount.GetUserMountsAsync(ctx.UserId);
            if (mounts.Count == 0) return "你名下暂无坐骑，快去【寻找坐骑】吧！";

            var sb = new StringBuilder();
            sb.AppendLine("🏇 【我的马厩】");
            sb.AppendLine("━━━━━━━━━━━━━━");
            foreach (var m in mounts)
            {
                var template = MountTemplate.All.GetValueOrDefault(m.TemplateId);
                var statusIcon = m.Status == MountStatus.Riding ? "✨ [骑乘中]" : "";
                sb.AppendLine($"{m.RarityName} {m.Name} (Lv.{m.Level}) {statusIcon}");
                sb.AppendLine($"  - 速度: {m.Speed:F1} | 力量: {m.Power:F1} | 幸运: {m.Luck:F1}");
            }
            sb.AppendLine("━━━━━━━━━━━━━━");
            sb.Append("💡 提示：使用【召唤坐骑 名字】来驾驭它们！");
            return sb.ToString();
        }

        private async Task<string> RideMountAsync(IPluginContext ctx, string[] args)
        {
            if (args.Length == 0) return "请输入你想召唤的坐骑名称！";
            var targetName = args[0];

            var mounts = await Mount.GetUserMountsAsync(ctx.UserId);
            var target = mounts.FirstOrDefault(m => m.Name == targetName);
            if (target == null) return $"你马厩里没有叫 {targetName} 的坐骑。";

            // 取消其他骑乘状态
            foreach (var m in mounts.Where(x => x.Status == MountStatus.Riding))
            {
                m.Status = MountStatus.Idle;
                await m.UpdateAsync();
            }

            target.Status = MountStatus.Riding;
            await target.UpdateAsync();

            var template = MountTemplate.All.GetValueOrDefault(target.TemplateId);
            var sb = new StringBuilder();
            sb.AppendLine(template?.AsciiArt ?? "");
            sb.AppendLine($"🌟 你成功召唤了 {target.Name}！");
            sb.AppendLine($"感受到一股强大的力量正在涌动，你的全属性得到了加成！");
            return sb.ToString();
        }

        private async Task<string> CaptureMountAsync(IPluginContext ctx)
        {
            var mounts = await Mount.GetUserMountsAsync(ctx.UserId);
            if (mounts.Count >= _config.MaxMountCount) return "你的马厩已经满了，无法容纳更多坐骑！";

            // 简单的随机逻辑
            var templates = MountTemplate.All.Values.ToList();
            var roll = Random.Shared.NextDouble();
            
            MountTemplate selected;
            if (roll < 0.05) selected = templates.First(t => t.Rarity == MountRarity.Legendary);
            else if (roll < 0.2) selected = templates.First(t => t.Rarity == MountRarity.Rare);
            else selected = templates.First(t => t.Rarity == MountRarity.Common);

            var newMount = new Mount
            {
                UserId = ctx.UserId,
                Name = selected.Name,
                TemplateId = selected.Id,
                Rarity = selected.Rarity,
                Speed = selected.BaseSpeed,
                Power = selected.BasePower,
                Luck = selected.BaseLuck,
                Status = MountStatus.Idle,
                CreateTime = DateTime.Now
            };

            await newMount.InsertAsync();

            var sb = new StringBuilder();
            sb.AppendLine("🌲 你在野外探险时...");
            sb.AppendLine(selected.AsciiArt);
            sb.AppendLine($"🎊 奇迹发生了！你成功驯服了 {selected.RarityName} 级别的坐骑：【{selected.Name}】！");
            return sb.ToString();
        }

        private async Task<string> TrainMountAsync(IPluginContext ctx, string[] args)
        {
            var active = await Mount.GetActiveMountAsync(ctx.UserId);
            if (active == null) return "你必须先【召唤坐骑】才能进行训练！";

            if (DateTime.Now - active.LastActionTime < TimeSpan.FromMinutes(10))
                return "坐骑太累了，先让它休息一会儿吧（训练冷却：10分钟）。";

            var expGain = 20 * (1 + (int)active.Rarity * 0.5);
            var oldLevel = active.Level;
            active.GainExp(expGain);
            active.LastActionTime = DateTime.Now;
            await active.UpdateAsync();

            var sb = new StringBuilder();
            sb.AppendLine($"💪 经过一番艰苦的训练，{active.Name} 获得了 {expGain:F0} 点经验！");
            if (active.Level > oldLevel)
            {
                sb.AppendLine($"🎊 突破！坐骑等级提升至 Lv.{active.Level}！");
                sb.AppendLine($"📈 属性得到了全面强化！");
            }
            return sb.ToString();
        }
    }
}
