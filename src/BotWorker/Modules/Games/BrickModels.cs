using System;
using System.Collections.Generic;
using System.Data;
using System.Linq;
using System.Threading.Tasks;
using BotWorker.Domain.Models.BotMessages;
using BotWorker.Domain.Repositories;
using Dapper.Contrib.Extensions;
using Microsoft.Extensions.DependencyInjection;

namespace BotWorker.Modules.Games
{
    /// <summary>
    /// 闷砖记录
    /// </summary>
    [Table("BrickRecords")]
    public class BrickRecord
    {
        private static IBrickRecordRepository Repository => 
            BotMessage.ServiceProvider?.GetRequiredService<IBrickRecordRepository>() 
            ?? throw new InvalidOperationException("IBrickRecordRepository not registered");

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

        /// <summary>
        /// 获取用户最后一次拍砖时间
        /// </summary>
        public static async Task<DateTime> GetLastActionTimeAsync(string userId)
        {
            return await Repository.GetLastActionTimeAsync(userId);
        }

        /// <summary>
        /// 获取拍砖排行榜 (拍人最多的)
        /// </summary>
        public static async Task<List<(string UserId, int Count)>> GetTopAttackersAsync(int limit = 10)
        {
            return await Repository.GetTopAttackersAsync(limit);
        }

        public async Task<bool> InsertAsync(IDbTransaction? trans = null)
        {
            return await Repository.InsertAsync(this, trans);
        }
    }
}
