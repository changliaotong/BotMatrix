using System;
using System.Threading.Tasks;
using BotWorker.Domain.Entities;
using BotWorker.Domain.Repositories;
using Dapper.Contrib.Extensions;

namespace BotWorker.Infrastructure.Persistence.Repositories
{
    public class GuildEventRepository : BaseRepository<GuildEvent>, IGuildEventRepository
    {
        public GuildEventRepository(string? connectionString = null) : base("GuildEvent", connectionString)
        {
        }

        public async Task<int> AddAsync(GuildEvent guildEvent)
        {
            if (guildEvent.InsertDate == default)
            {
                guildEvent.InsertDate = DateTime.Now;
            }
            return await base.AddAsync(guildEvent);
        }
    }
}
