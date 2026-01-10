using BotWorker.Domain.Interfaces;
using BotWorker.Domain.Entities;
using System.Text;

namespace BotWorker.Modules.Games
{
    [BotPlugin(
        Id = "game.robbery",
        Name = "江湖打劫系统",
        Version = "1.0.0",
        Author = "Matrix",
        Description = "富贵险中求：打劫他人积分，但要小心反杀！",
        Category = "Games"
    )]
    public class RobberyService : IPlugin
    {
        private const int ROB_COOLDOWN_MINUTES = 20; // 打劫冷却时间
        private const int PROTECTION_MINUTES = 30;   // 被打劫保护时间
        private const double BASE_SUCCESS_RATE = 0.4; // 基础成功率

        public List<Intent> Intents => [
            new() { Name = "打劫", Keywords = ["打劫", "rob"] }
        ];

        public async Task InitAsync(IRobot robot)
        {
            await RobberyRecord.EnsureTableCreatedAsync();
            await robot.RegisterSkillAsync(new SkillCapability
            {
                Name = "江湖打劫",
                Commands = ["打劫 @某人"],
                Description = "【打劫 @某人】尝试抢夺对方积分。成功则获利，失败可能反被罚！"
            }, HandleCommandAsync);
        }

        public async Task StopAsync() => await Task.CompletedTask;

        private async Task<string> HandleCommandAsync(IPluginContext ctx, string[] args)
        {
            if (ctx.GroupId == null) return "❌ 打劫只能在群聊中进行。";

            // 获取目标用户
            var target = ctx.MentionedUsers.FirstOrDefault();
            if (target == null) return "❌ 请艾特你要打劫的对象！例如：打劫 @小明";

            if (target.UserId == ctx.UserId) return "🤣 你想打劫你自己？这操作我看不懂。";

            long robberId = long.Parse(ctx.UserId);
            long victimId = long.Parse(target.UserId);
            long botId = long.Parse(ctx.BotId);
            long groupId = long.Parse(ctx.GroupId);

            // 1. 检查打劫者 CD
            var lastRobTime = await RobberyRecord.GetLastRobTimeAsync(ctx.UserId);
            var nextRobTime = lastRobTime.AddMinutes(ROB_COOLDOWN_MINUTES);
            if (DateTime.Now < nextRobTime)
            {
                var waitMin = (int)(nextRobTime - DateTime.Now).TotalMinutes;
                return $"⏱️ 你的体力还没恢复，请休息 {waitMin} 分钟后再行凶。";
            }

            // 2. 检查被劫者保护期
            var protectionEnd = await RobberyRecord.GetProtectionEndTimeAsync(target.UserId);
            if (DateTime.Now < protectionEnd)
            {
                var protectMin = (int)(protectionEnd - DateTime.Now).TotalMinutes;
                return $"🛡️ 【{target.Name}】正处于官府保护期，还剩 {protectMin} 分钟，现在下手太危险了！";
            }

            // 3. 获取双方积分
            long victimCredit = await UserInfo.GetCreditAsync(groupId, victimId);
            if (victimCredit < 100) return $"❌ 【{target.Name}】太穷了（积分不足100），连土匪都看不上他。";

            long robberCredit = await UserInfo.GetCreditAsync(groupId, robberId);

            // 4. 计算打劫金额 (抢夺 5% - 15%)
            double percent = Random.Shared.Next(5, 16) / 100.0;
            long amount = (long)(victimCredit * percent);
            if (amount < 10) amount = 10;

            // 5. 判定结果
            bool isSuccess = Random.Shared.NextDouble() < BASE_SUCCESS_RATE;
            
            var record = new RobberyRecord
            {
                RobberId = ctx.UserId,
                VictimId = target.UserId,
                GroupId = ctx.GroupId,
                Amount = amount,
                IsSuccess = isSuccess,
                RobTime = DateTime.Now
            };

            var sb = new StringBuilder();
            if (isSuccess)
            {
                // 打劫成功：积分转移
                var transRes = await UserInfo.TransferCreditAsync(
                    botId, groupId, ctx.GroupName ?? "江湖",
                    victimId, target.Name,
                    robberId, ctx.UserName,
                    amount, amount, "江湖打劫");

                if (transRes.Result == 0)
                {
                    sb.AppendLine($"⚔️ 【{ctx.UserName}】蒙面潜入 【{target.Name}】 的住所...");
                    sb.AppendLine($"💰 成功得手！抢走了对方 {amount} 积分！");
                    sb.AppendLine($"📈 你的当前积分：{transRes.ReceiverCredit}");
                }
                else
                {
                    return "❌ 打劫过程中官府干预，交易失败（系统错误）。";
                }
            }
            else
            {
                // 打劫失败：反被罚款 (扣除打劫者尝试金额的 50% 补偿给对方，或直接没收)
                long penalty = amount / 2;
                if (robberCredit < penalty) penalty = robberCredit;

                if (penalty > 0)
                {
                    var transRes = await UserInfo.TransferCreditAsync(
                        botId, groupId, ctx.GroupName ?? "江湖",
                        robberId, ctx.UserName,
                        victimId, target.Name,
                        penalty, penalty, "打劫失败赔偿");
                    
                    sb.AppendLine($"💀 【{ctx.UserName}】试图打劫 【{target.Name}】，结果被对方一顿反杀！");
                    sb.AppendLine($"💸 逃跑时不小心掉落了 {penalty} 积分，便宜了对方。");
                }
                else
                {
                    sb.AppendLine($"🚶 【{ctx.UserName}】试图打劫 【{target.Name}】，结果对方早有防备，打劫失败！");
                }
            }

            record.ResultMessage = sb.ToString();
            await record.InsertAsync();

            return sb.ToString();
        }
    }
}
