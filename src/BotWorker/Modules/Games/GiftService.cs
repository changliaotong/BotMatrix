using BotWorker.Domain.Interfaces;
using BotWorker.Domain.Entities;
using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;

using System.Runtime.CompilerServices;

[assembly: InternalsVisibleTo("BotWorker.Tests")]

namespace BotWorker.Modules.Games
{
    [BotPlugin(
        Id = "game.gift",
        Name = "礼物互动系统",
        Version = "1.0.0",
        Author = "Matrix",
        Description = "购买精美礼物，赠送给心仪的Ta，增进彼此情谊。",
        Category = "Games"
    )]
    public class GiftService : IPlugin
    {
        public List<Intent> Intents => [
            new() { Name = "礼物商店", Keywords = ["礼物商店", "礼物列表", "gift shop"] },
            new() { Name = "购买礼物", Keywords = ["购买礼物", "buy gift"] },
            new() { Name = "我的背包", Keywords = ["我的背包", "我的礼物", "backpack"] },
            new() { Name = "送礼物", Keywords = ["送礼物", "赠送礼物", "send gift"] },
            new() { Name = "礼物日志", Keywords = ["礼物日志", "礼物记录", "gift logs"] }
        ];

        public async Task InitAsync(IRobot robot)
        {
            await GiftStoreItem.EnsureTableCreatedAsync();
            await GiftBackpack.EnsureTableCreatedAsync();
            await GiftRecord.EnsureTableCreatedAsync();

            // 初始化默认礼物
                long count = await GiftStoreItem.CountAsync();
                Console.WriteLine($"[礼物系统] 当前礼物数量: {count}");
                if (count == 0)
                {
                    var defaults = new List<GiftStoreItem>
                    {
                        new() { GiftName = "鲜花", GiftCredit = 50, GiftType = 1, IsValid = true },
                        new() { GiftName = "巧克力", GiftCredit = 200, GiftType = 1, IsValid = true },
                        new() { GiftName = "蛋糕", GiftCredit = 500, GiftType = 1, IsValid = true },
                        new() { GiftName = "钻戒", GiftCredit = 2000, GiftType = 2, IsValid = true },
                        new() { GiftName = "跑车", GiftCredit = 10000, GiftType = 2, IsValid = true }
                    };
                    foreach (var item in defaults)
                    {
                        await item.InsertAsync();
                        Console.WriteLine($"[礼物系统] 插入默认礼物: {item.GiftName}");
                    }
                    Console.WriteLine($"[礼物系统] 已初始化 {defaults.Count} 个默认礼物。");
                }

            await robot.RegisterSkillAsync(new SkillCapability
            {
                Name = "礼物互动",
                Commands = ["礼物商店", "购买礼物", "我的背包", "送礼物", "礼物日志"],
                Description = "【礼物商店】查看可购买的礼物；【购买礼物】使用积分购买礼物到背包；【我的背包】查看拥有的礼物；【送礼物】将礼物送给他人；【礼物日志】查看往来记录。"
            }, HandleCommandAsync);
        }

        internal async Task<string> HandleCommandAsync(IPluginContext ctx, string[] args)
        {
            var cmd = ctx.RawMessage.Trim();
            if (cmd.StartsWith("礼物商店") || cmd.StartsWith("礼物列表")) return await GetShopListAsync();
            if (cmd.StartsWith("购买礼物")) return await BuyGiftAsync(ctx, args);
            if (cmd.StartsWith("我的背包") || cmd.StartsWith("我的礼物")) return await GetBackpackAsync(ctx);
            if (cmd.StartsWith("送礼物") || cmd.StartsWith("赠送礼物")) return await SendGiftAsync(ctx, args);
            if (cmd.StartsWith("礼物日志") || cmd.StartsWith("礼物记录")) return await GetGiftLogsAsync(ctx);

            return "未知指令。可用：礼物商店、购买礼物、我的背包、送礼物、礼物日志。";
        }

        private async Task<string> GetShopListAsync()
        {
            var gifts = await GiftStoreItem.GetValidGiftsAsync();
            if (gifts.Count == 0) return "商店目前空空如也。";

            var sb = new StringBuilder();
            sb.AppendLine("🎁 【礼物商店】");
            foreach (var g in gifts)
            {
                string typeStr = g.GiftType == 2 ? " [高级]" : "";
                sb.AppendLine($"- {g.GiftName}：{g.GiftCredit} 积分{typeStr}");
            }
            sb.AppendLine("\n💡 发送：购买礼物 <名称> [数量]");
            return sb.ToString();
        }

        private async Task<string> BuyGiftAsync(IPluginContext ctx, string[] args)
        {
            if (args.Length == 0) return "请输入要购买的礼物名称。";
            string giftName = args[0].Trim();
            int count = 1;
            if (args.Length > 1 && int.TryParse(args[1], out int c)) count = Math.Max(1, c);

            var gift = await GiftStoreItem.GetByNameAsync(giftName);
            if (gift == null) return $"找不到礼物【{giftName}】。";

            long totalCost = gift.GiftCredit * count;
            long botUin = long.TryParse(ctx.BotId, out var b) ? b : 0;
            long groupId = long.TryParse(ctx.GroupId, out var g) ? g : 0;
            long userId = long.TryParse(ctx.UserId, out var u) ? u : 0;

            long userCredit = await UserInfo.GetCreditAsync(botUin, groupId, userId);

            if (userCredit < totalCost)
                return $"您的积分不足。购买 {count} 个【{gift.GiftName}】需要 {totalCost} 积分，您当前只有 {userCredit} 积分。";

            // 扣除积分
            var user = await UserInfo.LoadAsync(userId);
            var minusRes = await UserInfo.AddCreditAsync(botUin, groupId, ctx.GroupName ?? "", userId, user?.Name ?? "", -totalCost, $"购买礼物：{gift.GiftName}*{count}");
            
            if (minusRes.Result == -1) return "购买失败，请稍后再试。";

            // 加入背包
            var backpackItem = await GiftBackpack.GetItemAsync(ctx.UserId, gift.Id);
            if (backpackItem == null)
            {
                backpackItem = new GiftBackpack { UserId = ctx.UserId, GiftId = gift.Id, ItemCount = count };
                await backpackItem.InsertAsync();
            }
            else
            {
                backpackItem.ItemCount += count;
                await backpackItem.UpdateAsync();
            }

            return $"🛍️ 购买成功！获得【{gift.GiftName}】x{count}，消耗 {totalCost} 积分。剩余积分：{minusRes.CreditValue}";
        }

        private async Task<string> GetBackpackAsync(IPluginContext ctx)
        {
            var items = await GiftBackpack.GetUserBackpackAsync(ctx.UserId);
            if (items.Count == 0) return "您的背包里还没有任何礼物，快去商店看看吧！";

            var sb = new StringBuilder();
            sb.AppendLine("🎒 【我的礼物背包】");
            foreach (var item in items)
            {
                var gift = (await GiftStoreItem.QueryWhere($"Id = {item.GiftId}", (System.Data.IDbTransaction?)null)).FirstOrDefault();
                if (gift != null)
                {
                    sb.AppendLine($"- {gift.GiftName} x{item.ItemCount}");
                }
                else
                {
                    sb.AppendLine($"- 未知礼物(ID:{item.GiftId}) x{item.ItemCount}");
                }
            }
            sb.AppendLine("\n💡 发送：送礼物 @某人 <名称> [数量]");
            return sb.ToString();
        }

        private async Task<string> SendGiftAsync(IPluginContext ctx, string[] args)
        {
            // 预期格式: 送礼物 @用户 礼物名 [数量]
            // args 可能包含 @用户, 礼物名, [数量]
            // 如果是艾特，args[0] 可能是 @用户
            if (args.Length < 1) return "命令格式：送礼物 @用户 <礼物名称> [数量]";

            // 获取目标用户
            string targetUserId = "";
            string giftName = "";
            int count = 1;

            if (ctx.MentionedUsers != null && ctx.MentionedUsers.Count > 0)
            {
                targetUserId = ctx.MentionedUsers[0].UserId;
                
                // 礼物名称应该是第一个非艾特的参数
                int nameIdx = 0;
                while (nameIdx < args.Length && (args[nameIdx].StartsWith("[CQ:at") || args[nameIdx].StartsWith("@")))
                {
                    nameIdx++;
                }

                if (nameIdx < args.Length) giftName = args[nameIdx];
                if (nameIdx + 1 < args.Length && int.TryParse(args[nameIdx + 1], out int c)) count = Math.Max(1, c);
            }
            else
            {
                // 如果没有艾特，尝试从 args 获取
                if (args.Length >= 2)
                {
                    targetUserId = args[0];
                    giftName = args[1];
                    if (args.Length >= 3 && int.TryParse(args[2], out int c)) count = Math.Max(1, c);
                }
            }

            if (string.IsNullOrEmpty(targetUserId)) return "请艾特或输入要赠送的目标用户。";
            if (targetUserId == ctx.UserId) return "不能给自己送礼物哦。";
            if (string.IsNullOrEmpty(giftName)) return "请输入要赠送的礼物名称。";

            var gift = await GiftStoreItem.GetByNameAsync(giftName);
            if (gift == null) return $"找不到礼物【{giftName}】。";

            // 检查背包
            var backpackItem = await GiftBackpack.GetItemAsync(ctx.UserId, gift.Id);
            if (backpackItem == null || backpackItem.ItemCount < count)
            {
                return $"您的背包里没有足够的【{giftName}】。当前拥有：{(backpackItem?.ItemCount ?? 0)}";
            }

            // 执行赠送
            backpackItem.ItemCount -= count;
            await backpackItem.UpdateAsync();

            // 记录日志
            long botUin = long.TryParse(ctx.BotId, out var b) ? b : 0;
            long groupId = long.TryParse(ctx.GroupId, out var g) ? g : 0;
            long userId = long.TryParse(ctx.UserId, out var u) ? u : 0;
            long targetUid = long.TryParse(targetUserId, out var tu) ? tu : 0;

            var sender = await UserInfo.LoadAsync(userId);
            var receiver = await UserInfo.LoadAsync(targetUid);
            
            var record = new GiftRecord
            {
                BotUin = botUin,
                GroupId = groupId,
                GroupName = ctx.GroupName ?? "",
                UserId = ctx.UserId,
                UserName = sender?.Name ?? "神秘人",
                GiftUserId = targetUserId,
                GiftUserName = receiver?.Name ?? "心仪的Ta",
                GiftId = gift.Id,
                GiftName = gift.GiftName,
                GiftCount = count,
                GiftCredit = gift.GiftCredit
            };
            await record.InsertAsync();

            // 给对方加分 (可选逻辑，根据原系统，赠送会给对方加分)
            long creditAdd = (gift.GiftCredit * count) / 2;
            await UserInfo.AddCreditAsync(botUin, groupId, ctx.GroupName ?? "", targetUid, receiver?.Name ?? "", creditAdd, $"收到礼物：{gift.GiftName}*{count}");

            return $"🎁 赠送成功！你向 {receiver?.Name ?? targetUserId} 赠送了【{gift.GiftName}】x{count}。";
        }

        private async Task<string> GetGiftLogsAsync(IPluginContext ctx)
        {
            var logs = await GiftRecord.QueryWhere("UserId = @p1 OR GiftUserId = @p1 ORDER BY InsertDate DESC", (System.Data.IDbTransaction?)null, GiftRecord.SqlParams(("@p1", ctx.UserId)));
            if (logs.Count == 0) return "暂无礼物往来记录。";

            var sb = new StringBuilder();
            sb.AppendLine("📜 【近期礼物记录】");
            foreach (var log in logs.Take(10))
            {
                string action = log.UserId == ctx.UserId ? $"送给 {log.GiftUserName}" : $"收到 {log.UserName} 的";
                sb.AppendLine($"[{log.InsertDate:MM-dd HH:mm}] {action} 【{log.GiftName}】x{log.GiftCount}");
            }
            return sb.ToString();
        }

        public Task StopAsync() => Task.CompletedTask;
    }
}
