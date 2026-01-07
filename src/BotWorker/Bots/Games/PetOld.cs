using Microsoft.Data.SqlClient;
using BotWorker.Bots.Entries;
using BotWorker.Common;
using BotWorker.Common.Exts;
using BotWorker.Core.MetaDatas;
using BotWorker.Bots.Users;

namespace BotWorker.Bots.Games
{
    class PetOld : MetaData<PetOld>
    {
        public override string TableName => "BuyFriends";
        public override string KeyField => "Id";

        public const string InfoClosed = "宠物系统已关闭";

        // 买入宠物
        public static string GetBuyPet(long botQQ, long _groupId, long groupId, string groupName, long qq, string name, string cmdPara)
        {
            if (_groupId == 0)
                groupId = _groupId;
            if (!GroupInfo.GetIsPet(groupId))
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

            long sellPrice = GetSellPrice(groupId, friendQQ);
            long fromQQ = GetCurrMaster(groupId, friendQQ);
            long petCount = GetPetCount(groupId, qq);


            long creditValue = UserInfo.GetCredit(groupId, qq);
            if (creditValue < buyCredit)
                return $"您的积分{creditValue}不足{buyCredit}";

            if (buyCredit < sellPrice)
                return $"至少要出{sellPrice}才能买TA";

            if (UserInfo.GetIsSuper(qq) | !UserInfo.GetIsSuper(fromQQ))
                sellPrice = buyCredit;

            int i = DoBuyPet(botQQ, groupId, groupName, qq, name, fromQQ, friendQQ, sellPrice, buyCredit);
            if (i == -1)
                return RetryMsg;

            long creditSell = buyCredit * 8 / 10;
            long creditFriendGet = buyCredit / 10;
            return $"✅ 您的宠物+1={petCount + 1}了！\n萌宠[@:{friendQQ}]+{creditFriendGet}分\n卖家[@:{fromQQ}] +{creditSell}分\n积分：-{sellPrice}分 累计：{{积分}}";
        }

        // 宠物主人
        public static long GetCurrMaster(long group_id, long friend_qq)
        {
            string res = GetWhere($"UserId", $"GroupId = {group_id} and FriendId = {friend_qq} and IsValid = 1");
            return res == "" 
                ? friend_qq 
                : res.AsLong();
        }

        /// 得到某人的当前市场价格
        public static long GetSellPrice(long groupId, long friendId)
        {
            long minPrice = 100;
            string res = Query($"SELECT sz84_robot.dbo.get_sell_price(SellPrice, InsertDate) AS res FROM {FullName} " +
                               $"WHERE GroupId = {groupId} AND FriendId = {friendId} AND IsValid = 1");
            long sellPrice = res == "" ? minPrice : res.AsLong();
            return sellPrice < minPrice ?  minPrice : sellPrice;
        }

        // 得到某人购买价格
        public static long GetBuyPrice(long groupId, long friendId)
        {
            return GetWhere($"BuyPrice", $"GroupId = {groupId} AND FriendId = {friendId} AND IsValid = 1").AsLong();
        }

        // 得到buyid
        public static int GetBuyId(long groupId, long friendQQ)
        {
            return GetWhere($"ISNULL(Id, 0)", $"GroupId = {groupId} AND FriendId = {friendQQ} AND IsValid = 1").AsInt();
        }

        // 宠物数量
        public static long GetPetCount(long groupId, long qq)
        {
            return CountWhere($"GroupId = {groupId} AND UserId = {qq} AND IsValid = 1");
        }

        // 身价榜
        public static string GetPriceList(long _groupId, long groupId, long userId, int topN = 3)
        {
            if (_groupId == 0)
                groupId = _groupId;
            if (!GroupInfo.GetIsPet(groupId))
                return InfoClosed;

            string res = QueryRes($"SELECT TOP {topN} FriendId, sz84_robot.dbo.get_sell_price(SellPrice, InsertDate) AS SellPrice FROM {FullName} " +
                                  $"where GroupId = {groupId} and IsValid = 1 order by SellPrice desc", 
                                  "【第{i}名】 [@:{0}] 身价：{1}\n");
            if (!res.Contains(userId.ToString()))
                res += "{身价排名}";
            return res;
        }

        // 我的身价
        public static string GetMyPriceList(long _groupId, long groupId, long userId, int topN = 3)
        {
            if (_groupId == 0)
                groupId = _groupId;
            if (!GroupInfo.GetIsPet(groupId))
                return InfoClosed;

            long myPirce = GetSellPrice(groupId, userId);
            string sql = $"SELECT COUNT(*)+1 AS res FROM {FullName} WHERE GroupId = {groupId} AND IsValid = 1 AND SellPrice > {myPirce}";

            return groupId == 0
                ?  QueryRes($"SELECT TOP {topN} GroupId, sz84_robot.dbo.get_sell_price(SellPrice, InsertDate) AS SellPrice " +
                    $"FROM {FullName} WHERE IsValid = 1 AND FriendId = {userId} ORDER BY SellPrice DESC",
                    "【{i}】 群：{0} 身价：{1}\n")
                : $"【第{Query(sql)}名】 [@:{userId}] 身价：{myPirce}";
        }

        // 买入宠物
        public static int DoBuyPet(long botUin, long groupId, string groupName, long qq, string name, long fromQQ, long friendQQ, long sellPrice, long buyCredit)
        {
            int prev_id = GetBuyId(groupId, friendQQ);
            if (!UserInfo.Exists(friendQQ))
                UserInfo.AppendUser(botUin, groupId, friendQQ, "");
            var sql  = UserInfo.SqlAddCredit(botUin, groupId, qq, -buyCredit);
            var sql2 = UserInfo.SqlAddCredit(botUin, groupId, fromQQ, sellPrice * 8 / 10);
            var sql3 = UserInfo.SqlAddCredit(botUin, groupId, friendQQ, sellPrice * 1 / 10);
            var sql4  = CreditLog.SqlHistory(botUin, groupId, groupName, qq, name, -buyCredit, $"购买：{friendQQ}");
            var sql5 = CreditLog.SqlHistory(botUin, groupId, groupName, fromQQ, "", sellPrice * 8 / 10, $"卖出：{friendQQ}");
            var sql6 = CreditLog.SqlHistory(botUin, groupId, groupName, friendQQ, "", sellPrice * 1 / 10, $"被转卖：{fromQQ}->{qq}");
            var sql7 = SqlPetHis(prev_id, groupId, qq, friendQQ, fromQQ, sellPrice, buyCredit * 2, 1);
            var sql8 = SqlUpdSellInfo(qq, sellPrice, prev_id);            
            return ExecTrans(sql, sql2, sql3, sql4, sql5, sql6, sql7, sql8);
        }

        // 赎身
        public static int DoFreeMe(long botUin, long groupId, string groupName, long qq, string name, long fromQQ, long creditMinus, long creditAdd)
        {
            int prev_id = GetBuyId(groupId, qq);
            var sql  = UserInfo.SqlAddCredit(botUin, groupId, qq, -creditMinus);
            var sql2 = UserInfo.SqlAddCredit(botUin, groupId, fromQQ, creditAdd);
            var sql3  = CreditLog.SqlHistory(botUin, groupId, groupName, qq, name, -creditMinus, $"赎身：{fromQQ}");
            var sql4 = CreditLog.SqlHistory(botUin, groupId, groupName, fromQQ, "", creditAdd, $"赎身：{qq}");
            var sql5 = SqlPetHis(botUin, prev_id, groupId, qq, qq, fromQQ, creditAdd);
            var sql6 = SqlUpdSellInfo(qq, creditAdd, prev_id);               
            return ExecTrans(sql, sql2, sql3, sql4, sql5, sql6);
        }

        // 宠物his sql
        public static (string, SqlParameter[]) SqlPetHis(long botUin, long prevId, long groupId, long qq, long friendQQ, long fromQQ, long buyPrice, long sellPrice = 0, int isValid = 0)
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
        public static (string, SqlParameter[]) SqlUpdSellInfo(long sellTO, long sellPrice, long buyId)
        {
            return SqlSetValues($"SellDate = GETDATE(), SellTo = {sellTO}, SellPrice = {sellPrice}, IsValid = 0", buyId);
        }

        // 我的宠物列表
        public static string GetMyPetList(long _groupId, long groupId, long qq, int topN = 3)
        {
            if (_groupId != 0 & !GroupInfo.GetIsPet(groupId))
                return InfoClosed;

            string sql = $"SELECT TOP {topN} FriendId, sz84_robot.dbo.get_sell_price(SellPrice, InsertDate) AS SellPrice FROM {FullName} " +
                         $"WHERE GroupId = {groupId} AND UserId = {qq} AND IsValid = 1 ORDER BY SellPrice DESC";
            string res = QueryRes(sql, "【第{i}名】 [@:{0}] 身价：{1}\n");
            return $"{res}您买入萌宠数量：{GetPetCount(groupId, qq)}";
        }
    }
}
