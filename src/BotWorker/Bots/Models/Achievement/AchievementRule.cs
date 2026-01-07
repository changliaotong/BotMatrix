namespace BotWorker.Bots.Models.Achievement
{
    public class AchievementRule
    {
        public int Level { get; set; }
        public int RequiredCount { get; set; }          // 达到多少累计值升级
        public int? RequiredStreakDays { get; set; }    // 连续天数条件（可选）
        public DateTime? ValidFrom { get; set; }        // 时间区间限制（可选）
        public DateTime? ValidTo { get; set; }
        public string Reward { get; set; } = string.Empty; // 奖励字符串
    }

}
