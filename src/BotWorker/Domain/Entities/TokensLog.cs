using Microsoft.Data.SqlClient;
using BotWorker.Domain.Entities;
using BotWorker.Common.Extensions;
using BotWorker.Infrastructure.Persistence.ORM;

namespace BotWorker.Domain.Entities;
public class TokensLog : MetaData<TokensLog>
{
    public override string TableName => "Tokens";
    public override string KeyField => "Id";


    public static (string, SqlParameter[]) SqlLog(long botUin, long groupId, string groupName, long userId, string userName, long tokensAdd, string tokensInfo)
    {
        long tokensValue = UserInfo.GetTokens(userId);
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

    public static async Task AddLogAsync(long botUin, long groupId, string groupName, long userId, string userName, long tokensAdd, string tokensInfo, SqlTransaction? trans = null)
    {
        var (sql, paras) = SqlLog(botUin, groupId, groupName, userId, userName, tokensAdd, tokensInfo);
        await ExecAsync(sql, trans, paras);
    }

}