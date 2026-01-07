namespace sz84.Core.MetaDatas
{
    [AttributeUsage(AttributeTargets.Property)]
    public class ColumnAttribute(string? name = null) : Attribute
    {
        public string? Name { get; } = name;
        public object? ConverterType { get; internal set; }
    }

    [AttributeUsage(AttributeTargets.Property)]
    public class TransientAttribute(bool ignoreInsert = true, bool ignoreUpdate = true) : Attribute
    {
        public bool IgnoreInsert { get; } = ignoreInsert;
        public bool IgnoreUpdate { get; } = ignoreUpdate;
    }

    [AttributeUsage(AttributeTargets.Property)]
    public class FieldBehaviorAttribute : Attribute
    {
        public bool IncludeInInsert { get; }
        public bool IncludeInUpdate { get; }
        public bool IncludeInSelect { get; }

        public FieldBehaviorAttribute(
            bool includeInInsert = true,
            bool includeInUpdate = true,
            bool includeInSelect = true)
        {
            IncludeInInsert = includeInInsert;
            IncludeInUpdate = includeInUpdate;
            IncludeInSelect = includeInSelect;
        }

        // 快速定义忽略一切的行为
        public static readonly FieldBehaviorAttribute IgnoreAll = new(false, false, false);
    }
}
