using BotWorker.Core.MetaDatas;

namespace BotWorker.Core.Data.AppsInfo
{
    public class TableInfo : MetaData<TableInfo>
    {
        public override string DataBase => "apps";
        public override string TableName => "table_info";
        public override string KeyField => "table_id";

        public int TableId { get; set; }
        public new string DbName { get; set; } = string.Empty; // Ĭ��ֵΪ���ַ���
        public string TableGroup { get; set; } = string.Empty; // Ĭ��ֵΪ���ַ���
        public string TableTitle { get; set; } = string.Empty; // Ĭ��ֵΪ���ַ���
        public int TableType { get; set; }
        public string OrderField { get; set; } = string.Empty; // Ĭ��ֵΪ���ַ���
        public string FilterField { get; set; } = string.Empty; // Ĭ��ֵΪ���ַ���
        public bool FilterAll { get; set; }
        public int? FilterType { get; set; }
        public string TableMemo { get; set; } = string.Empty; // Ĭ��ֵΪ���ַ���
        public bool IsOrder { get; set; }
        public DateTime? InsertDate { get; set; }
        public int? InsertBy { get; set; }
        public DateTime? UpdateDate { get; set; }
        public int? UpdateBy { get; set; }
    }
}


