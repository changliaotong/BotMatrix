using Microsoft.EntityFrameworkCore;
using Microsoft.Extensions.Configuration;
using Microsoft.Extensions.DependencyInjection;
using Serilog;
using sz84.Bots.Models.Limiter;
using BotWorker.Domain.Interfaces;

namespace sz84.Bots.Models.Gacha
{
    public class GachaService(ILimiter dailyLimiter) : IBotModule
    {
        ILimiter _dailyLimiter = dailyLimiter;

        public IModuleMetadata Metadata => new GachaMetadata();
            
        public void RegisterServices(IServiceCollection services, IConfiguration config)
        {
            services.AddSingleton<GachaService>();
            InfoMessage($"✅ [{nameof(GachaService)}] 注册成功");
        }

        public void ConfigureDbContext(ModelBuilder modelBuilder)
        {
            modelBuilder.Entity<GachaCardRecord>();
        }
    }
}
