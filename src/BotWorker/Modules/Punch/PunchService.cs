using Microsoft.EntityFrameworkCore;
using Microsoft.Extensions.Configuration;
using Microsoft.Extensions.DependencyInjection;
using BotWorker.Core.Interfaces;

namespace BotWorker.Domain.Entities.Punch
{
    public class PunchService : IBotModule
    {
        public IModuleMetadata Metadata => new PunchMetadata();

        public void ConfigureDbContext(ModelBuilder modelBuilder)
        {            
        }

        public string Punch()
        {
            int reward = new Random().Next(5, 15);
            //_credit.AddCredit(reward, "揍群主");
            return $"👊 你狠狠揍了群主，获得 {reward} 积分！当前总积分：";
        }

        public void RegisterServices(IServiceCollection services, IConfiguration config)
        {
            services.AddSingleton<PunchService>();
            InfoMessage($"✅ [{nameof(PunchService)}] 注册成功");
        }
    }

}
