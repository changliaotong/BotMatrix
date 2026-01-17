namespace BotWorker.Domain.Constants.AppsInfo
{
    public class FieldInfo
    {
        public int FieldId { get; set; }
        public int TableId { get; set; }
        public string DbName { get; set; } = string.Empty; 
        public string FieldName { get; set; } = string.Empty; 
        public string FieldCaption { get; set; } = string.Empty; 
        public string FieldMemo { get; set; } = string.Empty; 
        public int? FieldOrder { get; set; }
        public bool FieldRequire { get; set; }
        public bool Readonly { get; set; }
        public bool IsVisible { get; set; }
        public bool IsSelect { get; set; }
        public bool IsCurrency { get; set; }
        public int? FieldWidth { get; set; }
        public string FieldType { get; set; } = string.Empty; 
        public string LookupTable { get; set; } = string.Empty; 
        public string LookupKeyFields { get; set; } = string.Empty; 
        public string LookupResultField { get; set; } = string.Empty; 
        public string LookupFilter { get; set; } = string.Empty; 
        public int? HandleId { get; set; }
        public DateTime? InsertDate { get; set; }
        public int? InsertBy { get; set; }
        public DateTime? UpdateDate { get; set; }
        public int? UpdateBy { get; set; }
    }
}


