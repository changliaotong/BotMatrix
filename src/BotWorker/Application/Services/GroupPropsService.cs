using System;
using System.Threading.Tasks;
using BotWorker.Domain.Entities;
using BotWorker.Domain.Interfaces;
using BotWorker.Domain.Repositories;
using BotWorker.Infrastructure.Persistence;
using Microsoft.Extensions.Logging;

namespace BotWorker.Application.Services
{
    public class GroupPropsService : IGroupPropsService
    {
        private readonly IGroupPropsRepository _groupPropsRepo;
        private readonly IPropRepository _propRepo;
        private readonly IUserRepository _userRepo;
        private readonly IGroupRepository _groupRepo;
        private readonly ILogger<GroupPropsService> _logger;

        public const string PropClosed = "道具系统已关闭";

        public GroupPropsService(
            IGroupPropsRepository groupPropsRepo,
            IPropRepository propRepo,
            IUserRepository userRepo,
            IGroupRepository groupRepo,
            ILogger<GroupPropsService> logger)
        {
            _groupPropsRepo = groupPropsRepo;
            _propRepo = propRepo;
            _userRepo = userRepo;
            _groupRepo = groupRepo;
            _logger = logger;
        }

        public async Task<long> GetIdAsync(long groupId, long qq, long propId)
        {
            return await _groupPropsRepo.GetIdAsync(groupId, qq, propId);
        }

        public async Task<bool> HavePropAsync(long groupId, long userId, long propId)
        {
            return await _groupPropsRepo.HavePropAsync(groupId, userId, propId);
        }

        public async Task<int> UsePropAsync(long groupId, long userId, long propId, long qqProp)
        {
            return await _groupPropsRepo.UsePropAsync(groupId, userId, propId, qqProp);
        }

        public async Task<string> GetMyPropListAsync(long groupId, long userId)
        {
            if (await IsClosedAsync(groupId)) return PropClosed;
            return await _groupPropsRepo.GetMyPropListAsync(groupId, userId);
        }

        public async Task<bool> IsClosedAsync(long groupId)
        {
            return !await _groupRepo.GetBoolAsync("IsProp", groupId);
        }

        public async Task<string> GetBuyResAsync(long botUin, long groupId, string groupName, long qq, string name, string cmdPara)
        {
            if (await IsClosedAsync(groupId)) return PropClosed;

            if (string.IsNullOrEmpty(cmdPara) || cmdPara == "道具")
                return await _propRepo.GetPropListAsync();
            else
            {
                long prop_id = await _propRepo.GetIdAsync(cmdPara);
                if (prop_id != 0)
                {
                    long credit_value = await _userRepo.GetCreditAsync(botUin, groupId, qq);
                    int prop_price = await _propRepo.GetPropPriceAsync(prop_id);
                    if (credit_value < prop_price)
                        return $"您的积分{credit_value}不足{prop_price}";
                    
                    using var wrapper = await SqlHelper.BeginTransactionAsync();
                    try
                    {
                        // 1. 通用加积分函数 (含日志记录)
                        var res = await _userRepo.AddCreditAsync(botUin, groupId, groupName, qq, name, -prop_price, $"购买道具:{prop_id}", wrapper.Transaction);
                        if (!res.Success) throw new Exception("更新积分失败");

                        // 2. 插入道具购买记录
                        await _groupPropsRepo.InsertAsync(groupId, qq, prop_id, wrapper.Transaction);

                        await wrapper.CommitAsync();

                        await _userRepo.SyncCreditCacheAsync(botUin, groupId, qq, res.CreditValue);

                        return $"购买道具成功\n积分：-{prop_price}，累计：{res.CreditValue}";
                    }
                    catch (Exception ex)
                    {
                        await wrapper.RollbackAsync();
                        _logger.LogError(ex, "[GetBuyRes Error] {Message}", ex.Message);
                        return "操作失败，请重试";
                    }
                }
                else
                    return "没有此道具";
            }
        }
    }
}
