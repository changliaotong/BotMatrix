using Microsoft.Extensions.DependencyInjection;

namespace BotWorker.Infrastructure.Caching
{
    public static class CacheServiceExtensions
    {
        public static IServiceCollection AddCacheRepositories(this IServiceCollection services)
        {
            services.AddScoped(typeof(ICacheRepository<>), typeof(DefaultCacheRepository<>));
            return services;
        }
    }

}
