using BotWorker.Domain.Interfaces;
using BotWorker.Domain.Entities;
using System.Text;

namespace BotWorker.Modules.Games
{
    [BotPlugin(
        Id = "game.brick",
        Name = "江湖闷砖系统",
        Version = "1.0.0",
        Author = "Matrix",
        Description = "武功再高，也怕砖头：掏出板砖，闷声发大财！",
        Category = "Games"
    )]
    public class BrickService : IPlugin
    {
        private IRobot? _robot;
        private const int BRICK_COST = 50;           // 拍一次砖消耗50积分
        private const int ACTION_COOLDOWN_SEC = 300; // 冷却5分钟
        private const double SUCCESS_RATE = 0.65;    // 基础成功率

        public List<Intent> Intents => [
            new() { Name = "闷砖", Keywords = ["闷砖", "拍砖", "brick"] },
            new() { Name = "砖头榜", Keywords = ["砖头榜", "板砖榜"] }
        ];

        public async Task InitAsync(IRobot robot)
        {
            _robot = robot;
            await BrickRecord.EnsureTableCreatedAsync();
            await robot.RegisterSkillAsync(new SkillCapability
            {
                Name = "江湖闷砖",
                Commands = ["闷砖 @某人", "砖头榜"],
                Description = "【闷砖 @某人】消耗50积分尝试偷袭对方。成功可使其禁言并抢夺少量积分，失败则自食其果！"
            }, HandleCommandAsync);
        }

        public async Task StopAsync() => await Task.CompletedTask;

        private async Task<string> HandleCommandAsync(IPluginContext ctx, string[] args)
        {
            var cmd = ctx.RawMessage.Trim().Split(' ')[0];
            if (cmd == "砖头榜") return await GetRankAsync();

            if (ctx.GroupId == null) return "❌ 拍砖只能在群聊中进行，私聊拍空气吗？";

            // 获取目标用户
            var target = ctx.MentionedUsers.FirstOrDefault();
            if (target == null) return "❌ 请艾特你要闷砖的对象！例如：闷砖 @小明";

            if (target.UserId == ctx.UserId) return "🤕 你举起砖头看了看，最后决定拍在自己脑门上。好疼！";

            long attackerId = long.Parse(ctx.UserId);
            long victimId = long.Parse(target.UserId);
            long botId = long.Parse(ctx.BotId);
            long groupId = long.Parse(ctx.GroupId);

            // 1. 检查冷却
            var lastTime = await BrickRecord.GetLastActionTimeAsync(ctx.UserId);
            if (DateTime.Now < lastTime.AddSeconds(ACTION_COOLDOWN_SEC))
            {
                var remain = (int)(lastTime.AddSeconds(ACTION_COOLDOWN_SEC) - DateTime.Now).TotalSeconds;
                return $"⏱️ 你的板砖还没擦干净，请等待 {remain} 秒再行动。";
            }

            // 2. 检查积分是否足够
            long myCredit = await UserInfo.GetCreditAsync(groupId, attackerId);
            if (myCredit < BRICK_COST) return $"❌ 拍砖需要消耗 {BRICK_COST} 积分，你太穷了，连搬砖的力气都没有。";

            // 3. 执行扣分 (买砖头)
            await UserInfo.AddCreditAsync(botId, groupId, ctx.GroupName ?? "江湖", attackerId, ctx.UserName, -BRICK_COST, "购买板砖");

            // 4. 判定结果
            bool isSuccess = Random.Shared.NextDouble() < SUCCESS_RATE;
            int muteSec = Random.Shared.Next(60, 301); // 1-5分钟
            long stolenCredit = Random.Shared.Next(20, 101); // 抢20-100积分

            var record = new BrickRecord
            {
                AttackerId = ctx.UserId,
                TargetId = target.UserId,
                GroupId = ctx.GroupId,
                IsSuccess = isSuccess,
                ActionTime = DateTime.Now
            };

            var sb = new StringBuilder();
            if (isSuccess)
            {
                // 成功：抢分 + 禁言
                var transRes = await UserInfo.TransferCreditAsync(
                    botId, groupId, ctx.GroupName ?? "江湖",
                    victimId, target.Name,
                    attackerId, ctx.UserName,
                    stolenCredit, stolenCredit, "被闷砖抢夺");

                sb.AppendLine($"🧱 【{ctx.UserName}】掏出一块被报纸包着的板砖，趁【{target.Name}】不备猛地拍了下去！");
                
                if (transRes.Result == 0)
                {
                    sb.AppendLine($"💰 趁对方眼冒金星，你顺手摸走了 {stolenCredit} 积分。");
                }

                // 尝试禁言 (如果机器人有权限)
                try
                {
                    // 这里我们通过 Skill 调用禁言
                    await ctx.ReplyAsync(sb.ToString()); // 先回复文字
                    await Task.Delay(500);
                    
                    if (_robot != null)
                    {
                        await _robot.CallSkillAsync("MuteMember", ctx, ["Mute", victimId.ToString(), muteSec.ToString()]);
                    }
                    return $"🤫 【{target.Name}】被拍晕了，进入了 {muteSec / 60} 分钟的贤者模式。";
                }
                catch
                {
                    return sb.ToString() + "\n(官府禁言失败，看来对方后台很硬！)";
                }
            }
            else
            {
                // 失败：自食其果
                bool backfire = Random.Shared.NextDouble() < 0.4; // 40%概率反噬
                if (backfire)
                {
                    sb.AppendLine($"🙈 【{ctx.UserName}】试图偷袭 【{target.Name}】，结果脚下一滑，砖头脱手飞出砸到了自己！");
                    await UserInfo.AddCreditAsync(botId, groupId, ctx.GroupName ?? "江湖", attackerId, ctx.UserName, -stolenCredit, "拍砖反噬罚款");
                    sb.AppendLine($"💸 你不仅没拍到人，还因为医药费损失了 {stolenCredit} 积分。");
                    
                    if (_robot != null)
                    {
                        await _robot.CallSkillAsync("MuteMember", ctx, ["Mute", attackerId.ToString(), "60"]);
                        sb.AppendLine("🤐 你把自己拍晕了 1 分钟。");
                    }
                }
                else
                {
                    sb.AppendLine($"🛡️ 【{target.Name}】背后长了眼睛，一个闪身躲过了 【{ctx.UserName}】 的板砖。砖头碎了一地！");
                }
            }

            record.IsSuccess = isSuccess;
            await record.InsertAsync();

            return sb.ToString();
        }

        private async Task<string> GetRankAsync()
        {
            var tops = await BrickRecord.GetTopAttackersAsync();
            if (tops.Count == 0) return "🏮 江湖一片祥和，还没有人开始拍砖。";

            var sb = new StringBuilder();
            sb.AppendLine("🏆 【江湖板砖英雄榜】");
            sb.AppendLine("━━━━━━━━━━━━━━");
            int rank = 1;
            foreach (var t in tops)
            {
                sb.AppendLine($"{rank++}. 用户({t.UserId}) - 成功拍砖 {t.Count} 次");
            }
            sb.AppendLine("━━━━━━━━━━━━━━");
            sb.Append("💡 提示：多行不义必自毙，拍砖请谨慎！");
            return sb.ToString();
        }
    }
}
