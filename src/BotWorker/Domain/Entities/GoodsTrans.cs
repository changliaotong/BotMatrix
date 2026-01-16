using System;
using System.Threading.Tasks;
using System.Data;
using BotWorker.Domain.Repositories;
using BotWorker.Infrastructure.Persistence;
using Dapper.Contrib.Extensions;
using Microsoft.Extensions.DependencyInjection;

namespace BotWorker.Domain.Entities
{
    [Table("GoodsTrans")]
    public class GoodsTrans
    {
        private static IGoodsTransRepository Repository => 
            BotMessage.ServiceProvider?.GetRequiredService<IGoodsTransRepository>() 
            ?? throw new InvalidOperationException("IGoodsTransRepository not registered");

        private static IGoodsOrderRepository OrderRepository => 
            BotMessage.ServiceProvider?.GetRequiredService<IGoodsOrderRepository>() 
            ?? throw new InvalidOperationException("IGoodsOrderRepository not registered");

        private static IUserRepository UserRepository => 
            BotMessage.ServiceProvider?.GetRequiredService<IUserRepository>() 
            ?? throw new InvalidOperationException("IUserRepository not registered");

        private static IBalanceLogRepository BalanceLogRepository => 
            BotMessage.ServiceProvider?.GetRequiredService<IBalanceLogRepository>() 
            ?? throw new InvalidOperationException("IBalanceLogRepository not registered");

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

        public static int Append(long groupId, long sellerQQ, long sellerOrderId, long buyerQQ, long buyerOrderId, int goodsId, long amount, decimal price, decimal buyerFee, decimal sellerFee)
        {
            return Repository.AppendAsync(groupId, sellerQQ, sellerOrderId, buyerQQ, buyerOrderId, goodsId, amount, price, buyerFee, sellerFee).GetAwaiter().GetResult();
        }

        public static string TransIt(long botUin, long orderId, long groupId, string groupName, long qq, string name)
        {
            return TransItAsync(botUin, orderId, groupId, groupName, qq, name).GetAwaiter().GetResult();
        }

        public static async Task<string> TransItAsync(long botUin, long orderId, long groupId, string groupName, long qq, string name)
        {
            string res = "";
            //新买单或新卖单加入时自动撮合交易
            int orderType = await OrderRepository.GetValueAsync<int>("OrderType", orderId);
            long sellerQQ = await OrderRepository.GetValueAsync<long>("QQ", orderId);
            int goodsId = await OrderRepository.GetValueAsync<int>("GoodsId", orderId);
            long amount = await OrderRepository.GetValueAsync<long>("Amount", orderId);
            decimal price = await OrderRepository.GetValueAsync<decimal>("price", orderId);
            bool isFinish = false;
            while (!isFinish)
            {
                string? id2 = await OrderRepository.GetMatchingOrderIdAsync(goodsId, orderType, price);
                if (!string.IsNullOrEmpty(id2))
                {
                    long orderId2 = long.Parse(id2);
                    long amount2 = await OrderRepository.GetValueAsync<long>("amount", orderId2);
                    decimal price2 = await OrderRepository.GetValueAsync<decimal>("price", orderId2);
                    decimal fee = amount2 * price2 * 0.01m;
                    if (goodsId == 1)
                    {
                        //积分和余额交易
                        using var transWrapper = await TransactionWrapper.BeginTransactionAsync();
                        var trans = transWrapper.Transaction;
                        try
                        {
                            // 1. 积分操作
                            var addResBuyer = await UserRepository.AddCreditAsync(botUin, groupId, groupName, qq, name, amount, "买入积分", trans);
                            if (addResBuyer.Result == -1) throw new Exception("买入积分失败");

                            var addResSeller = await UserRepository.AddCreditAsync(botUin, groupId, groupName, sellerQQ, name, -amount2, "卖出积分", trans);
                            if (addResSeller.Result == -1) throw new Exception("卖出积分失败");

                            // 2. 余额操作
                            var balResBuyer = await UserRepository.AddBalanceAsync(botUin, groupId, groupName, qq, name, -amount * price, "购买积分", trans);
                            if (balResBuyer.Result == -1) throw new Exception("购买积分扣除余额失败");

                            var balResSeller = await UserRepository.AddBalanceAsync(botUin, groupId, groupName, sellerQQ, name, amount * price - fee, "卖出积分", trans);
                            if (balResSeller.Result == -1) throw new Exception("卖出积分增加余额失败");

                            // 3. 手续费日志
                            await BalanceLogRepository.AddLogAsync(botUin, groupId, groupName, sellerQQ, name, -fee, balResSeller.BalanceValue, "交易手续费", trans);

                            // 4. 交易记录
                            await Repository.AppendAsync(groupId, sellerQQ, orderId, qq, orderId2, goodsId, amount, price, 0, fee, trans);

                            await transWrapper.CommitAsync();

                            // 同步缓存
                            UserInfo.SyncCacheField(qq, groupId, "Credit", addResBuyer.CreditValue);
                            UserInfo.SyncCacheField(sellerQQ, groupId, "Credit", addResSeller.CreditValue);
                            UserInfo.SyncCacheField(qq, groupId, "Balance", balResBuyer.BalanceValue);
                            UserInfo.SyncCacheField(sellerQQ, groupId, "Balance", balResSeller.BalanceValue);
                            
                            isFinish = true;
                        }
                        catch (Exception ex)
                        {
                            await transWrapper.RollbackAsync();
                            res = "交易异常: " + ex.Message;
                            isFinish = true;
                        }
                    }
                    else
                    {
                        isFinish = true;
                    }
                }
                else
                {
                    isFinish = true;
                }
            }
            return res;
        }
    }
}
