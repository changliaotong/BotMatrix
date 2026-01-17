using System;
using System.Data;
using System.Linq;
using System.Threading.Tasks;
using BotWorker.Domain.Interfaces;
using BotWorker.Domain.Models.BotMessages;
using BotWorker.Domain.Repositories;
using Dapper.Contrib.Extensions;
using Microsoft.Extensions.DependencyInjection;

namespace BotWorker.Modules.Games
{
    [BotPlugin(
        Id = "game.secretlove",
        Name = "暗恋系统",
        Version = "1.0.0",
        Author = "Matrix",
        Description = "登记暗恋对象，如果对方也暗恋你，则会触发匹配通知",
        Category = "Games"
    )]
    public class SecretLovePlugin : IPlugin
    {
        public async Task InitAsync(IRobot robot)
        {
            await robot.RegisterSkillAsync(new SkillCapability
            {
                Name = "暗恋系统",
                Commands = ["暗恋", "我的暗恋", "谁暗恋我"],
                Description = "登记：暗恋 @某人；查询：我的暗恋 /谁暗恋我"
            }, HandleLoveAsync);
        }

        public Task StopAsync() => Task.CompletedTask;

        private async Task<string> HandleLoveAsync(IPluginContext ctx, string[] args)
        {
            var userId = long.Parse(ctx.UserId);
            var groupId = long.Parse(ctx.GroupId ?? "0");
            var botId = long.Parse(ctx.BotId);
            var cmd = ctx.RawMessage.Trim().Split(' ')[0];

            if (cmd == "暗恋")
            {
                // 简单的从参数或提到的人中获取 ID
                if (args.Length == 0) return "请指定暗恋对象，例如：暗恋 @某人";
                
                // 假设 args[0] 是 QQ 号或者包含 QQ 号的字符串
                if (!long.TryParse(args[0].Replace("@", ""), out long loveId))
                    return "暗恋对象 ID 格式错误";

                if (loveId == userId) return "不能暗恋自己哦";

                var love = new SecretLove
                {
                    UserId = userId,
                    LoveId = loveId,
                    GroupId = groupId,
                    BotUin = botId
                };
                await love.InsertAsync();
                
                if (await SecretLove.IsLoveEachotherAsync(userId, loveId))
                {
                    return $"💖 恭喜！你和 @{loveId} 互相暗恋，匹配成功！";
                }
                
                return "✅ 已悄悄登记，如果对方也登记了你，系统会通知你们。";
            }
            else if (cmd == "我的暗恋")
            {
                var count = await SecretLove.GetCountLoveAsync(userId);
                return $"你一共登记了 {count} 个暗恋对象。";
            }
            else if (cmd == "谁暗恋我")
            {
                var count = await SecretLove.GetCountLoveMeAsync(userId);
                return $"共有 {count} 个人正在悄悄暗恋你。";
            }

            return await SecretLove.GetLoveStatusAsync();
        }
    }

    [Table("love")]
    public class SecretLove
    {
        [ExplicitKey]
        public long UserId { get; set; }
        [ExplicitKey]
        public long LoveId { get; set; }
        public long GroupId { get; set; }
        public long BotUin { get; set; }
    }
}
