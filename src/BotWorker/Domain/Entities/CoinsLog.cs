using System.Data;

namespace BotWorker.Domain.Entities;
public class CoinsLog : MetaData<CoinsLog>
{
    public override string TableName => "Coins";
    public override string KeyField => "Id";

    public enum CoinsType { goldCoins, blackCoins, purpleCoins, gameCoins, groupCredit }
    public static List<string> conisFields = ["GoldCoins", "BlackCoins", "PurpleCoins", "GameCoins", "GroupCredit"];
    public static List<string> conisNames = ["金币", "黑金币", "紫币", "游戏币", "本群积分"];

    //异步增加日志 (支持事务)
    public static async Task<int> AddLogAsync(long botUin, long groupId, string groupName, long qq, string name, int coinsType, long coinsAdd, long coinsValue, string coinsInfo, IDbTransaction? trans = null)
    {
        var (sql, paras) = SqlHistory(botUin, groupId, groupName, qq, name, coinsType, coinsAdd, coinsValue, coinsInfo);
        string identitySql = IsPostgreSql ? " RETURNING Id" : ";SELECT SCOPE_IDENTITY();";
        return (await QueryScalarAsync<int>(sql + identitySql, trans, paras));
    }

    public static (string, IDataParameter[]) SqlHistory(long botUin, long groupId, string groupName, long qq, string name, int coinsType, long coinsAdd, long coinsValue, string coinsInfo)
    {
        return SqlInsert(new List<Cov> {
            new Cov("BotUin", botUin),
            new Cov("GroupId", groupId),
            new Cov("GroupName", groupName),
            new Cov("UserId", qq),
            new Cov("UserName", name),
            new Cov("CoinsType", coinsType),
            new Cov("CoinsAdd", coinsAdd),
            new Cov("CoinsValue", coinsValue + coinsAdd),
            new Cov("CoinsInfo", coinsInfo)
        });
    }
}