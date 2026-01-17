using Dapper.Contrib.Extensions;

namespace BotWorker.Domain.Entities
{
    [Table("handle_question")]
    public class HandleQuestion
    {
        [Key]
        public long Id { get; set; }
        public long Qid { get; set; }
        public string Question { get; set; } = "";
        public long? Qid2 { get; set; }
        public string Question2 { get; set; } = "";
        public double Score { get; set; }        
        public int UsedTimes { get; set; }
        public int UsedTimes2 { get; set; }
        public int CAnswerAll { get; set; }
        public string Answers { get; set; } = "";
    }
}
