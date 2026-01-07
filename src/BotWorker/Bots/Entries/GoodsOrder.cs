using BotWorker.Core.MetaDatas;

namespace BotWorker.Bots.Entries
{
    public class GoodsOrder : MetaData<GoodsOrder>
    {
        public override string TableName => "GoodsOrder";
        public override string KeyField => "OrderID";

        public static int Append(long groupId, long qq, int orderType, int goodsId, long amount, decimal price)
        {
            return Insert([
                        new Cov("GroupID", groupId),
                        new Cov("QQ", qq),
                        new Cov("OrderType", orderType),
                        new Cov("GoodsID", goodsId),
                        new Cov("Amount", amount),
                        new Cov("Price", price),
                    ]);
        }
    }
}
