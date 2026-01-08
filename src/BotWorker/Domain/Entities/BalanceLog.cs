using System.Data;
using BotWorker.Domain.Entities;
using BotWorker.Common.Extensions;
using BotWorker.Infrastructure.Persistence.ORM;

namespace BotWorker.Domain.Entities;

public class BalanceLog : MetaData<BalanceLog>
{
    public override string TableName => "Balance";
    public override string KeyField => "Id";

    public static (string, IDataParameter[]) SqlLog(long botUin, long groupId, string groupName, long userId, string name, decimal balanceAdd, string balanceInfo)
    {
        decimal balance = UserInfo.GetBalance(userId);
        return SqlLog(botUin, groupId, groupName, userId, name, balanceAdd, balanceInfo, balance + balanceAdd);
    }

    public static (string, IDataParameter[]) SqlLog(long botUin, long groupId, string groupName, long userId, string name, decimal balanceAdd, string balanceInfo, decimal balance_last)
    {
        return SqlInsert([
            new Cov("BotUin", botUin),
            new Cov("GroupId", groupId),
            new Cov("GroupName", groupName),
            new Cov("UserId", userId),
            new Cov("UserName", name),
            new Cov("BalanceAdd", balanceAdd),
            new Cov("BalanceValue", balance_last),
            new Cov("BalanceInfo", balanceInfo),
        ]);
    }
}