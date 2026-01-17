using System;
using System.Data;
using System.Threading.Tasks;
using BotWorker.Domain.Models.BotMessages;
using BotWorker.Domain.Repositories;
using Dapper.Contrib.Extensions;
using Microsoft.Extensions.DependencyInjection;

namespace BotWorker.Modules.Games.Gift
{
    [Table("gift_log")]
    public class GiftLog
    {
        [Key]
        public long Id { get; set; }
        public long BotUin { get; set; }
        public long GroupId { get; set; }
        public string GroupName { get; set; } = string.Empty;
        public long UserId { get; set; }
        public string UserName { get; set; } = string.Empty;
        public long RobotOwner { get; set; }
        public string OwnerName { get; set; } = string.Empty;
        public long GiftUserId { get; set; }
        public string GiftUserName { get; set; } = string.Empty;
        public long GiftId { get; set; }
        public string GiftName { get; set; } = string.Empty;
        public int GiftCount { get; set; }
        public long GiftCredit { get; set; }
        public DateTime CreatedAt { get; set; } = DateTime.Now;
    }
}
