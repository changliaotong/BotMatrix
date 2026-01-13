using System.Data;

namespace BotWorker.Domain.Entities;

public class BalanceLog : MetaData<BalanceLog>
{
    public override string TableName => "Balance";
    public override string KeyField => "Id";

    public static (string, IDataParameter[]) SqlLog(long botUin, long groupId, string groupName, long userId, string name, decimal balanceAdd, string balanceInfo, decimal balanceValue)
    {
        return SqlInsert(new
        {
            BotUin = botUin,
            GroupId = groupId,
            GroupName = groupName,
            UserId = userId,
            UserName = name,
            BalanceAdd = balanceAdd,
            BalanceValue = balanceValue,
            BalanceInfo = balanceInfo
        });
    }
}