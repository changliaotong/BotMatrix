using sz84.Bots.Users;
using sz84.Core.MetaDatas;

namespace sz84.Bots.Entries
{
    public class GoodsTrans : MetaData<GoodsTrans>
    {
        public override string TableName => "GoodsTrans";
        public override string KeyField => "TransID";

        public static int Append(long groupId, long sellerQQ, long sellerOrderId, long buyerQQ, long buyerOrderId, int goodsId, long amount, decimal price, decimal buyerFee, decimal sellerFee)
        {
            var (sql, parameters) = SqlInsert([
                                                new Cov("GroupID", groupId),
                                                new Cov("SellerQQ", sellerQQ),
                                                new Cov("SellerOrderID", sellerOrderId),
                                                new Cov("BuyerQQ", buyerQQ),
                                                new Cov("BuyerOrderID", buyerOrderId),
                                                new Cov("GoodsID", goodsId),
                                                new Cov("Amount", amount),
                                                new Cov("Price", price),
                                                new Cov("SellerFee", sellerFee),
                                                new Cov("BuyerFee", buyerFee),
                                                new Cov("SellerBalance", amount * price - sellerFee),
                                                new Cov("BuyerBalance", amount * price + buyerFee),
                                            ]);
            return Exec(sql, parameters);
        }

        public static string TransIt(long botUin, long orderId, long groupId, string groupName, long qq, string name)
        {
            string res = "";
            //新买单或新卖单加入时自动撮合交易
            int orderType = GoodsOrder.GetInt("OrderType", orderId);
            long sellerQQ = GoodsOrder.GetLong("QQ", orderId);
            int goodsId = GoodsOrder.GetInt("GoodsId", orderId);
            long amount = GoodsOrder.GetLong("Amount", orderId);
            decimal price = GoodsOrder.Get<decimal>("price", orderId);
            bool isFinish = false;
            while (!isFinish)
            {
                bool isSell = orderType == 0;
                string id2 = Query($"select top 1 orderId from {GoodsOrder.FullName} where GoodsID = {goodsId} and orderType != {orderType} and price {(isSell ? ">=" : "<=")} {price} order by price {(isSell ? "desc" : "")}");
                if (!id2.IsNull())
                {
                    long orderId2 = id2.AsLong();
                    long amount2 = GoodsOrder.GetLong("amount", orderId2);
                    decimal price2 = GoodsOrder.Get<decimal>("price", orderId2);
                    decimal fee = amount2 * price2 * 0.01m;
                    if (goodsId == 1)
                    {
                        //积分
                        var sql = CreditLog.SqlHistory(botUin, groupId, groupName, qq, name, amount, $"买入积分");
                        var sql2 = CreditLog.SqlHistory(botUin, groupId, groupName, sellerQQ, name, amount, $"卖出积分");
                        var sql3 = UserInfo.SqlAddCredit(botUin, groupId, qq, amount);
                        var sql4 = UserInfo.SqlAddCredit(botUin, groupId, sellerQQ, -amount2);
                        //余额
                        var sql5 = UserInfo.SqlAddBalance(qq, amount * price);
                        var sql6 = UserInfo.SqlAddBalance(sellerQQ, amount * price);
                        var sql7 = BalanceLog.SqlLog(botUin, groupId, groupName, qq, name, amount * price, $"购买积分");
                        var sql8 = BalanceLog.SqlLog(botUin, groupId, groupName, sellerQQ, name, amount * price, $"卖出积分");
                        var sql9 = BalanceLog.SqlLog(botUin, groupId, groupName, sellerQQ, name, fee, $"手续费");
                        //goods trans
                        var sql10 = SqlInsert([
                                                new Cov("GroupID", groupId),
                                                new Cov("SellerQQ", sellerQQ),
                                                new Cov("SellerOrderID", orderId),
                                                new Cov("BuyerQQ", qq),
                                                new Cov("BuyerOrderID", orderId2),
                                                new Cov("GoodsID", goodsId),
                                                new Cov("Amount", amount),
                                                new Cov("Price", price),
                                                new Cov("SellerFee", fee),
                                                new Cov("BuyerFee", 0),
                                                new Cov("SellerBalance", amount * price - fee),
                                                new Cov("BuyerBalance", amount * price + 0),
                                            ]);

                        int i = ExecTrans(sql, sql2, sql2, sql3, sql4, sql5, sql6, sql7, sql8, sql9, sql10);
                        return i == -1
                            ? RetryMsg 
                            : "✅ 交易成功";
                    }
                    else
                        return $"目前仅支持积分交易，敬请期待";
                    //先处理 goodsId = 1 积分的情况？  2 = 算力
                }
                else
                    isFinish = true;
            }
            return res;
        }
    }
}
