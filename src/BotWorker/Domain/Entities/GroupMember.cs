using System;
using System.Data;
using System.Threading.Tasks;
using BotWorker.Domain.Models.BotMessages;
using BotWorker.Domain.Interfaces;
using Dapper.Contrib.Extensions;
using Microsoft.Extensions.DependencyInjection;

namespace BotWorker.Domain.Entities
{
    [Table("group_member")]
    public partial class GroupMember
    {
        [Key]
        public long Id { get; set; }

        public long GroupId { get; set; }
        public long UserId { get; set; }

        [JsonIgnore]
        [HighFrequency]
        public long GroupCredit { get; set; }
        [JsonIgnore]
        [HighFrequency]
        public long GoldCoins { get; set; }
        [JsonIgnore]
        [HighFrequency]
        public long BlackCoins { get; set; }
        [JsonIgnore]
        [HighFrequency]
        public long PurpleCoins { get; set; }
        [JsonIgnore]
        [HighFrequency]
        public long GameCoins { get; set; }
        [JsonIgnore]
        public int SignTimes { get; set; }
        [JsonIgnore]
        public int SignLevel { get; set; }
        [JsonIgnore]
        public DateTime SignDate { get; set; }
        [JsonIgnore]
        public int SignTimesAll { get; set; }

        public DateTime UpdatedAt { get; set; }
    }
}
