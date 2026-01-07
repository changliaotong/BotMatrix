using Microsoft.EntityFrameworkCore;
using Microsoft.Extensions.Configuration;
using Microsoft.Extensions.DependencyInjection;
using BotWorker.Core.Interfaces;

namespace BotWorker.Bots.Models.Challenge
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
