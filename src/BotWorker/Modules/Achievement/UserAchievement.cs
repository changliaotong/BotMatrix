namespace BotWorker.Modules.Achievement
{
    public class UserAchievement
    {
        public long UserId { get; set; }
        public string AchievementId { get; set; } = null!;

        public int CurrentLevel { get; set; }
        public int CurrentValue { get; set; }

        public DateTime LastActionDate { get; set; }

        public int CurrentStreakDays { get; set; }      // 连续天数计数
        public DateTime? LastActionDatePrev { get; set; } // 上一次操作时间，便于判断连续天数
        public DateTime LastUpdated { get; internal set; }
    }

}
