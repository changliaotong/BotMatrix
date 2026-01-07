using Microsoft.Extensions.Logging;

namespace sz84.Infrastructure.Caching
{
    public class LoggerAdapter : ICacheLogger
    {
        private readonly ILogger _logger;

        public LoggerAdapter(ILogger logger)
        {
            _logger = logger ?? throw new ArgumentNullException(nameof(logger));
        }

        public void LogDebug(string message, params object[] args)
        {
            _logger.LogDebug(message, args);
        }

        public void LogInformation(string message, params object[] args)
        {
            _logger.LogInformation(message, args);
        }

        public void LogWarning(string message, params object[] args)
        {
            _logger.LogWarning(message, args);
        }

        public void LogError(string message, Exception? ex = null, params object[] args)
        {
            if (ex == null)
                _logger.LogError(message, args);
            else
                _logger.LogError(ex, message, args);
        }
    }

}
