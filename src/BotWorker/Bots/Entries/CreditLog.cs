using Microsoft.Data.SqlClient;
using sz84.Bots.Users;
using sz84.Core.MetaDatas;

namespace sz84.Bots.Entries;
public class CreditLog : MetaData<CreditLog>
{
    public override string TableName => "Credit";
    public override string KeyField => "Id";

    //积分变动记录
    public static (string, SqlParameter[]) SqlHistory(long botUin, long groupId, string groupName, long qq, string name, long creditAdd, string creditInfo)
    {
        long creditValue = UserInfo.GetCredit(groupId, qq);
        return SqlInsert([
                            new Cov("BotUin", botUin),
                                new Cov("GroupId", groupId),
                                new Cov("GroupName", groupName),
                                new Cov("UserId", qq),
                                new Cov("UserName", name),
                                new Cov("CreditAdd", creditAdd),
                                new Cov("CreditValue", creditValue + creditAdd),
                                new Cov("CreditInfo", creditInfo)
                    ]);
    }

    //积分变动次数（一段时间内） 例如：抽奖等
    public static int CreditCount(long userId, string creditInfo, int second = 60)
    {
        string sql = $"select count(Id) from {FullName} where UserId = {userId} and CreditInfo like '%{creditInfo}%' and abs(datediff(second, getdate(), InsertDate)) <= {second}";
        return Query(sql).AsInt();
    }

}