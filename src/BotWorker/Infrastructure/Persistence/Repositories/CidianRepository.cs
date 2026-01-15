using System.Collections.Generic;
using System.Threading.Tasks;
using BotWorker.Common;
using BotWorker.Domain.Entities;
using BotWorker.Domain.Repositories;
using Dapper;

namespace BotWorker.Infrastructure.Persistence.Repositories
{
    public class CidianRepository : BaseRepository<Cidian>, ICidianRepository
    {
        public CidianRepository() : base("ciba", GlobalConfig.BaseInfoConnection)
        {
        }

        public async Task<string> GetDescriptionAsync(string keyword)
        {
            return await GetValueAsync<string>("Description", "WHERE Keyword = @keyword", new { keyword }) ?? string.Empty;
        }

        public async Task<IEnumerable<Cidian>> SearchAsync(string keyword, int limit = 20)
        {
            return await GetListAsync("WHERE Keyword LIKE @keyword LIMIT @limit", new { keyword = $"{keyword}%", limit });
        }
    }
}
