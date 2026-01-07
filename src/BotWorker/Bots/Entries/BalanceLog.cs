using Microsoft.Data.SqlClient;
using BotWorker.Bots.Users;
using BotWorker.Core.MetaDatas;

namespace BotWorker.Bots.Entries;

public class BalanceLog : MetaData<BalanceLog>
{
    public override string TableName => "Balance";
    public override string KeyField => "Id";

    public static (string, SqlParameter[]) SqlLog(long botUin, long groupId, string groupName, long userId, string name, decimal balanceAdd, string balanceInfo)
    {
        decimal balance = UserInfo.GetBalance(userId);
        return SqlInsert([
            new Cov("BotUin", botUin),
                new Cov("GroupId", groupId),
                new Cov("GroupName", groupName),
                new Cov("UserId", userId),
                new Cov("UserName", name),
                new Cov("BalanceAdd", balanceAdd),
                new Cov("BalanceValue", balance + balanceAdd),
                new Cov("BalanceInfo", balanceInfo),
                ]);
    }
}