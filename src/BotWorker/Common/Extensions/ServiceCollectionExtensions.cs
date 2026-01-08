using Microsoft.Extensions.DependencyInjection;

namespace BotWorker.BotWorker.BotWorker.Common.Exts
{
    public static class ServiceCollectionExtensions
    {
        public static IServiceCollection AddIfNotRegistered<TService, TImpl>(this IServiceCollection services, ServiceLifetime lifetime = ServiceLifetime.Scoped)
            where TService : class
            where TImpl : class, TService
        {
            if (!services.Any(s => s.ServiceType == typeof(TService)))
            {
                var descriptor = new ServiceDescriptor(typeof(TService), typeof(TImpl), lifetime);
                services.Add(descriptor);
            }
            return services;
        }
    }
}


