using BotWorker.Domain.Entities;
using BotWorker.Common.Extensions;
using BotWorker.Infrastructure.Persistence.ORM;

namespace BotWorker.Domain.Entities
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
            return TransItAsync(botUin, orderId, groupId, groupName, qq, name).GetAwaiter().GetResult();
        }

        public static async Task<string> TransItAsync(long botUin, long orderId, long groupId, string groupName, long qq, string name)
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
                string id2 = await QueryScalarAsync<string>($"select {SqlTop(1)} orderId from {GoodsOrder.FullName} where GoodsID = {goodsId} and orderType != {orderType} and price {(isSell ? ">=" : "<=")} {price} order by price {(isSell ? "desc" : "")}{SqlLimit(1)}") ?? "";
                if (!id2.IsNull())
                {
                    long orderId2 = id2.AsLong();
                    long amount2 = GoodsOrder.GetLong("amount", orderId2);
                    decimal price2 = GoodsOrder.Get<decimal>("price", orderId2);
                    decimal fee = amount2 * price2 * 0.01m;
                    if (goodsId == 1)
                    {
                        //积分和余额交易
                        using var trans = await BeginTransactionAsync();
                        try
                        {
                            // 1. 积分操作
                            var addResBuyer = await UserInfo.AddCreditAsync(botUin, groupId, groupName, qq, name, amount, "买入积分", trans);
                            if (addResBuyer.Result == -1) throw new Exception("买入积分失败");

                            var addResSeller = await UserInfo.AddCreditAsync(botUin, groupId, groupName, sellerQQ, name, -amount2, "卖出积分", trans);
                            if (addResSeller.Result == -1) throw new Exception("卖出积分失败");

                            // 2. 余额操作
                            var balResBuyer = await UserInfo.AddBalanceAsync(botUin, groupId, groupName, qq, name, -amount * price, "购买积分", trans);
                            if (balResBuyer.Result == -1) throw new Exception("购买积分扣除余额失败");

                            var balResSeller = await UserInfo.AddBalanceAsync(botUin, groupId, groupName, sellerQQ, name, amount * price - fee, "卖出积分", trans);
                            if (balResSeller.Result == -1) throw new Exception("卖出积分增加余额失败");

                            // 3. 手续费日志
                            // 注意：UserInfo.AddBalanceAsync 已经记录了流水日志，但如果需要额外记录手续费，可以在这里手动添加，或者合并到 AddBalanceAsync 的描述中
                            // 这里我们保留手动记录手续费日志，或者可以将其作为 AddBalanceAsync 的一部分逻辑处理。
                            // 由于 balResSeller.BalanceValue 是增加后的余额，我们直接用它作为日志的当前余额。
                            var (sqlFee, parasFee) = BalanceLog.SqlLog(botUin, groupId, groupName, sellerQQ, name, -fee, "交易手续费", balResSeller.BalanceValue);
                            await ExecAsync(sqlFee, trans, parasFee);

                            // 4. 交易记录
                            var (sqlTrans, parasTrans) = SqlInsert([
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
                            await ExecAsync(sqlTrans, trans, parasTrans);

                            await trans.CommitAsync();

                            // 同步缓存
                            UserInfo.SyncCacheField(qq, groupId, "Credit", addResBuyer.CreditValue);
                            UserInfo.SyncCacheField(sellerQQ, groupId, "Credit", addResSeller.CreditValue);
                            UserInfo.SyncCacheField(qq, "Balance", balResBuyer.BalanceValue);
                            UserInfo.SyncCacheField(sellerQQ, "Balance", balResSeller.BalanceValue);

                            return "✅ 交易成功";
                        }
                        catch (Exception ex)
                        {
                            await trans.RollbackAsync();
                            Console.WriteLine($"[TransIt Error] {ex.Message}");
                            return RetryMsg;
                        }
                    }
                    else
                        return $"目前仅支持积分交易，敬请期待";
                }
                else
                    isFinish = true;
            }
            return res;
        }
    }
}
