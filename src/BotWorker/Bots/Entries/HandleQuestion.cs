using sz84.Core.MetaDatas;

namespace sz84.Bots.Entries
{
    public class HandleQuestion :MetaData<HandleQuestion>
    {
        public long Qid { get; set; }
        public string Question { get; set; } = "";
        public long? Qid2 { get; set; }
        public string Question2 { get; set; } = "";
        public double Score { get; set; }        
        public int UsedTimes { get; set; }
        public int UsedTimes2 { get; set; }
        public int CAnswerAll { get; set; }
        public string Answers { get; set; } = "";

        public override string TableName => "HandleQuestion";

        public override string KeyField => "Id";
    }
}
