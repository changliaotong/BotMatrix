using Microsoft.EntityFrameworkCore;
using Microsoft.Extensions.Configuration;
using Microsoft.Extensions.DependencyInjection;
using Serilog;
using BotWorker.Bots.Models.Limiter;
using BotWorker.Core.Interfaces;

namespace BotWorker.Bots.Models.JustKidding
{
    public class JustKiddingService(ILimiter dailyLimiter) : IBotModule
    {
        ILimiter _dailyLimiter = dailyLimiter;

        public IModuleMetadata Metadata => new JustKiddingMetadata();

        public void RegisterServices(IServiceCollection services, IConfiguration config)
        {
            services.AddSingleton<JustKiddingService>();
            InfoMessage($"✅ [{nameof(JustKiddingService)}] 注册成功");
        }

        public void ConfigureDbContext(ModelBuilder modelBuilder)
        {
            //modelBuilder.Entity<GachaCardRecord>();
        }
    }
}
