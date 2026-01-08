using Microsoft.Data.SqlClient;
using sz84.Bots.Entries;
using BotWorker.Common;
using BotWorker.Common.Exts;
using BotWorker.Infrastructure.Persistence.ORM;

namespace sz84.Bots.Users
{
    public partial class UserInfo : MetaDataGuid<UserInfo>
    {
        public static decimal GetBalance(long qq)
        {
            return Get<decimal>("balance", qq);
        }

        //增加余额
        public static int AddBalance(long botUin, long groupId, string groupName, long qq, string name, decimal balanceAdd, string balanceInfo)
        {
            Append(botUin, groupId, qq, name, GroupInfo.GetGroupOwner(groupId));
            var balanceValue = GetBalance(qq) + balanceAdd;
            var sql = SqlAddBalance(qq, balanceAdd);
            var sql2 = BalanceLog.SqlLog(botUin, groupId, groupName, qq, name, balanceAdd, balanceInfo);
            int i = ExecTrans(sql, sql2);
            if (i == 0)
            {
                SyncCacheField(qq, "Balance", balanceValue);
            }
            return i;
        }

        //减少余额
        public static int MinusBalance(long botUin, long groupId, string groupName, long qq, string name, decimal balanceMinus, string balanceInfo)
        {
            return AddBalance(botUin, groupId, groupName, qq, name, -balanceMinus, balanceInfo);
        }

        //转账
        public static int Transfer(long botUin, long groupId, string groupName, long qq, string name, long qqTo, string nameTo, decimal balanceMinus, decimal balanceAdd)
        {
            var balanceValue = GetBalance(qq) - balanceMinus;
            var balanceValueTo = GetBalance(qqTo) + balanceAdd;

            var sql = SqlAddBalance(qq, -balanceMinus);
            var sql2 = SqlAddBalance(qqTo, balanceAdd);
            var sql3 = BalanceLog.SqlLog(botUin, groupId, groupName, qq, name, -balanceMinus, $"转账给：{qqTo}");
            var sql4 = BalanceLog.SqlLog(botUin, groupId, groupName, qqTo, nameTo, balanceAdd, $"转账来自：{qq}");
            int i = ExecTrans(sql, sql2, sql3, sql4);
            if (i == 0)
            {
                SyncCacheField(qq, "Balance", balanceValue);
                SyncCacheField(qqTo, "Balance", balanceValueTo);
            }
            return i;
        }

        //转账操作
        public static string GetTransferBalance(long botUin, long groupId, string groupName, long qq, string name, string cmdPara)
        {
            cmdPara = cmdPara.Trim();
            if (cmdPara.NotMatch(Regexs.Transfer))
                return "格式：\n转账 + QQ + 余额\n例如：\n转账 {客服QQ} 9.99";
            long qqTransfer = cmdPara.RegexGetValue(Regexs.Transfer, "UserId").AsLong();
            decimal balanceTransfer = cmdPara.RegexGetValue(Regexs.Transfer, "balance").AsDecimal();
            if (qqTransfer == qq)
                return "不能转给自己";

            if (balanceTransfer < 1)
                return "至少转1.00R";

            decimal balanceMinus = balanceTransfer;
            if (GetBalance(qq) < balanceMinus)
                return $"余额{GetBalance(qq)}不足{balanceMinus}。";

            int i = Transfer(botUin, groupId, groupName, qq, name, qqTransfer, "", balanceMinus, balanceTransfer);
            return i == -1
                ? RetryMsg
                : $"✅ 成功转出：{balanceTransfer}\n[@:{qqTransfer}] 的余额：{GetBalance(qqTransfer)}\n你的余额：{{余额}}";
        }

        //冻结余额
        public static decimal GetFreezeBalance(long qq)
        {
            return Get<decimal>("BalanceFreeze", qq);
        }

        //冻结余额
        public static int FreezeBalance(long botUin, long groupId, string groupName, long qq, string name, decimal balanceFreeze)
        {
            decimal balanceValue = GetBalance(qq);
            if (balanceValue < balanceFreeze)
                return -1;

            decimal freezeValue = GetFreezeBalance(qq);

            var sql = BalanceLog.SqlLog(botUin, groupId, groupName, qq, name, -balanceFreeze, "冻结余额");
            var sql2 = SqlSetValues($"balance = balance - ({balanceFreeze}), BalanceFreeze = isnull(BalanceFreeze,0) + ({balanceFreeze})", qq);
            int i = ExecTrans(sql, sql2);
            if (i == 0)
            {
                SyncCacheField(qq, "Balance", balanceValue - balanceFreeze);
                SyncCacheField(qq, "BalanceFreeze", freezeValue + balanceFreeze);
            }
            return i;
        }

        //解冻余额
        public static int UnfreezeBalance(long botUin, long groupId, string groupName, long qq, string name, decimal balanceUnfreeze)
        {
            decimal freezeValue = GetFreezeBalance(qq);
            if (freezeValue < balanceUnfreeze) return -1;

            decimal balanceValue = GetBalance(qq);

            var sql = BalanceLog.SqlLog(botUin, groupId, groupName, qq, name, balanceUnfreeze, "解冻余额");
            var sql2 = SqlSetValues($"balance = balance + ({balanceUnfreeze}), BalanceFreeze = isnull(BalanceFreeze, 0) - ({balanceUnfreeze})", qq);
            int i = ExecTrans(sql, sql2);
            if (i == 0)
            {
                SyncCacheField(qq, "Balance", balanceValue + balanceUnfreeze);
                SyncCacheField(qq, "BalanceFreeze", freezeValue - balanceUnfreeze);
            }
            return i;
        }

        //增加余额sql
        public static (string, SqlParameter[]) SqlAddBalance(long userId, decimal balancePlus)
        {
            return Exists(userId)
                ? SqlPlus("Balance", balancePlus, userId)
                : SqlInsert([
                                new Cov("UserId", userId),
                                new Cov("Balance", balancePlus),
                            ]);
        }

        public static string GetBalanceList(long groupId, long qq)
        {
            string res = QueryRes($"select top 10 Id, balance from {FullName} where UserId in " +
                                  $"(select UserId from {CreditLog.FullName} where GroupId = {groupId}) order by balance desc",
                                  "【第{i}名】 [@:{0}] 余额：{1:N}\n");
            return res.Contains(qq.ToString())
                ? res
                : $"{res}{{余额排名}}";
        }

        public static string GetMyBalanceList(long groupId, long qq)
        {
            decimal balance = GetBalance(qq);
            string res = Query($"select count(*)+1 as res from {FullName} where balance > {balance} and UserId in " +
                               $"(select UserId from {CreditLog.FullName} where GroupId = {groupId})");
            return $"【第{res}名】 [@:{qq}] 余额：{balance:N}";
        }
    }
}
