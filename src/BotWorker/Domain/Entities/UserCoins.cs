using BotWorker.Infrastructure.Persistence.ORM;

namespace BotWorker.Domain.Entities
{
    public partial class UserInfo : MetaDataGuid<UserInfo>
    {
        public static long GetCoins(long UserId)
        {
            return GetLong("Coins", UserId);
        }
    }
}
