using BotWorker.Infrastructure.Persistence.ORM;

namespace BotWorker.Modules.Games
{
    public class Baby : MetaData<Baby>
    {
        [BotWorker.Infrastructure.Utils.Schema.Attributes.PrimaryKey]
        public Guid Id { get; set; } = Guid.NewGuid();
        public string UserId { get; set; } = string.Empty;
        public string Name { get; set; } = string.Empty;
        public DateTime Birthday { get; set; } = DateTime.Now;
        public int GrowthValue { get; set; } = 0;
        public int DaysOld { get; set; } = 0;
        public int Level { get; set; } = 1;
        public int Points { get; set; } = 0;
        public string Status { get; set; } = "active"; // active, abandoned
        public DateTime CreatedAt { get; set; } = DateTime.Now;
        public DateTime UpdatedAt { get; set; } = DateTime.Now;
        public DateTime LastDailyUpdate { get; set; } = DateTime.MinValue;

        public override string TableName => "Babies";
        public override string KeyField => "Id";

        public static async Task<Baby?> GetByUserIdAsync(string userId)
        {
            return (await QueryWhere("UserId = @p1 AND Status = 'active'", SqlParams(("@p1", userId)))).FirstOrDefault();
        }
    }

    public class BabyEvent : MetaData<BabyEvent>
    {
        [BotWorker.Infrastructure.Utils.Schema.Attributes.PrimaryKey]
        public Guid Id { get; set; } = Guid.NewGuid();
        public Guid BabyId { get; set; }
        public string EventType { get; set; } = string.Empty; // birthday, learn, work, interact
        public string Content { get; set; } = string.Empty;
        public DateTime CreatedAt { get; set; } = DateTime.Now;

        public override string TableName => "BabyEvents";
        public override string KeyField => "Id";
    }

    public class BabyConfig : MetaData<BabyConfig>
    {
        [BotWorker.Infrastructure.Utils.Schema.Attributes.PrimaryKey]
        public int Id { get; set; } = 1;
        public bool IsEnabled { get; set; } = true;
        public int GrowthRate { get; set; } = 1000;
        public DateTime UpdatedAt { get; set; } = DateTime.Now;

        public override string TableName => "BabyConfig";
        public override string KeyField => "Id";

        public static async Task<BabyConfig> GetAsync()
        {
            var config = (await QueryWhere("Id = 1")).FirstOrDefault();
            if (config == null)
            {
                config = new BabyConfig();
                await config.InsertAsync();
            }
            return config;
        }
    }
}
