using BotWorker.Domain.Interfaces;
using BotWorker.Domain.Entities;
using BotWorker.Domain.Repositories;
using System.Text;

namespace BotWorker.Modules.Games
{
    [BotPlugin(
        Id = "game.cultivation",
        Name = "玄幻修炼系统",
        Version = "1.0.0",
        Author = "Matrix",
        Description = "吸纳天地灵气，突破生死玄关，终成无上仙道。",
        Category = "Games"
    )]
    public class CultivationService : IPlugin
    {
        private readonly ICultivationProfileRepository _profileRepo;
        private readonly ICultivationRecordRepository _recordRepo;
        private readonly IUserRepository _userRepo;
        private const int CULTIVATE_COOLDOWN_MINUTES = 10;
        private const double BASE_BREAKTHROUGH_RATE = 0.95;

        public CultivationService(
            ICultivationProfileRepository profileRepo, 
            ICultivationRecordRepository recordRepo,
            IUserRepository userRepo)
        {
            _profileRepo = profileRepo;
            _recordRepo = recordRepo;
            _userRepo = userRepo;
        }

        public List<Intent> Intents => [
            new() { Name = "修炼", Keywords = ["修炼", "修行", "cultivate"] },
            new() { Name = "突破", Keywords = ["突破", "晋升", "breakthrough"] },
            new() { Name = "境界", Keywords = ["境界", "修为", "我的修为", "status"] },
            new() { Name = "修为榜", Keywords = ["修为榜", "强者榜", "rank"] }
        ];

        public async Task InitAsync(IRobot robot)
        {
            await _profileRepo.EnsureTableCreatedAsync();
            await _recordRepo.EnsureTableCreatedAsync();
            await robot.RegisterSkillAsync(new SkillCapability
            {
                Name = "玄幻修炼",
                Commands = ["修炼", "突破", "境界", "修为榜"],
                Description = "【修炼】吸收灵气增加修为；【突破】当修为圆满时冲击更高境界；【境界】查看个人修仙进度。"
            }, HandleCommandAsync);
        }

        public async Task StopAsync() => await Task.CompletedTask;

        private async Task<string> HandleCommandAsync(IPluginContext ctx, string[] args)
        {
            var cmd = ctx.RawMessage.Trim();
            if (cmd.StartsWith("修炼") || cmd.ToLower().StartsWith("cultivate")) return await CultivateAsync(ctx);
            if (cmd.StartsWith("突破") || cmd.ToLower().StartsWith("breakthrough")) return await BreakthroughAsync(ctx);
            if (cmd.StartsWith("境界") || cmd.Contains("修为") || cmd.ToLower().StartsWith("status")) return await GetStatusAsync(ctx);
            if (cmd.StartsWith("修为榜") || cmd.ToLower().StartsWith("rank")) return await GetRankAsync(ctx);

            return "未知指令。可用：修炼、突破、境界、修为榜。";
        }

        private async Task<string> CultivateAsync(IPluginContext ctx)
        {
            var profile = await GetOrCreateProfileAsync(ctx.UserId);

            // 检查冷却
            var nextTime = profile.LastCultivateTime.AddMinutes(CULTIVATE_COOLDOWN_MINUTES);
            if (DateTime.Now < nextTime)
            {
                var remain = (nextTime - DateTime.Now);
                return $"🧘 灵气尚未平复，请等待 {remain.Minutes} 分 {remain.Seconds} 秒后再试。";
            }

            // 计算收益
            int gain = Random.Shared.Next(profile.CultivationSpeed, profile.CultivationSpeed * 2);
            profile.Exp += gain;
            profile.LastCultivateTime = DateTime.Now;
            await _profileRepo.UpdateEntityAsync(profile);

            await _recordRepo.InsertAsync(new CultivationRecord
            {
                UserId = ctx.UserId,
                ActionType = "修炼",
                Detail = $"获得灵气 {gain}"
            });

            var sb = new StringBuilder();
            sb.AppendLine($"✨ 你盘膝而坐，运转功法，引天地灵气入体。");
            sb.AppendLine($"📈 修为提升了 {gain} 点！");
            sb.Append($"📊 当前进度：{profile.Exp}/{profile.MaxExp}");
            if (profile.Exp >= profile.MaxExp)
            {
                sb.Append("\n🌟 修为已达瓶颈，速速【突破】！");
            }

            return sb.ToString();
        }

        private async Task<string> BreakthroughAsync(IPluginContext ctx)
        {
            var profile = await GetOrCreateProfileAsync(ctx.UserId);

            if (profile.Exp < profile.MaxExp)
            {
                return $"❌ 修为不足，尚不足以冲击瓶颈！(当前: {profile.Exp}/{profile.MaxExp})";
            }

            // 计算成功率：随等级提升而降低，最低 30%
            double rate = Math.Max(0.3, BASE_BREAKTHROUGH_RATE - (profile.Level / 100.0) * 0.5);
            bool success = Random.Shared.NextDouble() < rate;

            if (success)
            {
                profile.Level++;
                profile.Exp = 0;
                profile.MaxExp = CalculateMaxExp(profile.Level);
                profile.CultivationSpeed = 10 + (profile.Level / 5) * 5; // 每5级提升基础速度
                await _profileRepo.UpdateEntityAsync(profile);

                await _recordRepo.InsertAsync(new CultivationRecord
                {
                    UserId = ctx.UserId,
                    ActionType = "突破",
                    Detail = $"成功突破至 {profile.GetRankDescription()}"
                });

                return $"🎉 恭喜！你成功冲破玄关，晋升至 【{profile.GetRankDescription()}】！灵觉大增，修炼速度提升。";
            }
            else
            {
                // 失败扣除一部分修为
                long loss = (long)(profile.MaxExp * 0.2);
                profile.Exp = Math.Max(0, profile.Exp - loss);
                await _profileRepo.UpdateEntityAsync(profile);

                await _recordRepo.InsertAsync(new CultivationRecord
                {
                    UserId = ctx.UserId,
                    ActionType = "走火入魔",
                    Detail = $"突破失败，损失修为 {loss}"
                });

                return $"💥 哎呀！突破时气息不稳导致走火入魔，损失了 {loss} 点修为. 莫要灰心，再接再厉！";
            }
        }

        private async Task<string> GetStatusAsync(IPluginContext ctx)
        {
            var profile = await GetOrCreateProfileAsync(ctx.UserId);
            var sb = new StringBuilder();
            sb.AppendLine($"👤 【{ctx.UserName}】的修仙进度");
            sb.AppendLine($"🌌 境界：{profile.GetRankDescription()} (Lv.{profile.Level})");
            sb.AppendLine($"🔮 修为：{profile.Exp} / {profile.MaxExp}");
            sb.AppendLine($"⚡ 修炼速度：{profile.CultivationSpeed} ~ {profile.CultivationSpeed * 2}");
            
            var nextTime = profile.LastCultivateTime.AddMinutes(CULTIVATE_COOLDOWN_MINUTES);
            if (DateTime.Now < nextTime)
            {
                var remain = nextTime - DateTime.Now;
                sb.AppendLine($"⏱️ 修炼冷却：还需 {remain.Minutes}分{remain.Seconds}秒");
            }
            else
            {
                sb.AppendLine("✅ 状态：灵气充沛，可随时修炼。");
            }

            return sb.ToString();
        }

        private async Task<string> GetRankAsync(IPluginContext ctx)
        {
            var top = await _profileRepo.GetTopCultivatorsAsync(10);
            if (top.Count == 0) return "暂时还没有修仙者出世。";

            var sb = new StringBuilder();
            sb.AppendLine("🏆 【修为强者榜】");
            for (int i = 0; i < top.Count; i++)
            {
                var p = top[i];
                string name = "神秘修仙者";
                if (long.TryParse(p.UserId, out long uid))
                {
                    var user = await _userRepo.GetByIdAsync(uid);
                    if (user != null) name = user.Name;
                }
                sb.AppendLine($"{i + 1}. {name} - {p.GetRankDescription()} (Lv.{p.Level})");
            }
            return sb.ToString();
        }

        private async Task<CultivationProfile> GetOrCreateProfileAsync(string userId)
        {
            var profile = await _profileRepo.GetByUserIdAsync(userId);
            if (profile == null)
            {
                profile = new CultivationProfile { UserId = userId };
                await _profileRepo.InsertAsync(profile);
            }
            return profile;
        }

        private long CalculateMaxExp(int level)
        {
            // 指数级增长或平滑增长
            return (long)(level * 100 * Math.Pow(1.1, level / 5.0));
        }
    }
}
