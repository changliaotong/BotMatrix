using Microsoft.EntityFrameworkCore;
using Microsoft.Extensions.Configuration;
using Microsoft.Extensions.DependencyInjection;
using BotWorker.Core.Interfaces;

namespace BotWorker.Bots.Models.PvP
{
    public class PvPModule : IBotModule
    {
        public IModuleMetadata Metadata => new PvPMetadata();

        public void ConfigureDbContext(ModelBuilder modelBuilder)
        {
            throw new NotImplementedException();
        }

        public void RegisterServices(IServiceCollection services, IConfiguration config)
        {
            services.AddSingleton<PvPService>();
            Console.WriteLine("✅ [PvPModule] 注册成功");
        }
    }
}
