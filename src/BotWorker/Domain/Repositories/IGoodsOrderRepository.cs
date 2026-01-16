using System;
using System.Threading.Tasks;
using BotWorker.Domain.Entities;

namespace BotWorker.Domain.Repositories
{
    public interface IGoodsOrderRepository
    {
        Task<long> AppendAsync(long groupId, long qq, int orderType, int goodsId, long amount, decimal price);
        Task<T> GetValueAsync<T>(string field, long orderId);
        Task<GoodsOrder?> GetByIdAsync(long orderId);
        Task<string?> GetMatchingOrderIdAsync(int goodsId, int orderType, decimal price);
    }
}
