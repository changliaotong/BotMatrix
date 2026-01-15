
using System.Threading.Tasks;

namespace BotWorker.Domain.Entities
{
    public partial class UserInfo
    {
        public static async Task<long> GetCoinsAsync(long userId)
        {
            return await Repository.GetCoinsAsync(userId);
        }
    }
}
