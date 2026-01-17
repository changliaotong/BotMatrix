using System;
using System.Data;
using Dapper.Contrib.Extensions;

namespace BotWorker.Domain.Entities
{
    [Table("goods_trans")]
    public class GoodsTrans
    {
        [Key]
        public long TransId { get; set; }
        public long GroupId { get; set; }
        public long SellerQQ { get; set; }
        public long SellerOrderId { get; set; }
        public long BuyerQQ { get; set; }
        public long BuyerOrderId { get; set; }
        public int GoodsId { get; set; }
        public long Amount { get; set; }
        public decimal Price { get; set; }
        public decimal BuyerFee { get; set; }
        public decimal SellerFee { get; set; }
        public decimal SellerBalance { get; set; }
        public decimal BuyerBalance { get; set; }
    }
}
