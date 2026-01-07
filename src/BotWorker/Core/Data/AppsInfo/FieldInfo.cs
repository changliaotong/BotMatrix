using sz84.Core.MetaDatas;

namespace sz84.Core.Data.AppsInfo
{
    public class FieldInfo : MetaData<FieldInfo>
    {
        public override string DataBase => "apps";
        public override string TableName => "field_info";
        public override string KeyField => "table_id";
        public override string KeyField2 => "field_name";

        public int FieldId { get; set; }
        public int TableId { get; set; }
        public new string DbName { get; set; } = string.Empty; // 默认值为空字符串
        public string FieldName { get; set; } = string.Empty; // 默认值为空字符串
        public string FieldCaption { get; set; } = string.Empty; // 默认值为空字符串
        public string FieldMemo { get; set; } = string.Empty; // 默认值为空字符串
        public int? FieldOrder { get; set; }
        public bool FieldRequire { get; set; }
        public bool Readonly { get; set; }
        public bool IsVisible { get; set; }
        public bool IsSelect { get; set; }
        public bool IsCurrency { get; set; }
        public int? FieldWidth { get; set; }
        public string FieldType { get; set; } = string.Empty; // 默认值为空字符串
        public string LookupTable { get; set; } = string.Empty; // 默认值为空字符串
        public string LookupKeyFields { get; set; } = string.Empty; // 默认值为空字符串
        public string LookupResultField { get; set; } = string.Empty; // 默认值为空字符串
        public string LookupFilter { get; set; } = string.Empty; // 默认值为空字符串
        public int? HandleId { get; set; }
        public DateTime? InsertDate { get; set; }
        public int? InsertBy { get; set; }
        public DateTime? UpdateDate { get; set; }
        public int? UpdateBy { get; set; }

        public static string GetFieldCaption(string TableId, string FieldName)
        {
            return GetDef("field_caption", TableId, FieldName, FieldName);
        }
    }
}
