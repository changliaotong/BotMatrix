using System;

namespace BotWorker.Modules.Quest
{
    public class UserQuestProgress
    {
        public long UserId { get; set; }
        public string QuestId { get; set; } = string.Empty;
        public int CurrentValue { get; set; }
        public bool IsCompleted { get; set; }
        public DateTime LastUpdated { get; set; }
    }
}


