using BotWorker.Infrastructure.Persistence.ORM;
using BotWorker.Infrastructure.Utils.Schema.Attributes;
using System;
using System.Collections.Generic;
using System.Linq;
using System.Threading.Tasks;

namespace BotWorker.Modules.Games
{
    /// <summary>
    /// 礼物配置模型
    /// </summary>
    public class GiftStoreItem : MetaDataGuid<GiftStoreItem>
    {
        public string GiftName { get; set; } = string.Empty;
        public long GiftCredit { get; set; }
        public string GiftUrl { get; set; } = string.Empty;
        public string GiftImage { get; set; } = string.Empty;
        public int GiftType { get; set; } // 1: 普通, 2: 高级
        public bool IsValid { get; set; }

        public override string TableName => "GiftStoreItem";
        public override string KeyField => "Id";

        public static async Task<List<GiftStoreItem>> GetValidGiftsAsync()
        {
            return await QueryWhere("IsValid = 1", (System.Data.IDbTransaction?)null);
        }

        public static async Task<GiftStoreItem?> GetByNameAsync(string name)
        {
            return (await QueryWhere("GiftName = @p1", (System.Data.IDbTransaction?)null, SqlParams(("@p1", name)))).FirstOrDefault();
        }
    }

    /// <summary>
    /// 用户背包模型
    /// </summary>
    public class GiftBackpack : MetaDataGuid<GiftBackpack>
    {
        public string UserId { get; set; } = string.Empty;
        public long GiftId { get; set; } // 修改为 long 以匹配 GiftStoreItem.Id
        public int ItemCount { get; set; } // 重命名以避免与 MetaData.Count() 冲突

        public override string TableName => "GiftBackpack";
        public override string KeyField => "Id";

        public static async Task<List<GiftBackpack>> GetUserBackpackAsync(string userId)
        {
            return await QueryWhere("UserId = @p1 AND ItemCount > 0", (System.Data.IDbTransaction?)null, SqlParams(("@p1", userId)));
        }

        public static async Task<GiftBackpack?> GetItemAsync(string userId, long giftId)
        {
            return (await QueryWhere("UserId = @p1 AND GiftId = @p2", (System.Data.IDbTransaction?)null, SqlParams(("@p1", userId), ("@p2", giftId)))).FirstOrDefault();
        }
    }

    /// <summary>
    /// 礼物赠送记录
    /// </summary>
    public class GiftRecord : MetaDataGuid<GiftRecord>
    {
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

        public override string TableName => "GiftLog";
        public override string KeyField => "Id";
    }
}
