using sz84.Core.Interfaces;

namespace sz84.Bots.Models.Limiter
{
    public class LimiterMetadata : IModuleMetadata
    {
        public string Name => "Limit";
        public string Version => "1.0";
        public string Author => "derlin";
        public string Description => "周期限制";
    }
}
