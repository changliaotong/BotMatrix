using System.Threading.Tasks;
using BotWorker.Common;
using BotWorker.Domain.Entities;
using BotWorker.Domain.Repositories;
using Dapper;

namespace BotWorker.Infrastructure.Persistence.Repositories
{
    public class BotHintsRepository : BaseRepository<BotHints>, IBotHintsRepository
    {
        public BotHintsRepository() : base("BotHints", GlobalConfig.BaseInfoConnection)
        {
        }

        public async Task<string> GetHintAsync(string cmd)
        {
            return await GetValueAsync<string>("Hint", "WHERE Cmd = @cmd", new { cmd }) ?? string.Empty;
        }
    }
}
