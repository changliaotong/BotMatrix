namespace BotWorker.Bots.Models.Achievement
{
    public class Achievement
    {
        public string Id { get; set; } = string.Empty;               // 成就唯一 ID
        public string Title { get; set; } = string.Empty;            // 成就名称
        public string Description { get; set; } = string.Empty;      // 成就说明
        public int MaxLevel { get; set; } = 1;                       // 成就等级（0 表示隐藏/一次性）
        public string Category { get; set; } = string.Empty;         // 所属分类
        public string IconUrl { get; set; } = string.Empty;          // 成就图标
        public string Reward { get; set; } = string.Empty;           // 奖励内容，如积分、头衔、称号
        public List<AchievementRule> Rules { get; set; } = [];
        public int RequiredCount { get; internal set; }
        public long RewardCredit { get; internal set; }
        public string CounterKey { get; internal set; } = string.Empty;
    }
}
