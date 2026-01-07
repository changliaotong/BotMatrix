using System.Text.RegularExpressions;
using Microsoft.Data.SqlClient;
using BotWorker.Bots.Entries;
using BotWorker.Bots.Models.Office;
using BotWorker.Bots.Users;
using BotWorker.Common;
using BotWorker.Core.MetaDatas;

namespace BotWorker.Groups
{
    public partial class GroupMember : MetaData<GroupMember>
    {
        public override string TableName => "GroupMember";
        public override string KeyField => "GroupId";
        public override string KeyField2 => "UserId";

        //充值/扣除 积分 金币 黑金币 紫币 游戏币等
        public static string AddCoinsRes(long botUin, long groupId, string groupName, long qq, string name, string cmdName, string cmdPara, string cmdPara2, string cmdPara3)
        {
            if (!UserInfo.IsOwner(groupId, qq))
                return $"您无权{cmdName}{cmdPara}";

            if (!cmdPara3.IsNum())
                return "数量不正确";

            long coins_oper = long.Parse(cmdPara3);

            long coins_qq = long.Parse(cmdPara2);
            if (!Exists(groupId, coins_qq))
                Append(groupId, coins_qq, "");
            

            if ((cmdPara == "本群积分") | (cmdPara == "积分"))
                cmdPara = "群积分";

            int coins_type = CoinsLog.conisNames.IndexOf(cmdPara);
            long minus_value = coins_oper;
            long minus_credit = coins_oper;
            long credit_group = groupId;

            long credit_value = UserInfo.GetCredit(credit_group, qq);
            if (coins_type == (int)CoinsLog.CoinsType.groupCredit)
            {
                if (!GroupInfo.GetIsCredit(groupId))
                    return $"没有开启本群积分";
            }

            if (cmdName == "充值")
            {
                if (credit_value < minus_value)
                    return $"您有{credit_value}分不足{minus_value}，请先兑换";
            }
            else //扣除
            {
                long coins_value = GetCoins(coins_type, groupId, coins_qq);
                if (coins_value < minus_value)
                    return $"[@:{coins_qq}]{cmdPara}{coins_value}不足{coins_oper}，无法扣除";

                minus_credit = -minus_credit;
                coins_oper = -coins_oper;
            }
            credit_value -= minus_credit;

            //扣除积分 积分记录 增加金币 金币记录
            long coins_last = 0;
            var sqlAddCredit = UserInfo.SqlAddCredit(botUin, credit_group, qq, -minus_credit);
            var sqlCreditHis = CreditLog.SqlHistory(botUin, groupId, groupName, qq, name, -minus_credit, $"{cmdName}{cmdPara}*{coins_oper}");
            var sqlPlusCoins = SqlPlus(CoinsLog.conisFields[coins_type], coins_oper, groupId, coins_qq);
            var sqlCoinsHis = CoinsLog.SqlCoins(botUin, groupId, groupName, coins_qq, "", coins_type, coins_oper, ref coins_last, $"{cmdName}{cmdPara}*{coins_oper}");
            int i = ExecTrans(sqlAddCredit, sqlCreditHis, sqlPlusCoins, sqlCoinsHis);                
            return i == -1
                ? RetryMsg
                : $"{cmdName}{cmdPara}：{coins_oper}成功！\n[@:{coins_qq}]{cmdPara}:{coins_last}\n您：{-minus_credit}分，累计：{credit_value}";
            
        }

        // 本群积分
        public static long GetGroupCredit(long groupId, long qq)
        {
            return GetCoins((int)CoinsLog.CoinsType.groupCredit, groupId, qq);
        }

        // 金币余额
        public static long GetCoins(int coinsType, long groupId, long qq)
        {
            return GetLong(CoinsLog.conisFields[coinsType], groupId, qq);
        }

        // 金币余额
        public static long GetGoldCoins(long groupId, long qq)
        {
            return GetLong("GoldCoins", groupId, qq);
        }

        // 紫币
        public static long GetPurpleCoins(long groupId, long qq)
        {
            return GetLong("PurpleCoins", groupId, qq);
        }

        // 黑金币
        public static long GetBlackCoins(long groupId, long qq)
        {
            return GetLong("BlackCoins", groupId, qq);
        }

        // 游戏币
        public static long GetGameCoins(long groupId, long qq)
        {
            return GetLong("GameCoins", groupId, qq);
        }

        // 加金币/黑金币/紫币/游戏币
        public static int AddCoins(long botUin, long groupId, string groupName, long qq, string name, int coinsType, long coinsAdd, ref long coinsValue, string coinsInfo)
        {
            if (!Exists(groupId, qq) && Append(groupId, qq, name) == -1)
                return -1;

            return ExecTrans(
                SqlPlus(CoinsLog.conisFields[coinsType], coinsAdd, groupId, qq),
                CoinsLog.SqlCoins(botUin, groupId, groupName, qq, name, coinsType, coinsAdd, ref coinsValue, coinsInfo));
        }

        // 扣除金币
        public static int MinusCoins(long botUin, long groupId, string groupName, long qq, string name, int coinsType, long coinsMinus, ref long coinsValue, string coinsInfo)
        {
            return AddCoins(botUin, groupId, groupName, qq, name, coinsType, -(coinsMinus), ref coinsValue, coinsInfo);
        }

        // 虚拟币转账
        public static int TransferCoins(long botUin, long groupId, string groupName, long qq, string name, long qqTo, int coinsType, long coinsMinus, long coinsAdd, ref long coinsValue, ref long coinsValue2)
        {
            if (!Exists(groupId, qqTo) && Append(groupId, qqTo, "") == -1)
                return -1;

            return ExecTrans(
                SqlPlus(CoinsLog.conisFields[coinsType], -coinsMinus, groupId, qq),
                SqlPlus(CoinsLog.conisFields[coinsType], coinsAdd, groupId, qqTo),
                CoinsLog.SqlCoins(botUin, groupId, groupName, qq, name, coinsType, -coinsMinus, ref coinsValue, $"{CoinsLog.conisNames[coinsType]}转出:{qqTo}"),
                CoinsLog.SqlCoins(botUin, groupId, groupName, qqTo, "", coinsType, coinsAdd, ref coinsValue2, $"{CoinsLog.conisNames[coinsType]}转入：{qq}"));
        }

        public static (string, SqlParameter[]) SqlSaveCredit(long groupId, long userId, long creditSave)
        {
            return SqlSetValues($"GroupCredit = GroupCredit - ({creditSave}), SaveCredit = ISNULL(SaveCredit, 0) + ({creditSave})", groupId, userId);
        }

        public static (string, SqlParameter[]) SqlAddCredit(long groupId, long userId, long creditAdd)
        {
            return Exists(groupId, userId)
                ? SqlPlus("GroupCredit", creditAdd, groupId, userId)
                : SqlInsert([
                                new Cov("GroupId", groupId),
                                new Cov("UserId", userId),
                                new Cov("GroupCredit", creditAdd),
                            ]);
        }

        // 添加群成员
        public static int Append(long groupId, long userId, string name, string displayName = "", long groupCredit = 50, string confirmCode = "")
        {
            if (userId.In(2107992324, 3677524472, 3662527857, 2174158062, 2188157235, 3375620034, 1611512438, 3227607419, 3586811032,
                3835195413, 3527470977, 3394199803, 2437953621, 3082166471, 2375832958, 1807139582, 2704647312, 1420694846, 3788007880)) return 0;

            var sql = Exists(groupId, userId)
                            ? SqlSetValues($"UserName = {name.Quotes()}, DisplayName = {displayName.Quotes()}, ConfirmCode = {confirmCode.Quotes()}, Status = 1", groupId, userId)
                            : SqlInsert([
                                            new Cov("GroupId", groupId),
                                            new Cov("UserId", userId),
                                            new Cov("UserName", name),
                                            new Cov("DisplayName", displayName),
                                            new Cov("GroupCredit", groupCredit),
                                            new Cov("ConfirmCode", confirmCode),
                                        ]);
            return Exec(sql);
        }

        //上下分
        public static string GetShangFen(long botUin, long groupId, string groupName, long userId, string cmdName, string cmdPara)
        {
            if (!GroupInfo.IsOwner(groupId, userId) || !BotInfo.IsAdmin(botUin, userId))
                return OwnerOnlyMsg;

            if (Income.Total(userId) < 400)            
                return "您无权使用此命令，请联系客服";
            
            if (BotInfo.GetIsCredit(botUin) || GroupInfo.GetIsCredit(groupId))
            {
                long creditQQ = 0;
                string regexShangFen;
                if (cmdPara.IsMatch(Regexs.CreditParaAt))
                    regexShangFen = Regexs.CreditParaAt;
                else if (cmdPara.IsMatch(Regexs.CreditParaAt2))
                    regexShangFen = Regexs.CreditParaAt2;
                else if (cmdPara.IsMatch(Regexs.CreditPara))
                    regexShangFen = Regexs.CreditPara;
                else
                    return $"格式：{cmdName} + QQ + 数量\n例如：{cmdName} {{客服QQ}} 5000";

                long creditAdd = 0;

                //分析命令
                foreach (Match match in cmdPara.Matches(regexShangFen))
                {
                    creditQQ = match.Groups["UserId"].Value.AsLong();
                    creditAdd = match.Groups["credit"].Value.AsLong();
                }

                if (creditAdd < 10)
                    return $"至少{(cmdName == "上分" ? "上" : "下")}10分";

                var creditValue = UserInfo.GetCredit(groupId, creditQQ);

                if (cmdName == "下分")
                {
                    if (creditValue < creditAdd)
                        return $"对方只有{creditValue}分";
                    creditAdd = -creditAdd;
                }

                var sql = UserInfo.SqlAddCredit(botUin, groupId, creditQQ, creditAdd);
                var sql2 = CreditLog.SqlHistory(botUin, groupId, groupName, creditQQ, "", creditAdd, cmdName);
                int i = ExecTrans(sql, sql2);
                return i == -1
                    ? RetryMsg
                    : $"[@:{creditQQ}] {cmdName}成功！\n积分：{creditAdd}，累计：{creditValue + creditAdd}";
            }
            else
                return $"此群未开通本群积分，不能上下分";
        }

    }
}
