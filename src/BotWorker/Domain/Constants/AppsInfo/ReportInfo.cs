namespace BotWorker.Domain.Constants.AppsInfo
{
    public class ReportInfo : MetaData<ReportInfo>
    {
        public override string DataBase => "apps";
        public override string TableName => "report_info";
        public override string KeyField => "rpt_id";
    }
}


