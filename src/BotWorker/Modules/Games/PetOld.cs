using System.Data;
using BotWorker.Domain.Entities;
using BotWorker.Common;
using BotWorker.Common.Extensions;
using BotWorker.Domain.Interfaces;
using System.Text.RegularExpressions;

namespace BotWorker.Modules.Games
{
    [BotPlugin(
        Id = "game.pet.old",
        Name = "经典宠物系统",
        Version = "1.0.0",
        Author = "Matrix",
        Description = "经典的买卖宠物系统，支持身价排行、赎身等互动",
        Category = "Games"
    )]
    public class PetOldPlugin : IPlugin
    {
        public async Task InitAsync(IRobot robot)
        {
            await robot.RegisterSkillAsync(new SkillCapability
            {
                Name = "经典宠物",
                Commands = ["我的宠物", "身价榜", "我的身价", "买入", "赎身"],
                Description = "经典宠物买卖系统"
            }, HandlePetCommandAsync);
        }

        public Task StopAsync() => Task.CompletedTask;

        private async Task<string> HandlePetCommandAsync(IPluginContext ctx, string[] args)
        {
            var cmd = ctx.RawMessage.Trim().Split(' ')[0];
            var cmdPara = string.Join(" ", args);
            var userId = long.Parse(ctx.UserId);
            var groupId = long.Parse(ctx.GroupId ?? "0");
            var botId = long.Parse(ctx.BotId);

            if (cmd == "我的宠物")
                return await BuyFriends.GetMyPetListAsync(groupId, groupId, userId);

            if (cmd == "身价榜")
                return await BuyFriends.GetPriceListAsync(groupId, groupId, userId);

            if (cmd == "我的身价")
                return await BuyFriends.GetMyPriceListAsync(groupId, groupId, userId);

            if (cmd == "买入")
            {
                // 检查是否是买入宠物（参数包含QQ号或@）
                if (Regex.IsMatch(cmdPara, Regexs.CreditParaAt) || 
                    Regex.IsMatch(cmdPara, Regexs.CreditParaAt2) || 
                    Regex.IsMatch(cmdPara, Regexs.CreditPara))
                {
                    // 如果参数包含“积分”或“禁言卡”，则不属于此插件处理
                    if (cmdPara.Contains("积分") || cmdPara.Contains("禁言卡") || cmdPara.Contains("飞机票") || cmdPara.Contains("道具"))
                        return string.Empty;

                    return await BuyFriends.GetBuyPetAsync(botId, groupId, groupId, ctx.Group?.GroupName ?? "", userId, ctx.UserName, cmdPara);
                }
                return string.Empty;
            }

            if (cmd == "赎身")
            {
                if (ctx.Group == null || !ctx.Group.IsPet)
                    return BuyFriends.InfoClosed;

                // 以当前主人购买时的价格成交，对方只能得到80%，系统扣除20%
                long currMaster = await BuyFriends.GetCurrMasterAsync(groupId, userId);
                if (currMaster == userId || currMaster == 0)
                    return "您已是自由身，无需赎身";

                long buyPrice = await BuyFriends.GetBuyPriceAsync(groupId, userId);
                long creditAdd = buyPrice;
                long creditMinus = buyPrice * 12 / 10;
                
                if (ctx.User != null && ctx.User.IsSuper)
                    creditMinus = creditMinus * 22 / 10;

                long creditValue = await UserInfo.GetCreditAsync(botId, groupId, userId);
                if (creditValue < creditMinus)
                    return $"您的积分{creditValue}不足{creditMinus}";

                if (!ctx.IsConfirm)
                {
                    if (ctx is PluginContext pluginCtx)
                        return await pluginCtx.ConfirmMessage($"赎身需扣分：-{creditMinus}");
                    return $"赎身需扣分：-{creditMinus}，请发送“确认”继续";
                }

                int res = await BuyFriends.DoFreeMeAsync(botId, groupId, ctx.Group.GroupName, userId, ctx.UserName, currMaster, creditMinus, creditAdd);
                if (res == -1)
                    return "操作失败，请重试";

                long currentCredit = await UserInfo.GetCreditAsync(botId, groupId, userId);
                long masterCredit = await UserInfo.GetCreditAsync(botId, groupId, currMaster);
                return $"✅ 赎身成功！\n[@:{currMaster}]积分：+{creditAdd}，累计：{masterCredit}\n您的积分：-{creditMinus}，累计：{currentCredit}";
            }

            return string.Empty;
        }
    }

    [Table("BuyFriends")]
    public class BuyFriends
    {
        private static IBuyFriendsRepository Repository => 
            BotMessage.ServiceProvider?.GetRequiredService<IBuyFriendsRepository>() 
            ?? throw new InvalidOperationException("IBuyFriendsRepository not registered");

        [ExplicitKey]
        public int Id { get; set; }
        public int PrevId { get; set; }
        public long BotUin { get; set; }
        public long GroupId { get; set; }
        public long UserId { get; set; }
        public long FriendId { get; set; }
        public long Fromid { get; set; }
        public long BuyPrice { get; set; }
        public long SellPrice { get; set; }
        public int IsValid { get; set; }
        public DateTime InsertDate { get; set; } = DateTime.Now;
        public DateTime? SellDate { get; set; }
        public long? SellTo { get; set; }

        public const string InfoClosed = "宠物系统已关闭";

        // 买入宠物
        public static string GetBuyPet(long botQQ, long _groupId, long groupId, string groupName, long qq, string name, string cmdPara)
            => GetBuyPetAsync(botQQ, _groupId, groupId, groupName, qq, name, cmdPara).GetAwaiter().GetResult();

        public static async Task<string> GetBuyPetAsync(long botQQ, long _groupId, long groupId, string groupName, long qq, string name, string cmdPara)
        {
            if (_groupId == 0)
                groupId = _groupId;
            if (!await GroupInfo.GetIsPetAsync(groupId))
                return InfoClosed;

            if (cmdPara == "")
                return "命令格式：买入 + qq + 积分\n例如：买入 {客服QQ} 5000";


            string regex_reward;
            if (cmdPara.IsMatch(Regexs.CreditParaAt))
                regex_reward = Regexs.CreditParaAt;
            else if (cmdPara.IsMatch(Regexs.CreditParaAt2))
                regex_reward = Regexs.CreditParaAt2;
            else if (cmdPara.IsMatch(Regexs.CreditPara))
                regex_reward = Regexs.CreditPara;
            else
                return $"格式：买入 + qq + 积分\n例如：买入 {BotInfo.CrmUin} 5000";

            //分析命令
            long friendQQ = cmdPara.RegexGetValue(regex_reward, "UserId").AsLong();
            long buyCredit = cmdPara.RegexGetValue(regex_reward, "credit").AsLong();

            long sellPrice = await GetSellPriceAsync(groupId, friendQQ);
            long fromQQ = await GetCurrMasterAsync(groupId, friendQQ);
            long petCount = await GetPetCountAsync(groupId, qq);


            long creditValue = await UserInfo.GetCreditAsync(botQQ, groupId, qq);
            if (creditValue < buyCredit)
                return $"您的积分{creditValue}不足{buyCredit}";

            if (buyCredit < sellPrice)
                return $"至少要出{sellPrice}才能买TA";

            if (await UserInfo.GetIsSuperAsync(qq) | !await UserInfo.GetIsSuperAsync(fromQQ))
                sellPrice = buyCredit;

            int i = await DoBuyPetAsync(botQQ, groupId, groupName, qq, name, fromQQ, friendQQ, sellPrice, buyCredit);
            if (i == -1)
                return "操作失败，请重试";

            long creditSell = buyCredit * 8 / 10;
            long creditFriendGet = buyCredit / 10;
            long currentCredit = await UserInfo.GetCreditAsync(botQQ, groupId, qq);
            return $"✅ 您的宠物+1={petCount + 1}了！\n萌宠[@:{friendQQ}]+{creditFriendGet}分\n卖家[@:{fromQQ}] +{creditSell}分\n积分：-{sellPrice}分 累计：{currentCredit}";
        }

        // 宠物主人
        public static async Task<long> GetCurrMasterAsync(long group_id, long friend_qq)
        {
            return await Repository.GetCurrMasterAsync(group_id, friend_qq);
        }

        public static long GetCurrMaster(long group_id, long friend_qq)
        {
            return GetCurrMasterAsync(group_id, friend_qq).GetAwaiter().GetResult();
        }

        /// 得到某人的当前市场价格
        public static async Task<long> GetSellPriceAsync(long groupId, long friendId)
        {
            return await Repository.GetSellPriceAsync(groupId, friendId);
        }

        public static long GetSellPrice(long groupId, long friendId)
        {
            return GetSellPriceAsync(groupId, friendId).GetAwaiter().GetResult();
        }

        // 得到某人购买价格
        public static async Task<long> GetBuyPriceAsync(long groupId, long friendId)
        {
            return await Repository.GetBuyPriceAsync(groupId, friendId);
        }

        public static long GetBuyPrice(long groupId, long friendId)
        {
            return GetBuyPriceAsync(groupId, friendId).GetAwaiter().GetResult();
        }

        // 得到buyid
        public static async Task<int> GetBuyIdAsync(long groupId, long friendQQ)
        {
            return await Repository.GetBuyIdAsync(groupId, friendQQ);
        }

        public static int GetBuyId(long groupId, long friendQQ)
        {
            return GetBuyIdAsync(groupId, friendQQ).GetAwaiter().GetResult();
        }

        // 宠物数量
        public static async Task<long> GetPetCountAsync(long groupId, long qq)
        {
            return await Repository.GetPetCountAsync(groupId, qq);
        }

        public static long GetPetCount(long groupId, long qq)
        {
            return GetPetCountAsync(groupId, qq).GetAwaiter().GetResult();
        }

        // 身价榜
        public static string GetPriceList(long _groupId, long groupId, long userId, int topN = 3)
            => GetPriceListAsync(_groupId, groupId, userId, topN).GetAwaiter().GetResult();

        public static async Task<string> GetPriceListAsync(long _groupId, long groupId, long userId, int topN = 3)
        {
            if (_groupId == 0)
                groupId = _groupId;
            if (!await GroupInfo.GetIsPetAsync(groupId))
                return InfoClosed;

            var list = await Repository.GetPriceListAsync(groupId, topN);
            var sb = new System.Text.StringBuilder();
            for (int i = 0; i < list.Count; i++)
            {
                sb.Append($"【第{i + 1}名】 [@:{list[i].FriendId}] 身价：{list[i].SellPrice}\n");
            }
            string res = sb.ToString();
            if (!res.Contains(userId.ToString()))
                res += "{身价排名}";
            return res;
        }

        // 我的身价
        public static string GetMyPriceList(long _groupId, long groupId, long userId, int topN = 3)
            => GetMyPriceListAsync(_groupId, groupId, userId, topN).GetAwaiter().GetResult();

        public static async Task<string> GetMyPriceListAsync(long _groupId, long groupId, long userId, int topN = 3)
        {
            if (_groupId == 0)
                groupId = _groupId;
            if (!await GroupInfo.GetIsPetAsync(groupId))
                return InfoClosed;

            long myPirce = await GetSellPriceAsync(groupId, userId);
            
            if (groupId == 0)
            {
                var list = await Repository.GetMyPriceListAsync(userId, topN);
                var sb = new System.Text.StringBuilder();
                for (int i = 0; i < list.Count; i++)
                {
                    sb.Append($"【{i + 1}】 群：{list[i].GroupId} 身价：{list[i].SellPrice}\n");
                }
                return sb.ToString();
            }
            else
            {
                int rank = await Repository.GetRankAsync(groupId, myPirce);
                return $"【第{rank}名】 [@:{userId}] 身价：{myPirce}";
            }
        }

        // 买入宠物
        public static async Task<int> DoBuyPetAsync(long botUin, long groupId, string groupName, long qq, string name, long fromQQ, long friendQQ, long sellPrice, long buyCredit)
        {
            int prev_id = await GetBuyIdAsync(groupId, friendQQ);
            if (!await UserInfo.ExistsAsync(friendQQ))
                await UserInfo.AppendUserAsync(botUin, groupId, friendQQ, "");

            using var conn = Repository.CreateConnection();
            conn.Open();
            using var trans = conn.BeginTransaction();
            try
            {
                var res1 = await UserInfo.AddCreditAsync(botUin, groupId, groupName, qq, name, -buyCredit, $"购买：{friendQQ}", trans);
                if (res1.Result == -1) throw new Exception("买家减分失败");

                var res2 = await UserInfo.AddCreditAsync(botUin, groupId, groupName, fromQQ, "", sellPrice * 8 / 10, $"卖出：{friendQQ}", trans);
                if (res2.Result == -1) throw new Exception("卖家加分失败");

                var res3 = await UserInfo.AddCreditAsync(botUin, groupId, groupName, friendQQ, "", sellPrice * 1 / 10, $"被转卖：{fromQQ}->{qq}", trans);
                if (res3.Result == -1) throw new Exception("宠物加分失败");

                var petHis = new BuyFriends
                {
                    PrevId = prev_id,
                    BotUin = botUin,
                    GroupId = groupId,
                    UserId = qq,
                    FriendId = friendQQ,
                    Fromid = fromQQ,
                    BuyPrice = sellPrice, // buyPrice is actually sellPrice in the original code logic
                    SellPrice = buyCredit * 2,
                    IsValid = 1,
                    InsertDate = DateTime.Now
                };
                await Repository.InsertAsync(petHis, trans);

                if (prev_id > 0)
                {
                    var prev = await Repository.GetByIdAsync(prev_id, trans);
                    if (prev != null)
                    {
                        prev.SellDate = DateTime.Now;
                        prev.SellTo = qq;
                        prev.SellPrice = sellPrice;
                        prev.IsValid = 0;
                        await Repository.UpdateAsync(prev, trans);
                    }
                }

                trans.Commit();

                UserInfo.SyncCacheField(qq, groupId, "Credit", res1.CreditValue);
                UserInfo.SyncCacheField(fromQQ, groupId, "Credit", res2.CreditValue);
                UserInfo.SyncCacheField(friendQQ, groupId, "Credit", res3.CreditValue);

                return 0;
            }
            catch (Exception ex)
            {
                trans.Rollback();
                Console.WriteLine($"[DoBuyPet Error] {ex.Message}");
                return -1;
            }
        }

        // 买入宠物
        public static int DoBuyPet(long botUin, long groupId, string groupName, long qq, string name, long fromQQ, long friendQQ, long sellPrice, long buyCredit)
        {
            return DoBuyPetAsync(botUin, groupId, groupName, qq, name, fromQQ, friendQQ, sellPrice, buyCredit).GetAwaiter().GetResult();
        }

        // 赎身
        public static async Task<int> DoFreeMeAsync(long botUin, long groupId, string groupName, long qq, string name, long fromQQ, long creditMinus, long creditAdd)
        {
            int prev_id = await GetBuyIdAsync(groupId, qq);
            using var conn = Repository.CreateConnection();
            conn.Open();
            using var trans = conn.BeginTransaction();
            try
            {
                var res1 = await UserInfo.AddCreditAsync(botUin, groupId, groupName, qq, name, -creditMinus, $"赎身：{fromQQ}", trans);
                if (res1.Result == -1) throw new Exception("赎身减分失败");

                var res2 = await UserInfo.AddCreditAsync(botUin, groupId, groupName, fromQQ, "", creditAdd, $"赎身：{qq}", trans);
                if (res2.Result == -1) throw new Exception("赎身卖家加分失败");

                var petHis = new BuyFriends
                {
                    PrevId = prev_id,
                    BotUin = botUin,
                    GroupId = groupId,
                    UserId = qq,
                    FriendId = qq,
                    Fromid = fromQQ,
                    BuyPrice = creditAdd,
                    SellPrice = 0,
                    IsValid = 0,
                    InsertDate = DateTime.Now
                };
                await Repository.InsertAsync(petHis, trans);

                if (prev_id > 0)
                {
                    var prev = await Repository.GetByIdAsync(prev_id, trans);
                    if (prev != null)
                    {
                        prev.SellDate = DateTime.Now;
                        prev.SellTo = qq;
                        prev.SellPrice = creditAdd;
                        prev.IsValid = 0;
                        await Repository.UpdateAsync(prev, trans);
                    }
                }

                trans.Commit();

                UserInfo.SyncCacheField(qq, groupId, "Credit", res1.CreditValue);
                UserInfo.SyncCacheField(fromQQ, groupId, "Credit", res2.CreditValue);

                return 0;
            }
            catch (Exception ex)
            {
                trans.Rollback();
                Console.WriteLine($"[DoFreeMe Error] {ex.Message}");
                return -1;
            }
        }

        // 赎身
        public static int DoFreeMe(long botUin, long groupId, string groupName, long qq, string name, long fromQQ, long creditMinus, long creditAdd)
        {
            return DoFreeMeAsync(botUin, groupId, groupName, qq, name, fromQQ, creditMinus, creditAdd).GetAwaiter().GetResult();
        }

        // 我的宠物列表
        public static string GetMyPetList(long _groupId, long groupId, long qq, int topN = 3)
            => GetMyPetListAsync(_groupId, groupId, qq, topN).GetAwaiter().GetResult();

        public static async Task<string> GetMyPetListAsync(long _groupId, long groupId, long qq, int topN = 3)
        {
            if (_groupId != 0 & !await GroupInfo.GetIsPetAsync(groupId))
                return InfoClosed;

            var list = await Repository.GetMyPetListAsync(groupId, qq, topN);
            var sb = new System.Text.StringBuilder();
            for (int i = 0; i < list.Count; i++)
            {
                sb.Append($"【第{i + 1}名】 [@:{list[i].FriendId}] 身价：{list[i].SellPrice}\n");
            }
            string res = sb.ToString();
            return $"{res}当前宠物状态：您买入的宠物数量：{await GetPetCountAsync(groupId, qq)}";
        }
    }
}
