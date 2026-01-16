using System;
using System.Data;
using System.Linq;
using System.Threading.Tasks;
using BotWorker.Domain.Models.BotMessages;
using BotWorker.Domain.Repositories;
using Dapper.Contrib.Extensions;
using Microsoft.Extensions.DependencyInjection;

namespace BotWorker.Modules.Games
{
    #region 结婚系统数据模型

    [Table("UserMarriages")]
    public class UserMarriage
    {
        private static IUserMarriageRepository Repository => 
            BotMessage.ServiceProvider?.GetRequiredService<IUserMarriageRepository>() 
            ?? throw new InvalidOperationException("IUserMarriageRepository not registered");

        [ExplicitKey]
        public Guid Id { get; set; } = Guid.NewGuid();
        public string UserId { get; set; } = string.Empty;
        public string SpouseId { get; set; } = string.Empty;
        public DateTime MarriageDate { get; set; } = DateTime.MinValue;
        public DateTime DivorceDate { get; set; } = DateTime.MinValue;
        public string Status { get; set; } = "single"; // single, married, divorced
        public int SweetsCount { get; set; } = 0;
        public int RedPacketsCount { get; set; } = 0;
        public int SweetHearts { get; set; } = 0;
        public DateTime CreatedAt { get; set; } = DateTime.Now;
        public DateTime UpdatedAt { get; set; } = DateTime.Now;

        public static async Task<UserMarriage?> GetByUserIdAsync(string userId)
        {
            return await Repository.GetByUserIdAsync(userId);
        }

        public static async Task<UserMarriage> GetOrCreateAsync(string userId)
        {
            return await Repository.GetOrCreateAsync(userId);
        }

        public async Task<bool> InsertAsync(IDbTransaction? trans = null)
        {
            return await Repository.InsertAsync(this, trans);
        }

        public async Task<bool> UpdateAsync(IDbTransaction? trans = null)
        {
            return await Repository.UpdateAsync(this, trans);
        }

        public static async Task<BotWorker.Infrastructure.Persistence.TransactionWrapper> BeginTransactionAsync()
        {
            return await Repository.BeginTransactionAsync();
        }

        public static async Task UpdateMarriageStatusAsync(string userId, string spouseId, string status, DateTime marriageDate, IDbTransaction? trans = null)
        {
            await Repository.UpdateMarriageStatusAsync(userId, spouseId, status, marriageDate, trans);
        }

        public static async Task DivorceAsync(string userId, string spouseId, DateTime divorceDate, IDbTransaction? trans = null)
        {
            await Repository.DivorceAsync(userId, spouseId, divorceDate, trans);
        }
    }

    [Table("MarriageProposals")]
    public class MarriageProposal
    {
        private static IMarriageProposalRepository Repository => 
            BotMessage.ServiceProvider?.GetRequiredService<IMarriageProposalRepository>() 
            ?? throw new InvalidOperationException("IMarriageProposalRepository not registered");

        [ExplicitKey]
        public Guid Id { get; set; } = Guid.NewGuid();
        public string ProposerId { get; set; } = string.Empty;
        public string RecipientId { get; set; } = string.Empty;
        public string Status { get; set; } = "pending"; // pending, accepted, rejected
        public DateTime CreatedAt { get; set; } = DateTime.Now;
        public DateTime UpdatedAt { get; set; } = DateTime.Now;

        public static async Task<MarriageProposal?> GetPendingAsync(string recipientId)
        {
            return await Repository.GetPendingAsync(recipientId);
        }

        public async Task<bool> InsertAsync(IDbTransaction? trans = null)
        {
            return await Repository.InsertAsync(this, trans);
        }

        public async Task<bool> UpdateAsync(IDbTransaction? trans = null)
        {
            return await Repository.UpdateAsync(this, trans);
        }

        public static async Task UpdateStatusAsync(Guid id, string status, IDbTransaction? trans = null)
        {
            await Repository.UpdateStatusAsync(id, status, trans);
        }
    }

    [Table("WeddingItems")]
    public class WeddingItem
    {
        private static IWeddingItemRepository Repository => 
            BotMessage.ServiceProvider?.GetRequiredService<IWeddingItemRepository>() 
            ?? throw new InvalidOperationException("IWeddingItemRepository not registered");

        [ExplicitKey]
        public Guid Id { get; set; } = Guid.NewGuid();
        public string UserId { get; set; } = string.Empty;
        public string ItemType { get; set; } = string.Empty; // dress, ring
        public string Name { get; set; } = string.Empty;
        public int Price { get; set; } = 0;
        public DateTime CreatedAt { get; set; } = DateTime.Now;

        public async Task<bool> InsertAsync(IDbTransaction? trans = null)
        {
            return await Repository.InsertAsync(this, trans);
        }
    }

    [Table("SweetHearts")]
    public class SweetHeart
    {
        private static ISweetHeartRepository Repository => 
            BotMessage.ServiceProvider?.GetRequiredService<ISweetHeartRepository>() 
            ?? throw new InvalidOperationException("ISweetHeartRepository not registered");

        [ExplicitKey]
        public Guid Id { get; set; } = Guid.NewGuid();
        public string SenderId { get; set; } = string.Empty;
        public string RecipientId { get; set; } = string.Empty;
        public int Amount { get; set; } = 0;
        public DateTime CreatedAt { get; set; } = DateTime.Now;

        public async Task<bool> InsertAsync(IDbTransaction? trans = null)
        {
            return await Repository.InsertAsync(this, trans);
        }
    }

    #endregion

    public enum MarriageAction { Propose, Accept, Reject, Divorce, SendSweets, SendRedPacket }
}
