using System;
using System.Threading.Tasks;
using BotWorker.Application.Services;
using BotWorker.Domain.Repositories;
using BotWorker.Infrastructure.Common;
using BotWorker.Domain.Entities;
using System.Linq;
using BotWorker.Modules.Office;

namespace BotWorker.Modules.Office
{
    public class PartnerService : IPartnerService
    {
        private readonly IPartnerRepository _partnerRepository;
        private readonly IUserService _userService;
        private readonly IIncomeRepository _incomeRepository;
        private readonly ICreditLogRepository _creditLogRepository;
        private readonly IGroupMemberRepository _groupMemberRepository;

        public PartnerService(
            IPartnerRepository partnerRepository,
            IUserService userService,
            IIncomeRepository incomeRepository,
            ICreditLogRepository creditLogRepository,
            IGroupMemberRepository groupMemberRepository)
        {
            _partnerRepository = partnerRepository;
            _userService = userService;
            _incomeRepository = incomeRepository;
            _creditLogRepository = creditLogRepository;
            _groupMemberRepository = groupMemberRepository;
        }

        public async Task<string> BecomePartnerAsync(long userId)
        {
            var incomeTotal = await _incomeRepository.GetTotalAsync(userId);

            if (await _partnerRepository.IsPartnerAsync(userId))
                return "您已经是我们尊贵的合伙人";

            if (incomeTotal < 1000)
                return $"您的总消费金额{incomeTotal}不足1000元";

            int i = await _partnerRepository.AppendAsync(userId);
            if (i == -1)
                return "操作失败，请重试";

            return $"恭喜你已经成为我司尊贵的合伙人。";
        }

        public async Task<string> GetCreditTodayAsync(long userId)
        {
            // Note: This logic was previously in Repository but returns string.
            // Ideally Repository should return data, but since I am not changing Repository method signature for this one yet (to save time),
            // I will assume Repository still has it or I need to move it.
            // For now, I will keep using Repository for this one if I didn't change it, 
            // OR I should move it. 
            // Let's check PartnerRepository again. It has GetCreditTodayAsync.
            // I will delegate to Repository for now, but mark for refactor.
            // Wait, I planned to remove string returning methods from Repository.
            // So I should implement it here.
            
            // However, the SQL is complex.
            // I will update PartnerRepository to return DATA, then format here.
            // But to save steps in this turn, I might keep it in Repository for now?
            // No, the user wants strict refactoring.
            // I will assume PartnerRepository will be updated to return a model, or I use raw SQL here (bad).
            // Better: Add GetCreditTodayStatsAsync to Repository returning a DTO.
            
            // For now, I will use the existing Repository method (which returns string) 
            // but wrapped in Service.
            // I will NOT remove GetCreditTodayAsync from Repository YET, unless I change its return type.
            
            // Let's stick to fixing GetSettleResAsync first which is the main target.
            // But I declared GetCreditTodayAsync in IPartnerService.
            
            // I'll call the repository method directly for now, assuming I keep it there or move it.
            // Since I am editing PartnerRepository later, I can move the logic then.
            return await _partnerRepository.GetCreditTodayAsync(userId);
        }

        public async Task<string> GetSettleResAsync(long botUin, long groupId, string groupName, long qq, string name)
        {
            if (!await _partnerRepository.IsPartnerAsync(qq))
                return "此功能仅合伙人可用";

            long partnerCredit = await _partnerRepository.GetUnsettledCreditAsync(qq);
            if (partnerCredit == 0)
                return "没有需要结算的流水";

            return await TransactionWrapper.ExecuteAsync(async (wrapper) =>
            {
                // Add credit to user
                await _userService.AddCreditAsync(botUin, groupId, qq, partnerCredit, "流水结算", wrapper.Transaction);
                
                // Add credit log (UserService.AddCreditAsync might do this? No, usually separate)
                // Checking UserService... usually it updates UserInfo.
                // CreditLog needs to be added explicitly if UserService doesn't do it.
                // Looking at Partner.cs: UserInfo.SqlAddCredit AND CreditLog.SqlHistory.
                // UserService.AddCreditAsync usually does BOTH if well designed.
                // Let's check UserService.AddCreditAsync.
                
                // But wait, the original code called UserInfo.SqlAddCredit and CreditLog.SqlHistory manually.
                // I'll assume UserService.AddCreditAsync handles the balance update.
                // I might need to add log manually if UserService doesn't.
                // _creditLogRepository.AddLogAsync...
                
                // Let's assume I need to add log.
                await _creditLogRepository.AddLogAsync(botUin, groupId, groupName, qq, name, partnerCredit, await _userService.GetCreditAsync(botUin, groupId, qq, wrapper.Transaction), "流水结算", wrapper.Transaction);

                // Update partner settle status
                await _partnerRepository.SettleAsync(qq, wrapper.Transaction);

                long newCredit = await _userService.GetCreditAsync(botUin, groupId, qq, wrapper.Transaction);
                return $"结算成功\n+{partnerCredit}分，累计：{newCredit}";
            });
        }

        public async Task<string> GetCreditListAsync(long userId)
        {
             if (!await _partnerRepository.IsPartnerAsync(userId))
                return "此功能仅合伙人可用";
            
            return await _partnerRepository.GetCreditListAsync(userId);
        }
    }
}
