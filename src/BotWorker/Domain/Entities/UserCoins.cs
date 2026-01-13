
namespace BotWorker.Domain.Entities
{
    public partial class UserInfo : MetaDataGuid<UserInfo>
    {

        public static async Task<long> GetCoinsAsync(long UserId)
        {
            return await GetLongAsync("Coins", UserId);
        }
    }
}
