using System.Threading.Tasks;
using BotWorker.Common;
using BotWorker.Domain.Entities;
using BotWorker.Domain.Repositories;
using Dapper;

namespace BotWorker.Infrastructure.Persistence.Repositories
{
    public class CityRepository : BaseRepository<City>, ICityRepository
    {
        public CityRepository() : base("city", GlobalConfig.BaseInfoConnection)
        {
        }

        public async Task<City?> GetByNameAsync(string cityName)
        {
            return await GetFirstOrDefaultAsync("WHERE CityName = @cityName", new { cityName });
        }
    }
}
