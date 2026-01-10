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
        string identitySql = IsPostgreSql ? " RETURNING Id" : ";SELECT SCOPE_IDENTITY();";
        return (await QueryScalarAsync<int>(sql + identitySql, trans, paras));
    }

    //积分变动记录
    public static (string, IDataParameter[]) SqlHistory(long botUin, long groupId, string groupName, long qq, string name, long creditAdd, string creditInfo)
        => SqlHistoryAsync(botUin, groupId, groupName, qq, name, creditAdd, creditInfo).GetAwaiter().GetResult();

    public static async Task<(string, IDataParameter[])> SqlHistoryAsync(long botUin, long groupId, string groupName, long qq, string name, long creditAdd, string creditInfo)
    {
        long creditValue = await UserInfo.GetCreditAsync(groupId, qq);
        return SqlInsert(new
        {
            BotUin = botUin,
            GroupId = groupId,
            GroupName = groupName,
            UserId = qq,
            UserName = name,
            CreditAdd = creditAdd,
            CreditValue = creditValue + creditAdd,
            CreditInfo = creditInfo
        });
    }

    //积分变动次数（一段时间内） 例如：抽奖等
    public static int CreditCount(long userId, string creditInfo, int second = 60)
        => CreditCountAsync(userId, creditInfo, second).GetAwaiter().GetResult();

    public static async Task<int> CreditCountAsync(long userId, string creditInfo, int second = 60)
    {
        return await QueryScalarAsync<int>($"select count(Id) from {FullName} where UserId = {{0}} and CreditInfo like {{1}} and abs({SqlDateDiff("second", SqlDateTime, "InsertDate")}) <= {{2}}", userId, $"%{creditInfo}%", second);
    }

}