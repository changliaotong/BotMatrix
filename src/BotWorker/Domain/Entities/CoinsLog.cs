using System.Data;

namespace BotWorker.Domain.Entities;
public class CoinsLog : MetaData<CoinsLog>
{
    public override string TableName => "Coins";
    public override string KeyField => "Id";

    public enum CoinsType { goldCoins, blackCoins, purpleCoins, gameCoins, groupCredit }
    public static List<string> conisFields = ["GoldCoins", "BlackCoins", "PurpleCoins", "GameCoins", "GroupCredit"];
    public static List<string> conisNames = ["金币", "黑金币", "紫币", "游戏币", "本群积分"];

    public static async Task<(string sql, IDataParameter[] parameters, long coinsValue)> SqlCoinsAsync(long botUin, long groupId, string groupName, long qq, string name, int coinsType, long coinsAdd, string coinsInfo, IDbTransaction? trans = null)
    {
        long coinsValue = await GroupMember.GetCoinsAsync(coinsType, groupId, qq, trans) + coinsAdd;
        var (sql, paras) = SqlInsert([
            new("BotUin", botUin),
            new("GroupId", groupId),
            new("GroupName", groupName),
            new("UserId", qq),
            new("UserName", name),
            new("CoinsType", coinsType),
            new("CoinsAdd", coinsAdd),
            new("CoinsValue", coinsValue),
            new("CoinsInfo", coinsInfo)
        ]);
        return (sql, paras, coinsValue);
    }

    public static (string sql, IDataParameter[] parameters, long coinsValue) SqlCoins(long botUin, long groupId, string groupName, long qq, string name, int coinsType, long coinsAdd, string coinsInfo, IDbTransaction? trans = null)
    {
        return SqlCoinsAsync(botUin, groupId, groupName, qq, name, coinsType, coinsAdd, coinsInfo, trans).GetAwaiter().GetResult();
    }

    public static async Task AddLogAsync(long botUin, long groupId, string groupName, long qq, string name, int coinsType, long coinsAdd, string coinsInfo, IDbTransaction? trans = null)
    {
        long coinsValue = await GroupMember.GetCoinsAsync(coinsType, groupId, qq);
        coinsValue += coinsAdd;
        var (sql, paras) = SqlInsert(new List<Cov> {
                            new Cov("BotUin", botUin),
                                new Cov("GroupId", groupId),
                                new Cov("GroupName", groupName),
                                new Cov("UserId", qq),
                                new Cov("UserName", name),
                                new Cov("CoinsType", coinsType),
                                new Cov("CoinsAdd", coinsAdd),
                                new Cov("CoinsValue", coinsValue),
                                new Cov("CoinsInfo", coinsInfo)
                        });
        await ExecAsync(sql, trans, paras);
    }
}