using BotWorker.Infrastructure.Persistence.ORM;

namespace BotWorker.Domain.Entities
{
    //问答系统
    public partial class AnswerInfo : MetaDataGuid<AnswerInfo>
    {
        public override string TableName => "Answer";
        public override string KeyField => "Id";
        public long QuestionId { get; set; }
        public string Question { get; set; } = string.Empty;
        public string Answer { get; set; } = string.Empty;
        public string AnswerBak { get; set; } = string.Empty;
        [DbIgnore]
        public DateTime InsertDate { get; set; }
        public long UserId { get; set; }        
        public long GroupId { get; set; }
        public long RobotId { get; set; }
        public long BotUin { get; set; }
        public int IsOnly { get; set; }
        public int Audit { get; set; }
        public long AuditBy { get; set; }
        public DateTime AuditDate { get; set; }
        public int Audit2 { get; set; }
        public long Audit2By { get; set; }
        public DateTime Audit2Date { get; set; }        
        public string Audit2Info { get; set; } = string.Empty;
        public DateTime UpdateDate { get; set; }
        [DbIgnore]
        public int UsedTimes { get; set; }
        [DbIgnore]
        public int GoonTimes { get; set; }
        [DbIgnore]
        public int UsedTimesGroup { get; set; }
        [DbIgnore]
        public int GoonTimesGroup { get; set; }
        [DbIgnore]
        public int Credit { get; set; }
        public static long Rid => BotInfo.DefaultRobotId;     
    }
}
