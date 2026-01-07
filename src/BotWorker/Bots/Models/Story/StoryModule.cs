using Microsoft.EntityFrameworkCore;
using Microsoft.Extensions.Configuration;
using Microsoft.Extensions.DependencyInjection;
using BotWorker.Core.Interfaces;

namespace BotWorker.Bots.Models.Story
{
    public class StoryModule : IBotModule
    {
        public IModuleMetadata Metadata => throw new NotImplementedException();

        public void ConfigureDbContext(ModelBuilder modelBuilder)
        {
            throw new NotImplementedException();
        }

        public void RegisterServices(IServiceCollection services, IConfiguration config)
        {
            throw new NotImplementedException();
        }
    }

}
