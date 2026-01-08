using BotWorker.Core;
using BotWorker.Core.MetaDatas;

namespace BotWorker.Core.Data.AppsInfo
{
    public class ReportField : MetaData<ReportField>
    {
        public override string DataBase => "apps";
        public override string TableName => "report_field_info";
        public override string KeyField => "rpt_id";
        public override string KeyField2 => "field_name";

        public static string GetFieldCaption(string RptId, string FieldName)
        {
            return GetDef("field_caption", RptId, FieldName, FieldName);
        }
    }
}


