using System.Data;
using BotWorker.Domain.Entities;
using BotWorker.Common.Extensions;
using BotWorker.Infrastructure.Persistence.ORM;

namespace BotWorker.Domain.Entities;
public class CreditLog : MetaData<CreditLog>
{
    public override string TableName => "Credit";
    public override string KeyField => "Id";

    //异步增加日志 (支持事务)
    public static async Task<int> AddLogAsync(long botUin, long groupId, string groupName, long qq, string name, long creditAdd, string creditInfo, IDbTransaction? trans = null)
    {
        var (sql, paras) = await SqlHistoryAsync(botUin, groupId, groupName, qq, name, creditAdd, creditInfo);
        return (await QueryScalarAsync<int>(sql + ";SELECT SCOPE_IDENTITY();", trans, paras));
    }

    //积分变动记录
    public static (string, IDataParameter[]) SqlHistory(long botUin, long groupId, string groupName, long qq, string name, long creditAdd, string creditInfo)
        => SqlHistoryAsync(botUin, groupId, groupName, qq, name, creditAdd, creditInfo).GetAwaiter().GetResult();

    public static async Task<(string, IDataParameter[])> SqlHistoryAsync(long botUin, long groupId, string groupName, long qq, string name, long creditAdd, string creditInfo)
    {
        long creditValue = await UserInfo.GetCreditAsync(groupId, qq);
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
        => CreditCountAsync(userId, creditInfo, second).GetAwaiter().GetResult();

    public static async Task<int> CreditCountAsync(long userId, string creditInfo, int second = 60)
    {
        string sql = $"select count(Id) from {FullName} where UserId = {userId} and CreditInfo like '%{creditInfo}%' and abs(datediff(second, getdate(), InsertDate)) <= {second}";
        return (await QueryAsync(sql)).AsInt();
    }

}