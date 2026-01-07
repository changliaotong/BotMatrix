namespace sz84.Infrastructure.Caching
{
    public interface ICacheLogger
    {
        void LogDebug(string message, params object[] args);
        void LogInformation(string message, params object[] args);
        void LogWarning(string message, params object[] args);
        void LogError(string message, Exception? ex = null, params object[] args);
    }


}
