using Microsoft.Extensions.Configuration;
using Microsoft.Extensions.DependencyInjection;

namespace BotWorker.Domain.Interfaces
{
    public interface IBotModule
    {
        IModuleMetadata Metadata { get; }

        void RegisterServices(IServiceCollection services, IConfiguration config);
    }
}
