using System;
using System.Threading.Tasks;
using BotWorker.Domain.Repositories;
using Dapper.Contrib.Extensions;
using Microsoft.Extensions.DependencyInjection;

namespace BotWorker.Domain.Entities
{
    [Table("GoodsOrder")]
    public class GoodsOrder
    {
        private static IGoodsOrderRepository Repository => 
            BotMessage.ServiceProvider?.GetRequiredService<IGoodsOrderRepository>() 
            ?? throw new InvalidOperationException("IGoodsOrderRepository not registered");

        [Key]
        public long OrderId { get; set; }
        public long GroupId { get; set; }
        public long QQ { get; set; }
        public int OrderType { get; set; }
        public int GoodsId { get; set; }
        public long Amount { get; set; }
        public decimal Price { get; set; }

        public static int Append(long groupId, long qq, int orderType, int goodsId, long amount, decimal price)
        {
            return (int)Repository.AppendAsync(groupId, qq, orderType, goodsId, amount, price).GetAwaiter().GetResult();
        }

        public static int GetInt(string field, long orderId)
        {
            return Repository.GetValueAsync<int>(field, orderId).GetAwaiter().GetResult();
        }

        public static long GetLong(string field, long orderId)
        {
            return Repository.GetValueAsync<long>(field, orderId).GetAwaiter().GetResult();
        }

        public static T Get<T>(string field, long orderId)
        {
            return Repository.GetValueAsync<T>(field, orderId).GetAwaiter().GetResult();
        }
    }
}
