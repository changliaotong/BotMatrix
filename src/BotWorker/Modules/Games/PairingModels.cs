using System;
using Dapper.Contrib.Extensions;

namespace BotWorker.Modules.Games
{
    #region 配对系统数据模型

    /// <summary>
    /// 用户社交资料
    /// </summary>
    [Table("user_pairing_profiles")]
    public class UserPairingProfile
    {
        [ExplicitKey]
        public Guid Id { get; set; } = Guid.NewGuid();
        public string UserId { get; set; } = string.Empty;
        public string Nickname { get; set; } = string.Empty;
        public string Gender { get; set; } = "未知"; // 男, 女, 隐藏
        public string Zodiac { get; set; } = "未知"; // 星座
        public string Intro { get; set; } = "这个人很懒，什么都没留下。";
        public bool IsLooking { get; set; } = true; // 是否正在寻找配对
        public DateTime LastActive { get; set; } = DateTime.Now;
        public DateTime CreatedAt { get; set; } = DateTime.Now;
    }

    /// <summary>
    /// 配对记录 (CP记录)
    /// </summary>
    [Table("pairing_records")]
    public class PairingRecord
    {
        [ExplicitKey]
        public Guid Id { get; set; } = Guid.NewGuid();
        public string User1Id { get; set; } = string.Empty;
        public string User2Id { get; set; } = string.Empty;
        public string Status { get; set; } = "pairing"; // pairing (匹配中), coupled (已成对), broken (已解绑)
        public DateTime PairDate { get; set; } = DateTime.Now;
    }

    #endregion
}
