using Microsoft.Extensions.DependencyInjection;

namespace sz84.Infrastructure.Caching
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
