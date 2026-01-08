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
            // Infrastructure（内存实现版本）
            services.AddSingleton<IAchievementService, AchievementService>();
            // 排行榜系�?            services.AddSingleton<IRankingService, RankingManager>(); 

            // 注册预设成就（可拆到成就系统内）
            services.PostConfigure<AchievementManager>(manager =>
            {
                //AchievementPresets.RegisterAll(manager);
            });

            return services;
        }
    }
}


