using System.Data;

namespace BotWorker.Domain.Entities;
public class CreditLog : MetaData<CreditLog>
{
    public override string TableName => "Credit";
    public override string KeyField => "Id";

    //异步增加日志 (支持事务)
    public static async Task<int> AddLogAsync(long botUin, long groupId, string groupName, long qq, string name, long creditAdd, long creditValue, string creditInfo, IDbTransaction? trans = null)
    {
        var (sql, paras) = SqlHistory(botUin, groupId, groupName, qq, name, creditAdd, creditValue, creditInfo);
        string identitySql = IsPostgreSql ? " RETURNING Id" : ";SELECT SCOPE_IDENTITY();";
        return (await QueryScalarAsync<int>(sql + identitySql, trans, paras));
    }

    public static (string, IDataParameter[]) SqlHistory(long botUin, long groupId, string groupName, long qq, string name, long creditAdd, long creditValue, string creditInfo)
    {
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

    public static async Task<int> CreditCountAsync(long userId, string creditInfo, int second = 60)
    {
        return await QueryScalarAsync<int>($"select count(Id) from {FullName} where UserId = {{0}} and CreditInfo like {{1}} and abs({SqlDateDiff("second", SqlDateTime, "InsertDate")}) <= {{2}}", userId, $"%{creditInfo}%", second);
    }

}