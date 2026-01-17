using BotWorker.Domain.Interfaces;
using BotWorker.Domain.Entities.Zodiac;
using BotWorker.Domain.Repositories;
using BotWorker.Modules.Zodiac;
using System.Text;

namespace BotWorker.Modules.Games
{
    [BotPlugin(
        Id = "game.pairing",
        Name = "缘分配对系统",
        Version = "1.0.0",
        Author = "Matrix",
        Description = "寻找你的灵魂伴侣：资料注册、缘分匹配、配对广场",
        Category = "Games"
    )]
    public class PairingService : IPlugin
    {
        private readonly IUserPairingProfileRepository _profileRepo;
        private readonly IPairingRecordRepository _pairingRepo;

        public PairingService(IUserPairingProfileRepository profileRepo, IPairingRecordRepository pairingRepo)
        {
            _profileRepo = profileRepo;
            _pairingRepo = pairingRepo;
        }

        public List<Intent> Intents => [
            new() { Name = "注册配对", Keywords = ["注册配对", "设置资料"] },
            new() { Name = "我的资料", Keywords = ["我的资料", "配对资料"] },
            new() { Name = "寻找配对", Keywords = ["寻找配对", "随机匹配"] },
            new() { Name = "配对广场", Keywords = ["配对广场", "单身榜"] },
            new() { Name = "解除配对", Keywords = ["解除配对", "结束缘分"] }
        ];

        public async Task InitAsync(IRobot robot)
        {
            await EnsureTablesCreatedAsync();
            await robot.RegisterSkillAsync(new SkillCapability
            {
                Name = "缘分配对",
                Commands = ["注册配对", "我的资料", "寻找配对", "配对广场", "解除配对"],
                Description = "【注册配对 性别 星座 简介】开启缘分；【寻找配对】寻找另一半；【配对广场】查看所有单身汪"
            }, HandleCommandAsync);
        }

        public async Task StopAsync() => await Task.CompletedTask;

        private async Task EnsureTablesCreatedAsync()
        {
            await _profileRepo.EnsureTableCreatedAsync();
            await _pairingRepo.EnsureTableCreatedAsync();
        }

        private async Task<string> HandleCommandAsync(IPluginContext ctx, string[] args)
        {
            var cmd = ctx.RawMessage.Trim().Split(' ')[0];
            return cmd switch
            {
                "注册配对" or "设置资料" => await RegisterProfileAsync(ctx, args),
                "我的资料" or "配对资料" => await GetMyProfileAsync(ctx),
                "寻找配对" or "随机匹配" => await MatchAsync(ctx),
                "配对广场" or "单身榜" => await GetSquareAsync(ctx),
                "解除配对" or "结束缘分" => await BreakPairAsync(ctx),
                _ => "💘 欢迎来到缘分配对中心！输入【注册配对】开始寻找你的另一半吧！"
            };
        }

        private async Task<string> RegisterProfileAsync(IPluginContext ctx, string[] args)
        {
            if (args.Length < 2) return "请输入：注册配对 [性别] [星座] [简介(可选)]\n例如：注册配对 男 狮子座 喜欢猫的阳光男孩";

            var gender = args[0];
            var zodiac = args[1];
            var intro = args.Length > 2 ? string.Join(" ", args.Skip(2)) : "这个人很懒，什么都没留下。";

            if (!zodiac.EndsWith("座")) zodiac += "座";

            var profile = await _profileRepo.GetByUserIdAsync(ctx.UserId);
            bool isNew = false;
            if (profile == null)
            {
                isNew = true;
                profile = new UserPairingProfile
                {
                    UserId = ctx.UserId,
                    Nickname = ctx.UserName
                };
            }

            profile.Gender = gender;
            profile.Zodiac = zodiac;
            profile.Intro = intro;
            profile.LastActive = DateTime.Now;
            profile.IsLooking = true;

            if (isNew)
                await _profileRepo.InsertAsync(profile);
            else
                await _profileRepo.UpdateAsync(profile);

            return $"✅ 资料注册成功！你已加入配对广场。\n🎭 昵称：{profile.Nickname}\n🚻 性别：{profile.Gender}\n✨ 星座：{profile.Zodiac}\n📝 简介：{profile.Intro}";
        }

        private async Task<string> GetMyProfileAsync(IPluginContext ctx)
        {
            var profile = await _profileRepo.GetByUserIdAsync(ctx.UserId);
            if (profile == null) return "你还没有注册配对资料，请输入【注册配对】。";

            var pair = await _pairingRepo.GetCurrentPairAsync(ctx.UserId);
            var pairStatus = pair != null ? $"💞 已与 【{(pair.User1Id == ctx.UserId ? pair.User2Id : pair.User1Id)}】 配对" : "🍃 目前单身";

            var sb = new StringBuilder();
            sb.AppendLine($"👤 【{profile.Nickname}】的缘分资料");
            sb.AppendLine($"━━━━━━━━━━━━━━");
            sb.AppendLine($"🚻 性别：{profile.Gender}");
            sb.AppendLine($"✨ 星座：{profile.Zodiac}");
            sb.AppendLine($"📝 简介：{profile.Intro}");
            sb.AppendLine($"💓 状态：{pairStatus}");
            sb.AppendLine($"🕒 最后活跃：{profile.LastActive:yyyy-MM-dd HH:mm}");
            sb.AppendLine($"━━━━━━━━━━━━━━");

            return sb.ToString();
        }

        private async Task<string> MatchAsync(IPluginContext ctx)
        {
            var me = await _profileRepo.GetByUserIdAsync(ctx.UserId);
            if (me == null) return "请先【注册配对】后再寻找缘分！";

            var currentPair = await _pairingRepo.GetCurrentPairAsync(ctx.UserId);
            if (currentPair != null) return "你已经有配对对象了，请先【解除配对】再寻找新缘分。";

            // 寻找活跃的单身用户 (排除自己)
            var seekers = await _profileRepo.GetActiveSeekersAsync(50);
            var filteredSeekers = seekers.Where(s => s.UserId != ctx.UserId).ToList();

            if (filteredSeekers.Count == 0) return "哎呀，广场上暂时没有其他正在寻找配对的人，请稍后再试。";

            // 随机选一个
            var target = filteredSeekers[Random.Shared.Next(filteredSeekers.Count)];

            // 计算星座契合度
            var matchInfo = ZodiacMatcher.GetMatchInfo(me.Zodiac, target.Zodiac);

            // 建立配对记录
            var record = new PairingRecord
            {
                User1Id = ctx.UserId,
                User2Id = target.UserId,
                Status = "coupled",
                PairDate = DateTime.Now
            };
            await _pairingRepo.InsertAsync(record);

            // 更新双方状态
            me.IsLooking = false;
            await _profileRepo.UpdateAsync(me);
            target.IsLooking = false;
            await _profileRepo.UpdateAsync(target);

            var sb = new StringBuilder();
            sb.AppendLine("💘 【缘分降临】 💘");
            sb.AppendLine($"恭喜！你与 【{target.Nickname}】 成功配对！");
            sb.AppendLine($"━━━━━━━━━━━━━━");
            sb.AppendLine($"✨ 对方星座：{target.Zodiac}");
            sb.AppendLine($"🔮 星座契合：{matchInfo}");
            sb.AppendLine($"📝 对方简介：{target.Intro}");
            sb.AppendLine($"━━━━━━━━━━━━━━");
            sb.AppendLine("💡 提示：快去打个招呼吧！如果不合适，可以输入【解除配对】。");
            return sb.ToString();
        }

        private async Task<string> GetSquareAsync(IPluginContext ctx)
        {
            var seekers = await _profileRepo.GetActiveSeekersAsync(10);
            if (seekers.Count == 0) return "配对广场目前空空如也，快来【注册配对】成为第一个吧！";

            var sb = new StringBuilder();
            sb.AppendLine("🏮 【配对广场 - 缘分速递】");
            sb.AppendLine($"━━━━━━━━━━━━━━");
            foreach (var s in seekers)
            {
                sb.AppendLine($"• {s.Nickname} ({s.Gender} | {s.Zodiac})");
                sb.AppendLine($"  \"{s.Intro}\"");
            }
            sb.AppendLine($"━━━━━━━━━━━━━━");
            sb.Append("💬 输入【寻找配对】开始随机匹配缘分！");

            return sb.ToString();
        }

        private async Task<string> BreakPairAsync(IPluginContext ctx)
        {
            var pair = await _pairingRepo.GetCurrentPairAsync(ctx.UserId);
            if (pair == null) return "你目前没有配对对象。";

            pair.Status = "broken";
            await _pairingRepo.UpdateAsync(pair);

            // 恢复单身状态
            var me = await _profileRepo.GetByUserIdAsync(ctx.UserId);
            if (me != null) { me.IsLooking = true; await _profileRepo.UpdateAsync(me); }

            var otherId = pair.User1Id == ctx.UserId ? pair.User2Id : pair.User1Id;
            var other = await _profileRepo.GetByUserIdAsync(otherId);
            if (other != null) { other.IsLooking = true; await _profileRepo.UpdateAsync(other); }

            return "💔 缘尽于此。你已恢复单身状态，资料重新进入配对广场。";
        }
    }
}
