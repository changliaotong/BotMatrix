using BotWorker.Domain.Entities;
using BotWorker.Common.Extensions;
using BotWorker.Infrastructure.Persistence.ORM;

namespace BotWorker.Modules.Games.Gift
{

    class Gift : MetaData<Gift>
    {
        public override string TableName => "Gift";
        public override string KeyField => "Id";

        // 礼物ID
        static public long GetGiftId(string giftName)
            => GetGiftIdAsync(giftName).GetAwaiter().GetResult();

        static public async Task<long> GetGiftIdAsync(string giftName)
        {
            return (await GetWhereAsync("Id", $"GiftName={giftName.Quotes()}")).AsLong();
        }

        // 随机一个礼物
        public static long GetRandomGift(long groupId, long qq)
            => GetRandomGiftAsync(groupId, qq).GetAwaiter().GetResult();

        public static async Task<long> GetRandomGiftAsync(long groupId, long qq)
        {
            return (await QueryAsync($"select top 1 Id from {FullName} where GiftCredit < {await UserInfo.GetCreditAsync(groupId, qq)} order by newid()")).AsLong();
        }

        // 礼物列表
        public static string GetGiftList(long groupId, long qq)
            => GetGiftListAsync(groupId, qq).GetAwaiter().GetResult();

        public static async Task<string> GetGiftListAsync(long groupId, long qq)
        {
            long credit = await UserInfo.GetCreditAsync(groupId, qq);
            string res = await QueryResAsync($"select top 5 Id, GiftName, GiftCredit from {FullName} where IsValid = 1 and GiftCredit <= {credit} order by newid()", "{1}={2}分\n");
            if (res == "")
                res = await QueryResAsync($"select top 5 Id, GiftName, GiftCredit from {FullName}  where IsValid = 1 and GiftCredit < 10000 order by newid()", "{1}={2}分\n");
            return res;
        }
    }

    

}
