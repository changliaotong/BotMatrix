using System;
using System.Collections.Generic;
using System.Threading.Tasks;
using BotWorker.Common;
using BotWorker.Domain.Repositories;
using Dapper.Contrib.Extensions;
using Microsoft.Extensions.DependencyInjection;

namespace BotWorker.Domain.Entities
{
    [Table("GuildEvent")]
    public class GuildEvent
    {
        private static IGuildEventRepository Repository => GlobalConfig.ServiceProvider!.GetRequiredService<IGuildEventRepository>();

        [Key]
        public long Id { get; set; }
        public long GroupId { get; set; }
        public string GroupName { get; set; } = string.Empty;
        public long BotUin { get; set; }
        public string BotName { get; set; } = string.Empty;
        public long UserId { get; set; }
        public string UserName { get; set; } = string.Empty;
        public string EventType { get; set; } = string.Empty;
        public string EventName { get; set; } = string.Empty;
        public string EventInfo { get; set; } = string.Empty;
        public DateTime InsertDate { get; set; }

        // Adjusted signature to be compatible with usage, but ignoring fields param effectively
        public static async Task<int> AppendAsync(GuildEvent @event, params string[] fields)
        {
            return await Repository.AddAsync(@event);
        }
    }
}
