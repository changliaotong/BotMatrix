using Microsoft.Extensions.DependencyInjection;
using BotWorker.Bots.Models.Achievement;
using BotWorker.Bots.Models.Ranking;

namespace BotWorker.Bots.Extensions
{
    public static class ServiceCollectionExtensions
    {
        public static IServiceCollection AddGameModules(this IServiceCollection services)
        {
            // Infrastructure（内存实现版本）
            services.AddSingleton<IAchievementService, AchievementService>();
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
