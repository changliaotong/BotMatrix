using Microsoft.EntityFrameworkCore;
using Microsoft.Extensions.Configuration;
using Microsoft.Extensions.DependencyInjection;
using BotWorker.Domain.Interfaces;

namespace BotWorker.Modules.Challenge
{
    public class ChallengeModule : IBotModule
    {
        public IModuleMetadata Metadata => new ChallengeMetadata();

        public void ConfigureDbContext(ModelBuilder modelBuilder)
        {
        }

        public void RegisterServices(IServiceCollection services, IConfiguration config)
        {
            services.AddSingleton<ChallengeService>();
        }
    }
}
