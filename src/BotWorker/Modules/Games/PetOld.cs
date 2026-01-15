using System.Data;
using BotWorker.Domain.Entities;
using BotWorker.Common;
using BotWorker.Common.Extensions;
using BotWorker.Infrastructure.Persistence.ORM;
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
                return await PetOld.GetMyPetListAsync(groupId, groupId, userId);

            if (cmd == "身价榜")
                return await PetOld.GetPriceListAsync(groupId, groupId, userId);

            if (cmd == "我的身价")
                return await PetOld.GetMyPriceListAsync(groupId, groupId, userId);

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

                    return await PetOld.GetBuyPetAsync(botId, groupId, groupId, ctx.Group?.GroupName ?? "", userId, ctx.UserName, cmdPara);
                }
                return string.Empty;
            }

            if (cmd == "赎身")
            {
                if (ctx.Group == null || !ctx.Group.IsPet)
                    return PetOld.InfoClosed;

                // 以当前主人购买时的价格成交，对方只能得到80%，系统扣除20%
                long currMaster = await PetOld.GetCurrMasterAsync(groupId, userId);
                if (currMaster == userId || currMaster == 0)
                    return "您已是自由身，无需赎身";

                long buyPrice = await PetOld.GetBuyPriceAsync(groupId, userId);
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

                int res = await PetOld.DoFreeMeAsync(botId, groupId, ctx.Group.GroupName, userId, ctx.UserName, currMaster, creditMinus, creditAdd);
                if (res == -1)
                    return "操作失败，请重试";

                long currentCredit = await UserInfo.GetCreditAsync(botId, groupId, userId);
                long masterCredit = await UserInfo.GetCreditAsync(botId, groupId, currMaster);
                return $"✅ 赎身成功！\n[@:{currMaster}]积分：+{creditAdd}，累计：{masterCredit}\n您的积分：-{creditMinus}，累计：{currentCredit}";
            }

            return string.Empty;
        }
    }

    public class PetOld : MetaData<PetOld>
    {
        public override string TableName => "BuyFriends";
        public override string KeyField => "Id";

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
                return RetryMsg;

            long creditSell = buyCredit * 8 / 10;
            long creditFriendGet = buyCredit / 10;
            long currentCredit = await UserInfo.GetCreditAsync(botQQ, groupId, qq);
            return $"✅ 您的宠物+1={petCount + 1}了！\n萌宠[@:{friendQQ}]+{creditFriendGet}分\n卖家[@:{fromQQ}] +{creditSell}分\n积分：-{sellPrice}分 累计：{currentCredit}";
        }

        // 宠物主人
        public static async Task<long> GetCurrMasterAsync(long group_id, long friend_qq)
        {
            string res = await GetWhereAsync<string>("UserId", $"GroupId = {group_id} and FriendId = {friend_qq} and IsValid = 1") ?? "";
            return res.AsLong();
        }

        public static long GetCurrMaster(long group_id, long friend_qq)
        {
            return GetCurrMasterAsync(group_id, friend_qq).GetAwaiter().GetResult();
        }

        /// 得到某人的当前市场价格
        public static async Task<long> GetSellPriceAsync(long groupId, long friendId)
        {
            long minPrice = 100;
            string func = IsPostgreSql ? "get_sell_price" : $"{DbName}.dbo.get_sell_price";
            string res = await QueryScalarAsync<string>($"SELECT {func}(SellPrice, InsertDate) AS res FROM {FullName} " +
                               $"WHERE GroupId = {groupId} AND FriendId = {friendId} AND IsValid = 1") ?? "";
            long sellPrice = res == "" ? minPrice : res.AsLong();
            return sellPrice < minPrice ? minPrice : sellPrice;
        }

        public static long GetSellPrice(long groupId, long friendId)
        {
            return GetSellPriceAsync(groupId, friendId).GetAwaiter().GetResult();
        }

        // 得到某人购买价格
        public static async Task<long> GetBuyPriceAsync(long groupId, long friendId)
        {
            return (await GetWhereAsync<string>("BuyPrice", $"GroupId = {groupId} AND FriendId = {friendId} AND IsValid = 1")).AsLong();
        }

        public static long GetBuyPrice(long groupId, long friendId)
        {
            return GetBuyPriceAsync(groupId, friendId).GetAwaiter().GetResult();
        }

        // 得到buyid
        public static async Task<int> GetBuyIdAsync(long groupId, long friendQQ)
        {
            return (await GetWhereAsync<string>(SqlIsNull("Id", "0"), $"GroupId = {groupId} AND FriendId = {friendQQ} AND IsValid = 1")).AsInt();
        }

        public static int GetBuyId(long groupId, long friendQQ)
        {
            return GetBuyIdAsync(groupId, friendQQ).GetAwaiter().GetResult();
        }

        // 宠物数量
        public static async Task<long> GetPetCountAsync(long groupId, long qq)
        {
            return await CountWhereAsync($"GroupId = {groupId} AND UserId = {qq} AND IsValid = 1");
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

            string func = IsPostgreSql ? "get_sell_price" : $"{DbName}.dbo.get_sell_price";
            string res = await QueryResAsync($"SELECT {SqlTop(topN)} FriendId, {func}(SellPrice, InsertDate) AS SellPrice FROM {FullName} " +
                                  $"where GroupId = {groupId} and IsValid = 1 order by SellPrice desc {SqlLimit(topN)}", 
                                  "【第{i}名】 [@:{0}] 身价：{1}\n");
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
            string sql = $"SELECT COUNT(*)+1 AS res FROM {FullName} WHERE GroupId = {groupId} AND IsValid = 1 AND SellPrice > {myPirce}";

            string func = IsPostgreSql ? "get_sell_price" : $"{DbName}.dbo.get_sell_price";
            return groupId == 0
                ?  await QueryResAsync($"SELECT {SqlTop(topN)} GroupId, {func}(SellPrice, InsertDate) AS SellPrice " +
                    $"FROM {FullName} WHERE IsValid = 1 AND FriendId = {userId} ORDER BY SellPrice DESC {SqlLimit(topN)}",
                    "【{i}】 群：{0} 身价：{1}\n")
                : $"【第{await QueryAsync(sql)}名】 [@:{userId}] 身价：{myPirce}";
        }

        // 买入宠物
        public static async Task<int> DoBuyPetAsync(long botUin, long groupId, string groupName, long qq, string name, long fromQQ, long friendQQ, long sellPrice, long buyCredit)
        {
            int prev_id = await GetBuyIdAsync(groupId, friendQQ);
            if (!await UserInfo.ExistsAsync(friendQQ))
                await UserInfo.AppendUserAsync(botUin, groupId, friendQQ, "");

            using var trans = await BeginTransactionAsync();
            try
            {
                var res1 = await UserInfo.AddCreditAsync(botUin, groupId, groupName, qq, name, -buyCredit, $"购买：{friendQQ}", trans);
                if (res1.Result == -1) throw new Exception("买家减分失败");

                var res2 = await UserInfo.AddCreditAsync(botUin, groupId, groupName, fromQQ, "", sellPrice * 8 / 10, $"卖出：{friendQQ}", trans);
                if (res2.Result == -1) throw new Exception("卖家加分失败");

                var res3 = await UserInfo.AddCreditAsync(botUin, groupId, groupName, friendQQ, "", sellPrice * 1 / 10, $"被转卖：{fromQQ}->{qq}", trans);
                if (res3.Result == -1) throw new Exception("宠物加分失败");

                var (sql7, paras7) = SqlPetHis(botUin, prev_id, groupId, qq, friendQQ, fromQQ, sellPrice, buyCredit * 2, 1);
                await ExecAsync(sql7, trans, paras7);

                var (sql8, paras8) = SqlUpdSellInfo(qq, sellPrice, prev_id);
                await ExecAsync(sql8, trans, paras8);

                await trans.CommitAsync();

                UserInfo.SyncCacheField(qq, groupId, "Credit", res1.CreditValue);
                UserInfo.SyncCacheField(fromQQ, groupId, "Credit", res2.CreditValue);
                UserInfo.SyncCacheField(friendQQ, groupId, "Credit", res3.CreditValue);

                return 0;
            }
            catch (Exception ex)
            {
                await trans.RollbackAsync();
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
            using var trans = await BeginTransactionAsync();
            try
            {
                var res1 = await UserInfo.AddCreditAsync(botUin, groupId, groupName, qq, name, -creditMinus, $"赎身：{fromQQ}", trans);
                if (res1.Result == -1) throw new Exception("赎身减分失败");

                var res2 = await UserInfo.AddCreditAsync(botUin, groupId, groupName, fromQQ, "", creditAdd, $"赎身：{qq}", trans);
                if (res2.Result == -1) throw new Exception("赎身卖家加分失败");

                var (sql5, paras5) = SqlPetHis(botUin, prev_id, groupId, qq, qq, fromQQ, creditAdd);
                await ExecAsync(sql5, trans, paras5);

                var (sql6, paras6) = SqlUpdSellInfo(qq, creditAdd, prev_id);
                await ExecAsync(sql6, trans, paras6);

                await trans.CommitAsync();

                UserInfo.SyncCacheField(qq, groupId, "Credit", res1.CreditValue);
                UserInfo.SyncCacheField(fromQQ, groupId, "Credit", res2.CreditValue);

                return 0;
            }
            catch (Exception ex)
            {
                await trans.RollbackAsync();
                Console.WriteLine($"[DoFreeMe Error] {ex.Message}");
                return -1;
            }
        }

        // 赎身
        public static int DoFreeMe(long botUin, long groupId, string groupName, long qq, string name, long fromQQ, long creditMinus, long creditAdd)
        {
            return DoFreeMeAsync(botUin, groupId, groupName, qq, name, fromQQ, creditMinus, creditAdd).GetAwaiter().GetResult();
        }

        // 宠物his sql
        public static (string, IDataParameter[]) SqlPetHis(long botUin, long prevId, long groupId, long qq, long friendQQ, long fromQQ, long buyPrice, long sellPrice = 0, int isValid = 0)
        {
            return SqlInsert([
                new Cov("PrevId", prevId),
                new Cov("BotUin", botUin),
                new Cov("GroupId", groupId),
                new Cov("UserId", qq),
                new Cov("FriendId", friendQQ),
                new Cov("Fromid", fromQQ),
                new Cov("BuyPrice", buyPrice),
                new Cov("SellPrice", sellPrice),
                new Cov("IsValid", isValid),
            ]);
        }

        // 更新卖出信息
        public static (string, IDataParameter[]) SqlUpdSellInfo(long sellTO, long sellPrice, long buyId)
        {
            return SqlSetValues($"SellDate = {SqlDateTime}, SellTo = {sellTO}, SellPrice = {sellPrice}, IsValid = 0", buyId);
        }

        // 我的宠物列表
        public static string GetMyPetList(long _groupId, long groupId, long qq, int topN = 3)
            => GetMyPetListAsync(_groupId, groupId, qq, topN).GetAwaiter().GetResult();

        public static async Task<string> GetMyPetListAsync(long _groupId, long groupId, long qq, int topN = 3)
        {
            if (_groupId != 0 & !await GroupInfo.GetIsPetAsync(groupId))
                return InfoClosed;

            string func = IsPostgreSql ? "get_sell_price" : $"{DbName}.dbo.get_sell_price";
            string sql = $"SELECT {SqlTop(topN)} FriendId, {func}(SellPrice, InsertDate) AS SellPrice FROM {FullName} " +
                         $"WHERE GroupId = {groupId} AND UserId = {qq} AND IsValid = 1 ORDER BY SellPrice DESC {SqlLimit(topN)}";
            string res = await QueryResAsync(sql, "【第{i}名】 [@:{0}] 身价：{1}\n");
            return $"{res}当前宠物状态：您买入的宠物数量：{await GetPetCountAsync(groupId, qq)}";
        }
    }
}
