using System;
using Dapper.Contrib.Extensions;

namespace BotWorker.Modules.Games
{
    #region 结婚系统数据模型

    [Table("user_marriages")]
    public class UserMarriage
    {
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
    }

    [Table("marriage_proposals")]
    public class MarriageProposal
    {
        [ExplicitKey]
        public Guid Id { get; set; } = Guid.NewGuid();
        public string ProposerId { get; set; } = string.Empty;
        public string RecipientId { get; set; } = string.Empty;
        public string Status { get; set; } = "pending"; // pending, accepted, rejected
        public DateTime CreatedAt { get; set; } = DateTime.Now;
        public DateTime UpdatedAt { get; set; } = DateTime.Now;
    }

    [Table("wedding_items")]
    public class WeddingItem
    {
        [ExplicitKey]
        public Guid Id { get; set; } = Guid.NewGuid();
        public string UserId { get; set; } = string.Empty;
        public string ItemType { get; set; } = string.Empty; // dress, ring
        public string Name { get; set; } = string.Empty;
        public int Price { get; set; } = 0;
        public DateTime CreatedAt { get; set; } = DateTime.Now;
    }

    [Table("sweet_hearts")]
    public class SweetHeart
    {
        [ExplicitKey]
        public Guid Id { get; set; } = Guid.NewGuid();
        public string SenderId { get; set; } = string.Empty;
        public string RecipientId { get; set; } = string.Empty;
        public int Amount { get; set; } = 0;
        public DateTime CreatedAt { get; set; } = DateTime.Now;
    }

    #endregion

    public enum MarriageAction { Propose, Accept, Reject, Divorce, SendSweets, SendRedPacket }
}
