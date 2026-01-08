using BotWorker.Infrastructure.Persistence.ORM;
using System.Reflection;

namespace BotWorker.Modules.Games
{
    #region 结婚系统数据模型

    public class UserMarriage : MetaData<UserMarriage>
    {
        [BotWorker.Infrastructure.Utils.Schema.Attributes.PrimaryKey]
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

        public override string TableName => "UserMarriages";
        public override string KeyField => "Id";

        public static async Task<UserMarriage?> GetByUserIdAsync(string userId)
        {
            return (await QueryAsync("WHERE UserId = @UserId", new { UserId = userId })).FirstOrDefault();
        }

        public static async Task<UserMarriage> GetOrCreateAsync(string userId)
        {
            var m = await GetByUserIdAsync(userId);
            if (m == null)
            {
                m = new UserMarriage { UserId = userId, Status = "single" };
                await m.InsertAsync();
            }
            return m;
        }
    }

    public class MarriageProposal : MetaData<MarriageProposal>
    {
        [BotWorker.Infrastructure.Utils.Schema.Attributes.PrimaryKey]
        public Guid Id { get; set; } = Guid.NewGuid();
        public string ProposerId { get; set; } = string.Empty;
        public string RecipientId { get; set; } = string.Empty;
        public string Status { get; set; } = "pending"; // pending, accepted, rejected
        public DateTime CreatedAt { get; set; } = DateTime.Now;
        public DateTime UpdatedAt { get; set; } = DateTime.Now;

        public override string TableName => "MarriageProposals";
        public override string KeyField => "Id";

        public static async Task<MarriageProposal?> GetPendingAsync(string recipientId)
        {
            return (await QueryAsync("WHERE RecipientId = @RecipientId AND Status = 'pending' ORDER BY CreatedAt DESC", new { RecipientId = recipientId })).FirstOrDefault();
        }
    }

    public class WeddingItem : MetaData<WeddingItem>
    {
        [BotWorker.Infrastructure.Utils.Schema.Attributes.PrimaryKey]
        public Guid Id { get; set; } = Guid.NewGuid();
        public string UserId { get; set; } = string.Empty;
        public string ItemType { get; set; } = string.Empty; // dress, ring
        public string Name { get; set; } = string.Empty;
        public int Price { get; set; } = 0;
        public DateTime CreatedAt { get; set; } = DateTime.Now;

        public override string TableName => "WeddingItems";
        public override string KeyField => "Id";
    }

    public class SweetHeart : MetaData<SweetHeart>
    {
        [BotWorker.Infrastructure.Utils.Schema.Attributes.PrimaryKey]
        public Guid Id { get; set; } = Guid.NewGuid();
        public string SenderId { get; set; } = string.Empty;
        public string RecipientId { get; set; } = string.Empty;
        public int Amount { get; set; } = 0;
        public DateTime CreatedAt { get; set; } = DateTime.Now;

        public override string TableName => "SweetHearts";
        public override string KeyField => "Id";
    }

    #endregion

    public enum MarriageAction { Propose, Accept, Reject, Divorce, SendSweets, SendRedPacket }
}
