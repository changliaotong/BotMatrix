using System.Data;

namespace BotWorker.Domain.Entities;
public class TokensLog : MetaData<TokensLog>
{
    public override string TableName => "Tokens";
    public override string KeyField => "Id";


    public static (string, IDataParameter[]) SqlLog(long botUin, long groupId, string groupName, long userId, string userName, long tokensAdd, long tokensValue, string tokensInfo)
    {
        return SqlInsert(new List<Cov> {
                            new Cov("BotUin", botUin),
                                new Cov("GroupId", groupId),
                                new Cov("GroupName", groupName),
                                new Cov("UserId", userId),
                                new Cov("UserName", userName),
                                new Cov("TokensAdd", tokensAdd),
                                new Cov("TokensValue", tokensValue + tokensAdd),
                                new Cov("TokensInfo", tokensInfo)
                        });
    }

    //异步增加日志 (支持事务)
    public static async Task<int> AddLogAsync(long botUin, long groupId, string groupName, long userId, string userName, long tokensAdd, long tokensValue, string tokensInfo, IDbTransaction? trans = null)
    {
        var (sql, paras) = SqlLog(botUin, groupId, groupName, userId, userName, tokensAdd, tokensValue, tokensInfo);
        string identitySql = IsPostgreSql ? " RETURNING Id" : ";SELECT SCOPE_IDENTITY();";
        return (await QueryScalarAsync<int>(sql + identitySql, trans, paras));
    }

}