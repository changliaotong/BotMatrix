namespace sz84.Bots.Entries
{
    public class Order
    {
        public enum PaymentMode { PreloadBalance = 1, DirectConsume = 2 }
        public enum OrderType { Power = 1, Points = 2, Service = 3 }

        public long OrderId { get; set; }
        public long UserId { get; set; }
        public OrderType Type { get; set; }    // 算力、积分、服务等
        public PaymentMode PayMode { get; set; }
        public long Amount { get; set; }       // 订单金额（单位分）
        public long Quantity { get; set; }     // 购买数量
        public string Status { get; set; } = string.Empty;   // Pending, Paid, Failed
        public DateTime CreateTime { get; set; }
        public DateTime? PayTime { get; set; }
    }
}
