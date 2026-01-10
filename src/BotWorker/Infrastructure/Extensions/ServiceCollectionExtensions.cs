using Microsoft.Extensions.DependencyInjection;

namespace BotWorker.Infrastructure.Extensions
{
    public static class ServiceCollectionExtensions
    {
        public static IServiceCollection AddGameModules(this IServiceCollection services)
        {
            return services;
        }
    }
}


