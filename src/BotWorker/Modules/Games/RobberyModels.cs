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
    /// <summary>
    /// 打劫记录
    /// </summary>
    [Table("RobberyRecords")]
    public class RobberyRecord
    {
        private static IRobberyRecordRepository Repository => 
            BotMessage.ServiceProvider?.GetRequiredService<IRobberyRecordRepository>() 
            ?? throw new InvalidOperationException("IRobberyRecordRepository not registered");

        [ExplicitKey]
        public Guid Id { get; set; } = Guid.NewGuid();
        
        public string RobberId { get; set; } = string.Empty; // 打劫者ID
        public string VictimId { get; set; } = string.Empty;  // 被打劫者ID
        public string GroupId { get; set; } = string.Empty;   // 群组ID
        
        public long Amount { get; set; }                      // 涉案金额
        public bool IsSuccess { get; set; }                   // 是否成功
        public string ResultMessage { get; set; } = string.Empty; // 结果描述
        
        public DateTime RobTime { get; set; } = DateTime.Now;

        /// <summary>
        /// 获取用户最后一次打劫时间
        /// </summary>
        public static async Task<DateTime> GetLastRobTimeAsync(string userId)
        {
            return await Repository.GetLastRobTimeAsync(userId);
        }

        /// <summary>
        /// 获取用户被打劫后的保护到期时间
        /// </summary>
        public static async Task<DateTime> GetProtectionEndTimeAsync(string userId)
        {
            return await Repository.GetProtectionEndTimeAsync(userId);
        }

        public async Task<bool> InsertAsync(IDbTransaction? trans = null)
        {
            return await Repository.InsertAsync(this, trans);
        }
    }
}
