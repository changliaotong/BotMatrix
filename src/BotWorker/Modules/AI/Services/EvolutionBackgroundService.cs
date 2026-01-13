using BotWorker.Modules.AI.Interfaces;
using Microsoft.Extensions.DependencyInjection;
using Microsoft.Extensions.Hosting;
using Microsoft.Extensions.Logging;

namespace BotWorker.Modules.AI.Services
{
    public class EvolutionBackgroundService : BackgroundService
    {
        private readonly IServiceProvider _serviceProvider;
        private readonly ILogger<EvolutionBackgroundService> _logger;
        private readonly TimeSpan _checkInterval = TimeSpan.FromHours(6); // 每6小时检查一次进化

        public EvolutionBackgroundService(IServiceProvider serviceProvider, ILogger<EvolutionBackgroundService> logger)
        {
            _serviceProvider = serviceProvider;
            _logger = logger;
        }

        protected override async Task ExecuteAsync(CancellationToken stoppingToken)
        {
            _logger.LogInformation("[EvolutionBackgroundService] Starting...");

            while (!stoppingToken.IsCancellationRequested)
            {
                try
                {
                    using (var scope = _serviceProvider.CreateScope())
                    {
                        var evolutionService = scope.ServiceProvider.GetRequiredService<IEvolutionService>();
                        _logger.LogInformation("[EvolutionBackgroundService] Running evolution cycle...");
                        await evolutionService.EvolveAllJobsAsync();
                        _logger.LogInformation("[EvolutionBackgroundService] Evolution cycle completed.");
                    }
                }
                catch (Exception ex)
                {
                    _logger.LogError(ex, "[EvolutionBackgroundService] Error in evolution cycle");
                }

                await Task.Delay(_checkInterval, stoppingToken);
            }
        }
    }
}
