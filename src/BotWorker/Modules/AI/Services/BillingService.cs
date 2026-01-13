using System;
using System.Collections.Generic;
using System.Linq;
using System.Threading.Tasks;
using BotWorker.Modules.AI.Interfaces;
using BotWorker.Modules.AI.Models.Billing;
using Microsoft.Extensions.Logging;

namespace BotWorker.Modules.AI.Services
{
    public class BillingService : IBillingService
    {
        private readonly IWalletRepository _walletRepository;
        private readonly ILeaseResourceRepository _resourceRepository;
        private readonly ILeaseContractRepository _contractRepository;
        private readonly IBillingTransactionRepository _transactionRepository;
        private readonly ILogger<BillingService> _logger;

        public BillingService(
            IWalletRepository walletRepository,
            ILeaseResourceRepository resourceRepository,
            ILeaseContractRepository contractRepository,
            IBillingTransactionRepository transactionRepository,
            ILogger<BillingService> logger)
        {
            _walletRepository = walletRepository;
            _resourceRepository = resourceRepository;
            _contractRepository = contractRepository;
            _transactionRepository = transactionRepository;
            _logger = logger;
        }

        public async Task<Wallet> GetOrCreateWalletAsync(long ownerId)
        {
            var wallet = await _walletRepository.GetByOwnerIdAsync(ownerId);
            if (wallet == null)
            {
                wallet = new Wallet
                {
                    OwnerId = ownerId,
                    Balance = 0,
                    FrozenBalance = 0,
                    TotalSpent = 0,
                    Currency = "CNY",
                    Config = "{}"
                };
                var id = await _walletRepository.AddAsync(wallet);
                wallet.Id = id;
            }
            return wallet;
        }

        public async Task<bool> HasSufficientBalanceAsync(long ownerId, decimal requiredAmount)
        {
            var wallet = await GetOrCreateWalletAsync(ownerId);
            return wallet.Balance >= requiredAmount;
        }

        public async Task<bool> ConsumeAsync(long ownerId, decimal amount, long? relatedId = null, string? relatedType = null, string? remark = null)
        {
            try
            {
                var wallet = await GetOrCreateWalletAsync(ownerId);
                if (wallet.Balance < amount)
                {
                    _logger.LogWarning("[Billing] Insufficient balance for owner {OwnerId}. Required: {Amount}, Available: {Balance}", 
                        ownerId, amount, wallet.Balance);
                    return false;
                }

                wallet.Balance -= amount;
                wallet.TotalSpent += amount;
                await _walletRepository.UpdateAsync(wallet);

                var transaction = new BillingTransaction
                {
                    WalletId = wallet.Id,
                    Type = "consume",
                    Amount = -amount,
                    RelatedId = relatedId,
                    RelatedType = relatedType,
                    Remark = remark,
                    CreatedAt = DateTime.Now
                };
                await _transactionRepository.AddAsync(transaction);

                return true;
            }
            catch (Exception ex)
            {
                _logger.LogError(ex, "[Billing] Error processing consumption for owner {OwnerId}", ownerId);
                return false;
            }
        }

        public async Task<bool> RechargeAsync(long ownerId, decimal amount, string? remark = null)
        {
            try
            {
                var wallet = await GetOrCreateWalletAsync(ownerId);
                wallet.Balance += amount;
                await _walletRepository.UpdateAsync(wallet);

                var transaction = new BillingTransaction
                {
                    WalletId = wallet.Id,
                    Type = "recharge",
                    Amount = amount,
                    Remark = remark,
                    CreatedAt = DateTime.Now
                };
                await _transactionRepository.AddAsync(transaction);

                return true;
            }
            catch (Exception ex)
            {
                _logger.LogError(ex, "[Billing] Error processing recharge for owner {OwnerId}", ownerId);
                return false;
            }
        }

        public async Task<bool> HasActiveLeaseAsync(long tenantId, string resourceType)
        {
            try
            {
                var contracts = await _contractRepository.GetByTenantIdAsync(tenantId);
                var activeContracts = contracts.Where(c => c.Status == "active" && (c.EndTime == null || c.EndTime > DateTime.Now));

                foreach (var contract in activeContracts)
                {
                    var resource = await _resourceRepository.GetByIdAsync(contract.ResourceId);
                    if (resource != null && resource.Type.Equals(resourceType, StringComparison.OrdinalIgnoreCase) && resource.IsActive)
                    {
                        return true;
                    }
                }
                return false;
            }
            catch (Exception ex)
            {
                _logger.LogError(ex, "[Billing] Error checking active lease for tenant {TenantId}", tenantId);
                return false;
            }
        }
    }
}
