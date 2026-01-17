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
    [Table("babies")]
    public class Baby
    {
        [ExplicitKey]
        public Guid Id { get; set; } = Guid.NewGuid();
        public string UserId { get; set; } = string.Empty;
        public string Name { get; set; } = string.Empty;
        public DateTime Birthday { get; set; } = DateTime.Now;
        public int GrowthValue { get; set; } = 0;
        public int DaysOld { get; set; } = 0;
        public int Level { get; set; } = 1;
        public int Points { get; set; } = 0;
        public string Status { get; set; } = "active"; // active, abandoned
        public DateTime CreatedAt { get; set; } = DateTime.Now;
        public DateTime UpdatedAt { get; set; } = DateTime.Now;
        public DateTime LastDailyUpdate { get; set; } = DateTime.MinValue;
    }

    [Table("BabyEvents")]
    public class BabyEvent
    {
        [ExplicitKey]
        public Guid Id { get; set; } = Guid.NewGuid();
        public Guid BabyId { get; set; }
        public string EventType { get; set; } = string.Empty; // birthday, learn, work, interact
        public string Content { get; set; } = string.Empty;
        public DateTime CreatedAt { get; set; } = DateTime.Now;
    }

    [Table("baby_config")]
    public class BabyConfig
    {
        [ExplicitKey]
        public int Id { get; set; } = 1;
        public bool IsEnabled { get; set; } = true;
        public int GrowthRate { get; set; } = 1000;
        public DateTime UpdatedAt { get; set; } = DateTime.Now;
    }
}
