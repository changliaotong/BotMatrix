using Microsoft.Data.SqlClient;

namespace BotWorker.Domain.Entities;
public class CoinsLog : MetaData<CoinsLog>
{
    public override string TableName => "Coins";
    public override string KeyField => "Id";

    public enum CoinsType { goldCoins, blackCoins, purpleCoins, gameCoins, groupCredit }
    public static List<string> conisFields = ["GoldCoins", "BlackCoins", "PurpleCoins", "GameCoins", "GroupCredit"];
    public static List<string> conisNames = ["金币", "黑金币", "紫币", "游戏币", "本群积分"];

    public static (string, SqlParameter[]) SqlCoins(long botUin, long groupId, string groupName, long qq, string name, int coinsType, long coinsAdd, ref long coinsValue, string coinsInfo)
    {
        coinsValue = GroupMember.GetCoins(coinsType, groupId, qq);
        coinsValue += coinsAdd;
        return SqlInsert([
                            new Cov("BotUin", botUin),
                                new Cov("GroupId", groupId),
                                new Cov("GroupName", groupName),
                                new Cov("UserId", qq),
                                new Cov("UserName", name),
                                new Cov("CoinsType", coinsType),
                                new Cov("CoinsAdd", coinsAdd),
                                new Cov("CoinsValue", coinsValue),
                                new Cov("CoinsInfo", coinsInfo)
                        ]);
    }
}