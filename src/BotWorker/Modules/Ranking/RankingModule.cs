using Microsoft.EntityFrameworkCore;
using Microsoft.Extensions.Configuration;
using Microsoft.Extensions.DependencyInjection;
using BotWorker.Core.Interfaces;

namespace BotWorker.Domain.Entities.Ranking
{
    public class RankingModule : IBotModule
    {
        public string Name => "Ranking";

        public IModuleMetadata Metadata => throw new NotImplementedException();

        public void ConfigureDbContext(ModelBuilder modelBuilder)
        {
            throw new NotImplementedException();
        }

        public void RegisterServices(IServiceCollection services, IConfiguration config)
        {
            services.AddSingleton<IRankingService, RankingManager>();
        }
    }

}
