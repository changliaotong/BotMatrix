using System;
using System.Threading.Tasks;
using System.Data;
using BotWorker.Domain.Entities;
using BotWorker.Domain.Repositories;
using Dapper;

namespace BotWorker.Infrastructure.Persistence.Repositories
{
    public class GoodsTransRepository : BaseRepository<GoodsTrans>, IGoodsTransRepository
    {
        public GoodsTransRepository() : base("GoodsTrans")
        {
        }

        public async Task<int> AppendAsync(long groupId, long sellerQQ, long sellerOrderId, long buyerQQ, long buyerOrderId, int goodsId, long amount, decimal price, decimal buyerFee, decimal sellerFee, IDbTransaction? trans = null)
        {
            var transEntity = new GoodsTrans
            {
                GroupId = groupId,
                SellerQQ = sellerQQ,
                SellerOrderId = sellerOrderId,
                BuyerQQ = buyerQQ,
                BuyerOrderId = buyerOrderId,
                GoodsId = goodsId,
                Amount = amount,
                Price = price,
                BuyerFee = buyerFee,
                SellerFee = sellerFee,
                SellerBalance = amount * price - sellerFee,
                BuyerBalance = amount * price + buyerFee
            };

            await InsertAsync(transEntity, trans);
            return 1;
        }
    }
}
