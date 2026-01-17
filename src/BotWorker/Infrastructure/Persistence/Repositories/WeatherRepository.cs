using System;
using System.Threading.Tasks;
using BotWorker.Domain.Entities;
using BotWorker.Domain.Repositories;
using Dapper;

namespace BotWorker.Infrastructure.Persistence.Repositories
{
    public class WeatherRepository : BaseRepository<Weather>, IWeatherRepository
    {
        public WeatherRepository(string? connectionString = null)
            : base("Weather", connectionString ?? GlobalConfig.BaseInfoConnection)
        {
        }

        public async Task<string?> GetRecentWeatherAsync(string cityName, int hours)
        {
            // PostgreSQL syntax for DATEDIFF(HOUR, GETDATE(), InsertDate) < hours
            // EXTRACT(EPOCH FROM (CURRENT_TIMESTAMP - InsertDate)) / 3600 < hours
            string sql = $@"
                SELECT WeatherInfo 
                FROM {_tableName} 
                WHERE CityName = @cityName 
                AND ABS(EXTRACT(EPOCH FROM (CURRENT_TIMESTAMP - InsertDate)) / 3600) < @hours 
                ORDER BY Id DESC 
                LIMIT 1";
            
            using var conn = CreateConnection();
            return await conn.QueryFirstOrDefaultAsync<string>(sql, new { cityName, hours });
        }
    }
}
