using Microsoft.EntityFrameworkCore;
using Microsoft.Extensions.Configuration;
using Microsoft.Extensions.DependencyInjection;
using BotWorker.Domain.Interfaces;

namespace BotWorker.Modules.Ranking
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
