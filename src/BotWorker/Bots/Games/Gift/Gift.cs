using sz84.Bots.Entries;
using sz84.Bots.Users;
using BotWorker.Common.Exts;
using sz84.Core.MetaDatas;

namespace sz84.Bots.Games.Gift
{

    public class Gift : MetaData<Gift>
    {
        public override string TableName => "Gift";
        public override string KeyField => "Id";

        // 礼物ID
        static public long GetGiftId(string giftName)
        {
            return GetWhere("Id", $"GiftName={giftName.Quotes()}").AsLong();
        }

        // 随机一个礼物
        public static long GetRandomGift(long groupId, long qq)
        {
            return Query($"select top 1 Id from {FullName} where GiftCredit < {UserInfo.GetCredit(groupId, qq)} order by newid()").AsLong();
        }

        // 礼物列表
        public static string GetGiftList(long groupId, long qq)
        {
            string res = QueryRes($"select top 5 Id, GiftName, GiftCredit from {FullName} where IsValid = 1 and GiftCredit <= {UserInfo.GetCredit(groupId, qq)} order by newid()", "{1}={2}分\n");
            if (res == "")
                res = QueryRes($"select top 5 Id, GiftName, GiftCredit from {FullName}  where IsValid = 1 and GiftCredit < 10000 order by newid()", "{1}={2}分\n");
            return res;
        }
    }

    

}
