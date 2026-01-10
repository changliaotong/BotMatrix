namespace BotWorker.Modules.AI.Models
{
    public class LLMCallLog
    {
        public long Id { get; set; }
        public Guid? AgentId { get; set; }
        public Guid? ConsumerUserId { get; set; }
        public Guid? OwnerUserId { get; set; }

        public string ModelId { get; set; } = "";
        public string ProviderId { get; set; } = "";

        public int InputTokens { get; set; }
        public int OutputTokens { get; set; }
        public decimal InputPrice { get; set; }
        public decimal OutputPrice { get; set; }

        public decimal TotalCost => InputPrice + OutputPrice;
        //public decimal Earnings => Math.Round(TotalCost * PlatformConfig.AgentSplitRatio, 4);

        public bool IsSuccess { get; set; }
        public DateTime CreatedAt { get; set; }
    }

}
