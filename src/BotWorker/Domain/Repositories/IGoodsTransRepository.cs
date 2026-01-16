using System;
using System.Threading.Tasks;
using System.Data;
using BotWorker.Domain.Entities;

namespace BotWorker.Domain.Repositories
{
    public interface IGoodsTransRepository
    {
        Task<int> AppendAsync(long groupId, long sellerQQ, long sellerOrderId, long buyerQQ, long buyerOrderId, int goodsId, long amount, decimal price, decimal buyerFee, decimal sellerFee, IDbTransaction? trans = null);
    }
}
