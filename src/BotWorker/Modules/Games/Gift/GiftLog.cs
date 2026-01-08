using Microsoft.Data.SqlClient;
using sz84.Bots.Entries;
using BotWorker.Common.Exts;
using BotWorker.Infrastructure.Persistence.ORM;

namespace BotWorker.Modules.Games.Gift
{
    class GiftLog : MetaData<GiftLog>
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
