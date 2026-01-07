using BotWorker.Core.MetaDatas;

namespace BotWorker.Bots.Users
{
    public partial class UserInfo : MetaDataGuid<UserInfo>
    {
        public static long GetCoins(long UserId)
        {
            return GetLong("Coins", UserId);
        }
    }
}
