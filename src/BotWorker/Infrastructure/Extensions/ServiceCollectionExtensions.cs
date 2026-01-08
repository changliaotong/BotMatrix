using Microsoft.Extensions.DependencyInjection;
using BotWorker.Modules.Achievement;
using BotWorker.Domain.Interfaces;
using BotWorker.Modules.Ranking;

namespace BotWorker.Infrastructure.Extensions
{
    public static class ServiceCollectionExtensions
    {
        public static IServiceCollection AddGameModules(this IServiceCollection services)
        {
            // Achievement Service
            services.AddSingleton<BotWorker.Modules.Achievement.IAchievementService, BotWorker.Modules.Achievement.AchievementService>();
            // 排行榜系统
            services.AddSingleton<IRankingService, RankingManager>(); 

            // 注册预设成就（可拆到成就系统内）
            services.PostConfigure<AchievementManager>(manager =>
            {
                //AchievementPresets.RegisterAll(manager);
            });

            return services;
        }
    }
}


