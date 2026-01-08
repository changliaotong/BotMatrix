using BotWorker.Infrastructure.Persistence.ORM;

namespace sz84.Bots.Users
{
    public partial class UserInfo : MetaDataGuid<UserInfo>
    {
        public static long GetCoins(long UserId)
        {
            return GetLong("Coins", UserId);
        }
    }
}
