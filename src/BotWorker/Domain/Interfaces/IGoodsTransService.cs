using System.Threading.Tasks;

namespace BotWorker.Domain.Interfaces
{
    public interface IGoodsTransService
    {
        Task<int> AppendAsync(long groupId, long sellerQQ, long sellerOrderId, long buyerQQ, long buyerOrderId, int goodsId, long amount, decimal price, decimal buyerFee, decimal sellerFee);
        Task<string> TransItAsync(long botUin, long orderId, long groupId, string groupName, long qq, string name);
    }
}
