using BotWorker.Infrastructure.Persistence.ORM;
using BotWorker.Infrastructure.Utils.Schema.Attributes;
using System.Text;

namespace BotWorker.Modules.Games
{
    public class CultivationProfile : MetaDataGuid<CultivationProfile>
    {
        public string UserId { get; set; } = string.Empty;

        public int Level { get; set; } = 1;

        public long Exp { get; set; } = 0;

        public long MaxExp { get; set; } = 100;

        public int CultivationSpeed { get; set; } = 10;

        public DateTime LastCultivateTime { get; set; } = DateTime.MinValue;

        public override string TableName => "CultivationProfiles";
        public override string KeyField => "Id";

        public static async Task<CultivationProfile?> GetByUserIdAsync(string userId)
        {
            return (await QueryWhere("UserId = @p1", SqlParams(("@p1", userId)))).FirstOrDefault();
        }

        public static async Task<List<CultivationProfile>> GetTopCultivatorsAsync(int limit = 10)
        {
            return await QueryWhere("1=1 ORDER BY Level DESC, Exp DESC LIMIT @p1", SqlParams(("@p1", limit)));
        }

        public string GetStageName()
        {
            return Level switch
            {
                < 10 => "炼气期",
                < 20 => "筑基期",
                < 30 => "金丹期",
                < 40 => "元婴期",
                < 50 => "化神期",
                < 60 => "炼虚期",
                < 70 => "合体期",
                < 80 => "大乘期",
                < 90 => "渡劫期",
                _ => "飞升成仙"
            };
        }

        public string GetRankDescription()
        {
            int subLevel = (Level - 1) % 10 + 1;
            return $"{GetStageName()} {subLevel} 层";
        }
    }

    public class CultivationRecord : MetaDataGuid<CultivationRecord>
    {
        public string UserId { get; set; } = string.Empty;

        public string ActionType { get; set; } = string.Empty; // 修炼, 突破, 走火入魔

        public string Detail { get; set; } = string.Empty;

        public DateTime CreateTime { get; set; } = DateTime.Now;

        public override string TableName => "CultivationRecords";
        public override string KeyField => "Id";
    }
}
