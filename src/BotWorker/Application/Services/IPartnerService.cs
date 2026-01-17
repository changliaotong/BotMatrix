using System.Threading.Tasks;

namespace BotWorker.Application.Services
{
    public interface IPartnerService
    {
        Task<string> GetSettleResAsync(long botUin, long groupId, string groupName, long userId, string name);
        Task<string> BecomePartnerAsync(long userId);
        Task<string> GetCreditTodayAsync(long userId);
        Task<string> GetCreditListAsync(long userId);
    }
}
