using BotWorker.Infrastructure.Persistence.ORM;

namespace BotWorker.Modules.AI.Models
{
    public class AgentTag: MetaDataGuid<AgentTag>
    {
        public override string TableName => "AgentTag";
        public override string KeyField => "Id";

        public string Name { get; set; } = string.Empty;
        public string Description { get; set; } = string.Empty;    
        public long UserId { get; set; }
        public DateTime CreateAt { get; set; } 
        public DateTime UpdateAt { get; set; } 
    }

    public class AgentTags : MetaData<AgentTags>
    {
        public override string TableName => "AgentTags";
        public override string KeyField => "AgentId";
        public override string KeyField2 => "TagId";

        public long AgentId { get; set; } = 0;
        public long TagId { get; set; } = 0;
        public long UserId { get; set; } = 0;
        public DateTime CreateAt { get; set; } = DateTime.MinValue;
    }
}
