using System.Threading.Tasks;

namespace BotWorker.Domain.Interfaces
{
    public interface IGroupPropsService
    {
        Task<long> GetIdAsync(long groupId, long qq, long propId);
        Task<bool> HavePropAsync(long groupId, long userId, long propId);
        Task<int> UsePropAsync(long groupId, long userId, long propId, long qqProp);
        Task<string> GetMyPropListAsync(long groupId, long userId);
        Task<bool> IsClosedAsync(long groupId);
        Task<string> GetBuyResAsync(long botUin, long groupId, string groupName, long qq, string name, string cmdPara);
    }
}
