namespace BotWorker.Domain.Constants.AppsInfo
{
    public class TableInfo
    {
        public int TableId { get; set; }
        public string DbName { get; set; } = string.Empty; 
        public string TableGroup { get; set; } = string.Empty; 
        public string TableTitle { get; set; } = string.Empty; 
        public int TableType { get; set; }
        public string OrderField { get; set; } = string.Empty; 
        public string FilterField { get; set; } = string.Empty; 
        public bool FilterAll { get; set; }
        public int? FilterType { get; set; }
        public string TableMemo { get; set; } = string.Empty; 
        public bool IsOrder { get; set; }
        public DateTime? InsertDate { get; set; }
        public int? InsertBy { get; set; }
        public DateTime? UpdateDate { get; set; }
        public int? UpdateBy { get; set; }
    }
}


