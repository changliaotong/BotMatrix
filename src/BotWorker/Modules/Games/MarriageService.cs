using BotWorker.Domain.Interfaces;
using System.Text;

namespace BotWorker.Modules.Games
{
    [BotPlugin(
        Id = "game.marriage.v2",
        Name = "婚姻与育儿",
        Version = "1.0.0",
        Author = "Matrix",
        Description = "完善的虚拟社交：求婚结婚、甜蜜互动",
        Category = "Games"
    )]
    public class MarriageService : IPlugin
    {
        public List<Intent> Intents => [
            new() { Name = "求婚", Keywords = ["求婚"] },
            new() { Name = "结婚", Keywords = ["接受求婚", "拒绝求婚"] },
            new() { Name = "离婚", Keywords = ["我要离婚"] },
            new() { Name = "婚姻状态", Keywords = ["我的婚姻", "婚姻面板"] }
        ];

        public async Task StopAsync() => await Task.CompletedTask;

        public async Task InitAsync(IRobot robot)
        {
            await EnsureTablesCreatedAsync();
            await robot.RegisterSkillAsync(new SkillCapability
            {
                Name = "婚姻系统",
                Commands = ["求婚", "接受求婚", "拒绝求婚", "我要离婚", "办理结婚证", "办理离婚证", "我的婚姻", "婚姻面板", "发喜糖", "发红包", "吃喜糖", "购买婚纱", "购买婚戒", "我的对象", "另一半签到", "另一半抢楼", "另一半抢红包", "领取结婚福利", "我的甜蜜爱心", "赠送甜蜜爱心", "使用甜蜜抽奖", "甜蜜爱心说明"],
                Description = "【求婚 @用户】开启浪漫；【我的婚姻】查看状态；结婚后可【发喜糖】"
            }, HandleCommandAsync);
        }

        private async Task EnsureTablesCreatedAsync()
        {
            await UserMarriage.EnsureTableCreatedAsync();
            await MarriageProposal.EnsureTableCreatedAsync();
            await WeddingItem.EnsureTableCreatedAsync();
            await SweetHeart.EnsureTableCreatedAsync();
        }

        private async Task<string> HandleCommandAsync(IPluginContext ctx, string[] args)
        {
            var cmd = ctx.RawMessage.Trim().Split(' ')[0];
            try
            {
                return cmd switch
                {
                    "求婚" => await ProposeAsync(ctx, args),
                    "接受求婚" or "办理结婚证" => await AcceptProposalAsync(ctx),
                    "拒绝求婚" => await RejectProposalAsync(ctx),
                    "我要离婚" or "办理离婚证" => await DivorceAsync(ctx),
                    "我的婚姻" or "婚姻面板" => await GetMarriageStatusAsync(ctx),
                    "发喜糖" => await SendSweetsAsync(ctx),
                    "发红包" => await SendRedPacketAsync(ctx),
                    "吃喜糖" => await EatSweetsAsync(ctx),
                    "购买婚纱" => await BuyWeddingItemAsync(ctx, "dress"),
                    "购买婚戒" => await BuyWeddingItemAsync(ctx, "ring"),
                    "我的对象" => await GetSpouseInfoAsync(ctx),
                    "另一半签到" => await SpouseActionAsync(ctx, "签到"),
                    "另一半抢楼" => await SpouseActionAsync(ctx, "抢楼"),
                    "另一半抢红包" => await SpouseActionAsync(ctx, "抢红包"),
                    "领取结婚福利" => await GetMarriageWelfareAsync(ctx),
                    "我的甜蜜爱心" => await GetSweetHeartsAsync(ctx),
                    "赠送甜蜜爱心" => await GiftSweetHeartsAsync(ctx, args),
                    "使用甜蜜抽奖" => await SweetHeartLuckyDrawAsync(ctx),
                    "甜蜜爱心说明" => GetSweetHeartHelp(),
                    _ => "未知婚姻指令"
                };
            }
            catch (Exception ex)
            {
                return $"❌ 婚姻登记处系统故障：{ex.Message}";
            }
        }

        #region 婚姻核心逻辑

        private async Task<string> ProposeAsync(IPluginContext ctx, string[] args)
        {
            var me = await UserMarriage.GetOrCreateAsync(ctx.UserId);
            if (me.Status == "married") return "你已经结婚了，请先保持忠诚！";

            // 解析被求婚者 (简单处理：假设第一个参数是被求婚者的UserId或通过Ctx获取提到的人)
            if (args.Length == 0) return "你想向谁求婚？请加上 @用户 或输入对方ID。";
            var targetId = args[0].Replace("@", "").Trim(); // 简单模拟

            if (targetId == ctx.UserId) return "你不能向自己求婚。";

            var target = await UserMarriage.GetOrCreateAsync(targetId);
            if (target.Status == "married") return "对方已经名花/草有主了。";

            var proposal = new MarriageProposal { ProposerId = ctx.UserId, RecipientId = targetId };
            await proposal.InsertAsync();

            return $"💍 【{ctx.UserId}】 向 【{targetId}】 发起了浪漫求婚！\n请输入【接受求婚】或【拒绝求婚】。";
        }

        private async Task<string> AcceptProposalAsync(IPluginContext ctx)
        {
            var proposal = await MarriageProposal.GetPendingAsync(ctx.UserId);
            if (proposal == null) return "当前没有人向你求婚。";

            var me = await UserMarriage.GetOrCreateAsync(ctx.UserId);
            var spouse = await UserMarriage.GetOrCreateAsync(proposal.ProposerId);

            if (me.Status == "married" || spouse.Status == "married") return "由于某些原因，求婚失效了（某方已婚）。";

            using var trans = await MetaData.BeginTransactionAsync();
            try
            {
                var now = DateTime.Now;
                string nowStr = now.ToString("yyyy-MM-dd HH:mm:ss");

                // 更新双方状态
                await UserMarriage.UpdateWhereAsync(new { Status = "married", SpouseId = spouse.UserId, MarriageDate = now, UpdatedAt = now }, "UserId = {0}", trans, me.UserId);
                await UserMarriage.UpdateWhereAsync(new { Status = "married", SpouseId = me.UserId, MarriageDate = now, UpdatedAt = now }, "UserId = {0}", trans, spouse.UserId);

                // 更新求婚记录
                await MarriageProposal.UpdateAsync(new { Status = "accepted", UpdatedAt = now }, proposal.Id, null, trans);
                
                MetaData.CommitTransaction(trans);

                // 上报成就
                _ = AchievementPlugin.ReportMetricAsync(ctx.UserId, "marriage.count", 1);
                _ = AchievementPlugin.ReportMetricAsync(proposal.ProposerId, "marriage.count", 1);

                return $"🎉 恭喜！【{me.UserId}】 与 【{spouse.UserId}】 正式结为夫妻！\n愿得一人心，白首不相离。";
            }
            catch (Exception ex)
            {
                MetaData.RollbackTransaction(trans);
                return $"出错了: {ex.Message}";
            }
        }

        private async Task<string> RejectProposalAsync(IPluginContext ctx)
        {
            var proposal = await MarriageProposal.GetPendingAsync(ctx.UserId);
            if (proposal == null) return "当前没有人向你求婚。";

            proposal.Status = "rejected";
            await proposal.UpdateAsync();
            return $"💔 你拒绝了 【{proposal.ProposerId}】 的求婚。";
        }

        private async Task<string> DivorceAsync(IPluginContext ctx)
        {
            var me = await UserMarriage.GetByUserIdAsync(ctx.UserId);
            if (me == null || me.Status != "married") return "你目前还是单身呢。";

            var spouseId = me.SpouseId;
            var now = DateTime.Now;

            using var trans = await MetaData.BeginTransactionAsync();
            try
            {
                await UserMarriage.UpdateWhereAsync(new { Status = "divorced", SpouseId = "", DivorceDate = now, UpdatedAt = now }, "UserId = {0}", trans, ctx.UserId);
                await UserMarriage.UpdateWhereAsync(new { Status = "divorced", SpouseId = "", DivorceDate = now, UpdatedAt = now }, "UserId = {0}", trans, spouseId);
                MetaData.CommitTransaction(trans);
                return $"🥀 缘尽于此。【{ctx.UserId}】 与 【{spouseId}】 已办理离婚手续。";
            }
            catch (Exception ex)
            {
                MetaData.RollbackTransaction(trans);
                return $"出错了: {ex.Message}";
            }
        }

        private async Task<string> GetMarriageStatusAsync(IPluginContext ctx)
        {
            var me = await UserMarriage.GetByUserIdAsync(ctx.UserId);
            if (me == null || me.Status == "single") return "👤 你目前是单身贵族。";

            var sb = new StringBuilder();
            sb.AppendLine("💍 【我的婚姻面板】");
            sb.AppendLine("━━━━━━━━━━━━━━");
            sb.AppendLine($"❤️ 伴侣: {me.SpouseId}");
            sb.AppendLine($"📅 结婚纪念日: {me.MarriageDate:yyyy-MM-dd}");
            sb.AppendLine($"🍬 喜糖数量: {me.SweetsCount}");
            sb.AppendLine($"🧧 红包数量: {me.RedPacketsCount}");
            sb.AppendLine($"💖 甜蜜爱心: {me.SweetHearts}");
            sb.AppendLine("━━━━━━━━━━━━━━");
            return sb.ToString();
        }

        private async Task<string> SendSweetsAsync(IPluginContext ctx)
        {
            var me = await UserMarriage.GetByUserIdAsync(ctx.UserId);
            if (me == null || me.Status != "married") return "只有结婚后才能发喜糖哦。";

            me.SweetsCount++;
            me.SweetHearts += 5;
            await me.UpdateAsync();
            return $"🍬 【{ctx.UserId}】 撒了一大把喜糖！大家快来抢啊！(甜蜜+5)";
        }

        private async Task<string> SendRedPacketAsync(IPluginContext ctx)
        {
            var me = await UserMarriage.GetByUserIdAsync(ctx.UserId);
            if (me == null || me.Status != "married") return "只有结婚后才能发红包哦。";

            me.RedPacketsCount++;
            me.SweetHearts += 10;
            await me.UpdateAsync();
            return $"🧧 【{ctx.UserId}】 发了一个超大红包！恭喜发财！(甜蜜+10)";
        }

        private async Task<string> EatSweetsAsync(IPluginContext ctx)
        {
            var me = await UserMarriage.GetOrCreateAsync(ctx.UserId);
            // 简单模拟抢喜糖
            var lucky = new Random().Next(1, 100);
            if (lucky > 50)
            {
                var points = new Random().Next(10, 50);
                return $"🍬 你抢到了一颗喜糖，真甜！(获得 {points} 积分)";
            }
            return "🍬 哎呀，喜糖被抢光了，下次快一点哦。";
        }

        private async Task<string> BuyWeddingItemAsync(IPluginContext ctx, string type)
        {
            var me = await UserMarriage.GetOrCreateAsync(ctx.UserId);
            var itemName = type == "dress" ? "婚纱" : "婚戒";
            var price = type == "dress" ? 500 : 1000;

            // 检查是否已购买
            var existing = (await WeddingItem.QueryWhere("UserId = {0} AND ItemType = {1}", ctx.UserId, type)).FirstOrDefault();
            if (existing != null) return $"你已经拥有【{(type == "dress" ? "婚纱" : "婚戒")}】了。";

            var item = new WeddingItem { UserId = ctx.UserId, ItemType = type, Name = itemName, Price = price };
            await item.InsertAsync();

            return $"🛍️ 购买成功！你获得了一件浪漫的【{itemName}】。";
        }

        private async Task<string> GetSpouseInfoAsync(IPluginContext ctx)
        {
            var me = await UserMarriage.GetByUserIdAsync(ctx.UserId);
            if (me == null || me.Status != "married") return "你目前还没有对象。";

            return $"❤️ 你的另一半是：【{me.SpouseId}】\n💕 你们已经相爱 { (DateTime.Now - me.MarriageDate).Days } 天了。";
        }

        private async Task<string> SpouseActionAsync(IPluginContext ctx, string action)
        {
            var me = await UserMarriage.GetByUserIdAsync(ctx.UserId);
            if (me == null || me.Status != "married") return "只有结婚后才能为另一半操作。";

            var spouse = await UserMarriage.GetByUserIdAsync(me.SpouseId);
            if (spouse == null) return "找不到配偶信息。";

            spouse.SweetHearts += 2;
            await spouse.UpdateAsync();
            return $"💞 你为 【{me.SpouseId}】 进行了【{action}】，对方获得了 2 点甜蜜爱心！";
        }

        private async Task<string> GetMarriageWelfareAsync(IPluginContext ctx)
        {
            var me = await UserMarriage.GetByUserIdAsync(ctx.UserId);
            if (me == null || me.Status != "married") return "只有已婚人士才能领取福利。";

            var days = (DateTime.Now - me.MarriageDate).Days;
            var reward = 100 + (days * 2); // 结婚时间越长福利越高

            me.SweetHearts += 5;
            await me.UpdateAsync();

            return $"🎁 领取成功！作为已婚人士，你获得了 {reward} 积分和 5 点甜蜜爱心。";
        }

        private async Task<string> GetSweetHeartsAsync(IPluginContext ctx)
        {
            var me = await UserMarriage.GetOrCreateAsync(ctx.UserId);
            return $"💖 你当前拥有 {me.SweetHearts} 点甜蜜爱心。";
        }

        private async Task<string> GiftSweetHeartsAsync(IPluginContext ctx, string[] args)
        {
            if (args.Length == 0) return "请输入要赠送的对象和数量，例如：赠送甜蜜爱心 @用户 10";
            var me = await UserMarriage.GetOrCreateAsync(ctx.UserId);

            var targetId = args[0].Replace("@", "").Trim();
            if (!int.TryParse(args.Length > 1 ? args[1] : "1", out var amount) || amount <= 0) return "请输入正确的赠送数量。";

            if (me.SweetHearts < amount) return $"❌ 你的甜蜜爱心不足，当前只有 {me.SweetHearts} 点。";

            var target = await UserMarriage.GetOrCreateAsync(targetId);

            me.SweetHearts -= amount;
            target.SweetHearts += amount;

            await me.UpdateAsync();
            await target.UpdateAsync();

            await new SweetHeart { SenderId = ctx.UserId, RecipientId = targetId, Amount = amount }.InsertAsync();

            return $"💝 赠送成功！你向 【{targetId}】 赠送了 {amount} 点甜蜜爱心。";
        }

        private async Task<string> SweetHeartLuckyDrawAsync(IPluginContext ctx)
        {
            var me = await UserMarriage.GetOrCreateAsync(ctx.UserId);
            if (me.SweetHearts < 10) return "❌ 抽奖需要 10 点甜蜜爱心，你当前只有 {me.SweetHearts} 点。";

            me.SweetHearts -= 10;
            await me.UpdateAsync();

            var lucky = new Random().Next(1, 100);
            var prize = lucky switch
            {
                > 90 => "超级大奖：500 积分",
                > 70 => "二等奖：200 积分",
                > 40 => "三等奖：50 积分",
                _ => "参与奖：10 积分"
            };

            return $"🎲 抽奖结果：【{prize}】！感谢参与。";
        }

        private string GetSweetHeartHelp()
        {
            var sb = new StringBuilder();
            sb.AppendLine("💖 【甜蜜爱心系统说明】");
            sb.AppendLine("━━━━━━━━━━━━━━");
            sb.AppendLine("1. 甜蜜爱心是衡量玩家魅力和社交活跃度的指标。");
            sb.AppendLine("2. 获取途径：发喜糖(+5)、发红包(+10)、为伴侣操作(+2)、领取结婚福利(+5)等。");
            sb.AppendLine("3. 每日活跃和与他人互动也能增加甜蜜爱心。");
            sb.AppendLine("4. 用途：可以用于【使用甜蜜抽奖】(10点/次)或【赠送甜蜜爱心】给心仪的TA。");
            sb.AppendLine("━━━━━━━━━━━━━━");
            sb.AppendLine("💡 提示：多在群里活跃，你的魅力值会不断提升哦。");
            return sb.ToString();
        }

        #endregion
    }
}
