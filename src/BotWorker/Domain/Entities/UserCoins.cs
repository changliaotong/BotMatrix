using BotWorker.Infrastructure.Persistence.ORM;

namespace BotWorker.Domain.Entities
{
    public partial class UserInfo : MetaDataGuid<UserInfo>
    {
        public static long GetCoins(long UserId)
            => GetCoinsAsync(UserId).GetAwaiter().GetResult();

        public static async Task<long> GetCoinsAsync(long UserId)
        {
            return await GetLongAsync("Coins", UserId);
        }
    }
}
