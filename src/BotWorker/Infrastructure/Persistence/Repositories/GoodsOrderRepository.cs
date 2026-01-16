using System;
using System.Threading.Tasks;
using BotWorker.Domain.Entities;
using BotWorker.Domain.Repositories;
using Dapper;

namespace BotWorker.Infrastructure.Persistence.Repositories
{
    public class GoodsOrderRepository : BaseRepository<GoodsOrder>, IGoodsOrderRepository
    {
        public GoodsOrderRepository() : base("GoodsOrder")
        {
        }

        public async Task<long> AppendAsync(long groupId, long qq, int orderType, int goodsId, long amount, decimal price)
        {
            var order = new GoodsOrder
            {
                GroupId = groupId,
                QQ = qq,
                OrderType = orderType,
                GoodsId = goodsId,
                Amount = amount,
                Price = price
            };
            return await InsertAsync(order);
        }

        public async Task<T> GetValueAsync<T>(string field, long orderId)
        {
            return await GetValueAsync<T>(field, "WHERE OrderID = @orderId", new { orderId });
        }

        public async Task<GoodsOrder?> GetByIdAsync(long orderId)
        {
            string sql = $"SELECT * FROM {_tableName} WHERE OrderID = @orderId";
            using var conn = CreateConnection();
            return await conn.QueryFirstOrDefaultAsync<GoodsOrder>(sql, new { orderId });
        }

        public async Task<string?> GetMatchingOrderIdAsync(int goodsId, int orderType, decimal price)
        {
            bool isSell = orderType == 0;
            string orderSql = isSell ? "DESC" : "";
            string priceOp = isSell ? ">=" : "<=";
            
            string sql = $@"
                SELECT OrderID 
                FROM {_tableName} 
                WHERE GoodsID = @goodsId 
                AND OrderType != @orderType 
                AND Price {priceOp} @price 
                ORDER BY Price {orderSql}
                LIMIT 1";

            using var conn = CreateConnection();
            return await conn.QueryFirstOrDefaultAsync<string>(sql, new { goodsId, orderType, price });
        }
    }
}
