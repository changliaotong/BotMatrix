using System;
using Dapper.Contrib.Extensions;

namespace BotWorker.Modules.Games
{
    /// <summary>
    /// 礼物配置模型
    /// </summary>
    [Table("gift_store_item")]
    public class GiftStoreItem
    {
        [Key]
        public long Id { get; set; }
        public string GiftName { get; set; } = string.Empty;
        public long GiftCredit { get; set; }
        public string GiftUrl { get; set; } = string.Empty;
        public string GiftImage { get; set; } = string.Empty;
        public int GiftType { get; set; } // 1: 普通, 2: 高级
        public bool IsValid { get; set; }
    }

    /// <summary>
    /// 用户背包模型
    /// </summary>
    [Table("gift_backpack")]
    public class GiftBackpack
    {
        [Key]
        public long Id { get; set; }
        public string UserId { get; set; } = string.Empty;
        public long GiftId { get; set; } 
        public int ItemCount { get; set; } // 重命名以避免与 MetaData.Count() 冲突
    }

    /// <summary>
    /// 礼物赠送记录
    /// </summary>
    [Table("gift_log")]
    public class GiftRecord
    {
        [Key]
        public long Id { get; set; }
        public long BotUin { get; set; }
        public long GroupId { get; set; }
        public string GroupName { get; set; } = string.Empty;
        public string UserId { get; set; } = string.Empty;
        public string UserName { get; set; } = string.Empty;
        public string GiftUserId { get; set; } = string.Empty;
        public string GiftUserName { get; set; } = string.Empty;
        public long GiftId { get; set; }
        public string GiftName { get; set; } = string.Empty;
        public int GiftCount { get; set; }
        public long GiftCredit { get; set; }
        public DateTime InsertDate { get; set; } = DateTime.Now;
    }
}
