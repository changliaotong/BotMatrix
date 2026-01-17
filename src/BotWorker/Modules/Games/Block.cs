using System;
using System.Data;
using Dapper.Contrib.Extensions;

namespace BotWorker.Modules.Games
{
    // 区块链+游戏
    [Table("block")]
    public class Block
    {
        [Key]
        public long Id { get; set; }
        public long PrevId { get; set; }
        public string PrevHash { get; set; } = string.Empty;
        public string PrevRes { get; set; } = string.Empty;
        public long BotUin { get; set; }
        public long GroupId { get; set; }
        public string GroupName { get; set; } = string.Empty;
        public long UserId { get; set; }
        public string UserName { get; set; } = string.Empty;
        public string BlockInfo { get; set; } = string.Empty;
        public string BlockSecret { get; set; } = string.Empty;
        public int BlockNum { get; set; }
        public string BlockRes { get; set; } = string.Empty;
        public string BlockRand { get; set; } = string.Empty;
        public string BlockHash { get; set; } = string.Empty;
        public int IsOpen { get; set; }
        public DateTime? OpenDate { get; set; }
        public long OpenBotUin { get; set; }
        public long OpenUserId { get; set; }
        public string OpenUserName { get; set; } = string.Empty;
    }

    [Table("block_random")]
    public class BlockRandom
    {
        [Key]
        public int Id { get; set; }
        public int BlockNum { get; set; }
    }

    [Table("block_type")]
    public class BlockType
    {
        [Key]
        public int Id { get; set; }
        public string TypeName { get; set; } = string.Empty;
        public decimal BlockOdds { get; set; }
    }

    [Table("block_win")]
    public class BlockWin
    {
        [Key]
        public int Id { get; set; }
        public int TypeId { get; set; }
        public int BlockNum { get; set; }
        public int IsWin { get; set; }
    }
}
