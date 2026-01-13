using BotWorker.Infrastructure.Persistence.ORM;

namespace BotWorker.Modules.Games
{
    /// <summary>
    /// 打劫记录
    /// </summary>
    public class RobberyRecord : MetaData<RobberyRecord>
    {
        [BotWorker.Infrastructure.Utils.Schema.Attributes.PrimaryKey]
        public Guid Id { get; set; } = Guid.NewGuid();
        
        public string RobberId { get; set; } = string.Empty; // 打劫者ID
        public string VictimId { get; set; } = string.Empty;  // 被打劫者ID
        public string GroupId { get; set; } = string.Empty;   // 群组ID
        
        public long Amount { get; set; }                      // 涉案金额
        public bool IsSuccess { get; set; }                   // 是否成功
        public string ResultMessage { get; set; } = string.Empty; // 结果描述
        
        public DateTime RobTime { get; set; } = DateTime.Now;

        public override string TableName => "RobberyRecords";
        public override string KeyField => "Id";

        /// <summary>
        /// 获取用户最后一次打劫时间
        /// </summary>
        public static async Task<DateTime> GetLastRobTimeAsync(string userId)
        {
            string topClause = SqlTop(1);
            string limitClause = SqlLimit(1);
            var last = (await QueryWhere($"{topClause} RobberId = @p1 ORDER BY RobTime DESC {limitClause}", SqlParams(("@p1", userId)))).FirstOrDefault();
            return last?.RobTime ?? DateTime.MinValue;
        }

        /// <summary>
        /// 获取用户被打劫后的保护到期时间
        /// </summary>
        public static async Task<DateTime> GetProtectionEndTimeAsync(string userId)
        {
            string topClause = SqlTop(1);
            string limitClause = SqlLimit(1);
            var last = (await QueryWhere($"{topClause} VictimId = @p1 AND IsSuccess = 1 ORDER BY RobTime DESC {limitClause}", SqlParams(("@p1", userId)))).FirstOrDefault();
            if (last == null) return DateTime.MinValue;
            // 被成功打劫后保护 30 分钟
            return last.RobTime.AddMinutes(30);
        }
    }
}
