using Microsoft.Data.SqlClient;
using BotWorker.Bots.Users;
using BotWorker.Core.MetaDatas;

namespace BotWorker.Bots.Entries;
public class TokensLog : MetaData<TokensLog>
{
    public override string TableName => "Tokens";
    public override string KeyField => "Id";


    public static (string, SqlParameter[]) SqlLog(long botUin, long groupId, string groupName, long userId, string userName, long tokensAdd, string tokensInfo)
    {
        long tokensValue = UserInfo.GetTokens(userId);
        return SqlInsert([
                            new Cov("BotUin", botUin),
                                new Cov("GroupId", groupId),
                                new Cov("GroupName", groupName),
                                new Cov("UserId", userId),
                                new Cov("UserName", userName),
                                new Cov("TokensAdd", tokensAdd),
                                new Cov("TokensValue", tokensValue + tokensAdd),
                                new Cov("TokensInfo", tokensInfo)
                        ]);
    }

}