using System;
using Dapper.Contrib.Extensions;

namespace BotWorker.Modules.Games
{
    /// <summary>
    /// 闷砖记录
    /// </summary>
    [Table("brick_records")]
    public class BrickRecord
    {
        [ExplicitKey]
        public Guid Id { get; set; } = Guid.NewGuid();
        
        public string AttackerId { get; set; } = string.Empty; // 拍砖人ID
        public string TargetId { get; set; } = string.Empty;   // 被拍人ID
        public string GroupId { get; set; } = string.Empty;    // 群组ID
        
        public bool IsSuccess { get; set; }                    // 是否成功
        public int MuteSeconds { get; set; }                   // 禁言时长（秒）
        public long CreditChange { get; set; }                 // 积分变动
        
        public DateTime ActionTime { get; set; } = DateTime.Now;

        [Write(false)] public string RankUserId { get; set; } = string.Empty;
        [Write(false)] public int RankCount { get; set; }
    }
}
