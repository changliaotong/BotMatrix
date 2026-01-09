using BotWorker.Infrastructure.Persistence.ORM;
using System.Reflection;

namespace BotWorker.Modules.Games
{
    #region 配对系统数据模型

    /// <summary>
    /// 用户社交资料
    /// </summary>
    public class UserPairingProfile : MetaData<UserPairingProfile>
    {
        [BotWorker.Infrastructure.Utils.Schema.Attributes.PrimaryKey]
        public Guid Id { get; set; } = Guid.NewGuid();
        public string UserId { get; set; } = string.Empty;
        public string Nickname { get; set; } = string.Empty;
        public string Gender { get; set; } = "未知"; // 男, 女, 隐藏
        public string Zodiac { get; set; } = "未知"; // 星座
        public string Intro { get; set; } = "这个人很懒，什么都没留下。";
        public bool IsLooking { get; set; } = true; // 是否正在寻找配对
        public DateTime LastActive { get; set; } = DateTime.Now;
        public DateTime CreatedAt { get; set; } = DateTime.Now;

        public override string TableName => "UserPairingProfiles";
        public override string KeyField => "Id";

        public static async Task<UserPairingProfile?> GetByUserIdAsync(string userId)
        {
            return (await QueryWhere("UserId = @p1", SqlParams(("@p1", userId)))).FirstOrDefault();
        }

        public static async Task<List<UserPairingProfile>> GetActiveSeekersAsync(int limit = 10)
        {
            return await QueryWhere("IsLooking = 1 ORDER BY LastActive DESC LIMIT @p1", SqlParams(("@p1", limit)));
        }
    }

    /// <summary>
    /// 配对记录 (CP记录)
    /// </summary>
    public class PairingRecord : MetaData<PairingRecord>
    {
        [BotWorker.Infrastructure.Utils.Schema.Attributes.PrimaryKey]
        public Guid Id { get; set; } = Guid.NewGuid();
        public string User1Id { get; set; } = string.Empty;
        public string User2Id { get; set; } = string.Empty;
        public string Status { get; set; } = "pairing"; // pairing (匹配中), coupled (已成对), broken (已解绑)
        public DateTime PairDate { get; set; } = DateTime.Now;

        public override string TableName => "PairingRecords";
        public override string KeyField => "Id";

        public static async Task<PairingRecord?> GetCurrentPairAsync(string userId)
        {
            return (await QueryWhere("(User1Id = @p1 OR User2Id = @p1) AND Status = 'coupled'", SqlParams(("@p1", userId)))).FirstOrDefault();
        }
    }

    #endregion
}
