using BotWorker.Infrastructure.Persistence.ORM;

namespace BotWorker.Modules.Games
{
    /// <summary>
    /// 闷砖记录
    /// </summary>
    public class BrickRecord : MetaData<BrickRecord>
    {
        [BotWorker.Infrastructure.Utils.Schema.Attributes.PrimaryKey]
        public Guid Id { get; set; } = Guid.NewGuid();
        
        public string AttackerId { get; set; } = string.Empty; // 拍砖人ID
        public string TargetId { get; set; } = string.Empty;   // 被拍人ID
        public string GroupId { get; set; } = string.Empty;    // 群组ID
        
        public bool IsSuccess { get; set; }                    // 是否成功
        public int MuteSeconds { get; set; }                   // 禁言时长（秒）
        public long CreditChange { get; set; }                 // 积分变动
        
        public DateTime ActionTime { get; set; } = DateTime.Now;

        [DbIgnore] public string RankUserId { get; set; } = string.Empty;
        [DbIgnore] public int RankCount { get; set; }

        public override string TableName => "BrickRecords";
        public override string KeyField => "Id";

        /// <summary>
        /// 获取用户最后一次拍砖时间
        /// </summary>
        public static async Task<DateTime> GetLastActionTimeAsync(string userId)
        {
            string topClause = SqlTop(1);
            string limitClause = SqlLimit(1);
            var last = (await QueryWhere($"{topClause} AttackerId = @p1 ORDER BY ActionTime DESC {limitClause}", SqlParams(("@p1", userId)))).FirstOrDefault();
            return last?.ActionTime ?? DateTime.MinValue;
        }

        /// <summary>
        /// 获取拍砖排行榜 (拍人最多的)
        /// </summary>
        public static async Task<List<(string UserId, int Count)>> GetTopAttackersAsync(int limit = 10)
        {
            string topClause = SqlTop(limit);
            string limitClause = SqlLimit(limit);
            string sql = $"SELECT {topClause} AttackerId as RankUserId, COUNT(*) as RankCount FROM BrickRecords WHERE IsSuccess = 1 GROUP BY AttackerId ORDER BY RankCount DESC {limitClause}";
            var results = await QueryAsync(sql);
            return results.Select(r => (r.RankUserId, r.RankCount)).ToList();
        }
    }
}
