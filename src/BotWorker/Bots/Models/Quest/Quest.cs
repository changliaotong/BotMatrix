namespace sz84.Bots.Models.Quest
{
    public enum QuestType
    {
        Daily,
        Weekly,
        Event
    }

    public class Quest
    {
        public string Id { get; set; } = string.Empty;
        public string Title { get; set; } = string.Empty;
        public string Description { get; set; } = string.Empty;
        public QuestType Type { get; set; }         // Daily, Weekly, Event
        public int TargetValue { get; set; }
        public string Reward { get; set; } = string.Empty;          // 积分、道具、称号等
    }

}
