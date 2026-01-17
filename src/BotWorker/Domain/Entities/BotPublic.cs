using Dapper.Contrib.Extensions;

namespace BotWorker.Domain.Entities
{
    [Table("public")]
    public class BotPublic
    {
        [ExplicitKey]
        public string PublicKey { get; set; } = string.Empty;
        public string PublicName { get; set; } = string.Empty;
        public long GroupId { get; set; }
        public long BotUin { get; set; }
        public long AdminId { get; set; }
    }
}
