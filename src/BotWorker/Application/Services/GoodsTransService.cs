using System;
using System.Threading.Tasks;
using BotWorker.Domain.Entities;
using BotWorker.Domain.Interfaces;
using BotWorker.Domain.Repositories;
using BotWorker.Infrastructure.Persistence;
using Microsoft.Extensions.Logging;

namespace BotWorker.Application.Services
{
    public class GoodsTransService : IGoodsTransService
    {
        private readonly IGoodsTransRepository _goodsTransRepo;
        private readonly IGoodsOrderRepository _orderRepo;
        private readonly IUserRepository _userRepo;
        private readonly IBalanceLogRepository _balanceLogRepo;
        private readonly ILogger<GoodsTransService> _logger;

        public GoodsTransService(
            IGoodsTransRepository goodsTransRepo,
            IGoodsOrderRepository orderRepo,
            IUserRepository userRepo,
            IBalanceLogRepository balanceLogRepo,
            ILogger<GoodsTransService> logger)
        {
            _goodsTransRepo = goodsTransRepo;
            _orderRepo = orderRepo;
            _userRepo = userRepo;
            _balanceLogRepo = balanceLogRepo;
            _logger = logger;
        }

        public async Task<int> AppendAsync(long groupId, long sellerQQ, long sellerOrderId, long buyerQQ, long buyerOrderId, int goodsId, long amount, decimal price, decimal buyerFee, decimal sellerFee)
        {
            return await _goodsTransRepo.AppendAsync(groupId, sellerQQ, sellerOrderId, buyerQQ, buyerOrderId, goodsId, amount, price, buyerFee, sellerFee);
        }

        public async Task<string> TransItAsync(long botUin, long orderId, long groupId, string groupName, long qq, string name)
        {
            string res = "";
            //新买单或新卖单加入时自动撮合交易
            int orderType = await _orderRepo.GetValueAsync<int>("OrderType", orderId);
            long sellerQQ = await _orderRepo.GetValueAsync<long>("QQ", orderId);
            int goodsId = await _orderRepo.GetValueAsync<int>("GoodsId", orderId);
            long amount = await _orderRepo.GetValueAsync<long>("Amount", orderId);
            decimal price = await _orderRepo.GetValueAsync<decimal>("price", orderId);
            bool isFinish = false;
            while (!isFinish)
            {
                string? id2 = await _orderRepo.GetMatchingOrderIdAsync(goodsId, orderType, price);
                if (!string.IsNullOrEmpty(id2))
                {
                    long orderId2 = long.Parse(id2);
                    long amount2 = await _orderRepo.GetValueAsync<long>("amount", orderId2);
                    decimal price2 = await _orderRepo.GetValueAsync<decimal>("price", orderId2);
                    decimal fee = amount2 * price2 * 0.01m;
                    if (goodsId == 1)
                    {
                        //积分和余额交易
                        using var transWrapper = await SqlHelper.BeginTransactionAsync();
                        var trans = transWrapper.Transaction;
                        try
                        {
                            // 1. 积分操作
                            var addResBuyer = await _userRepo.AddCreditAsync(botUin, groupId, groupName, qq, name, amount, "买入积分", trans);
                            if (!addResBuyer.Success) throw new Exception("买入积分失败");

                            var addResSeller = await _userRepo.AddCreditAsync(botUin, groupId, groupName, sellerQQ, name, -amount2, "卖出积分", trans);
                            if (!addResSeller.Success) throw new Exception("卖出积分失败");

                            // 2. 余额操作
                            var balResBuyer = await _userRepo.AddBalanceAsync(botUin, groupId, groupName, qq, name, -amount * price, "购买积分", trans);
                            if (balResBuyer.Result == -1) throw new Exception("购买积分扣除余额失败");

                            var balResSeller = await _userRepo.AddBalanceAsync(botUin, groupId, groupName, sellerQQ, name, amount * price - fee, "卖出积分", trans);
                            if (balResSeller.Result == -1) throw new Exception("卖出积分增加余额失败");

                            // 3. 手续费日志
                            await _balanceLogRepo.AddLogAsync(botUin, groupId, groupName, sellerQQ, name, -fee, balResSeller.BalanceValue, "交易手续费", trans);

                            // 4. 交易记录
                            await _goodsTransRepo.AppendAsync(groupId, sellerQQ, orderId, qq, orderId2, goodsId, amount, price, 0, fee, trans);

                            await transWrapper.CommitAsync();

                            // 同步缓存
                            await _userRepo.SyncCacheFieldAsync(qq, "Credit", addResBuyer.CreditValue);
                            await _userRepo.SyncCacheFieldAsync(sellerQQ, "Credit", addResSeller.CreditValue);
                            await _userRepo.SyncCacheFieldAsync(qq, "Balance", balResBuyer.BalanceValue);
                            await _userRepo.SyncCacheFieldAsync(sellerQQ, "Balance", balResSeller.BalanceValue);
                            
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
