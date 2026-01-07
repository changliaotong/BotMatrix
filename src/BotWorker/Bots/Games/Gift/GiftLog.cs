using Microsoft.Data.SqlClient;
using BotWorker.Bots.Entries;
using BotWorker.Core.MetaDatas;

namespace BotWorker.Bots.Games.Gift
{
    public class GiftLog : MetaData<GiftLog>
    {
        public override string TableName => "GiftLog";
        public override string KeyField => "Id";

        public static (string, SqlParameter[]) SqlAppend(long botUin, long groupId, string groupName, long qq, string name, long robotOwner, string ownerName, long qqGift, string giftClientName,
            long giftId, string giftName, int giftCount, long giftCredit)
        {
            return SqlInsert([
                                new Cov("BotUin", botUin),
                                new Cov("GroupId", groupId),
                                new Cov("GroupName", groupName),
                                new Cov("UserId", qq),
                                new Cov("UserName", name),
                                new Cov("RobotOwner", robotOwner),
                                new Cov("OwnerName", ownerName),
                                new Cov("GiftUserId", qqGift),
                                new Cov("GiftUserName", giftClientName),
                                new Cov("GiftId", giftId),
                                new Cov("GiftName", giftName),
                                new Cov("GiftCount", giftCount),
                                new Cov("GiftCredit", giftCredit),
                            ]);
        }
    }
}
