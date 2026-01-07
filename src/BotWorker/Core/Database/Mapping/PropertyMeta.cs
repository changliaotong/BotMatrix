using System.Reflection;
using BotWorker.Core.MetaDatas;

namespace BotWorker.Core.Database.Mapping
{
    public class PropertyMeta
    {
        public string PropertyName { get; }
        public string ColumnName { get; }
        public PropertyInfo Property { get; }

        public bool IncludeInInsert { get; }
        public bool IncludeInUpdate { get; }
        public bool IncludeInSelect { get; }
        public object PropertyInfo { get; internal set; }

        public PropertyMeta(PropertyInfo property)
        {
            Property = property;
            PropertyName = property.Name;

            var colAttr = property.GetCustomAttribute<ColumnAttribute>();
            ColumnName = colAttr?.Name ?? PropertyName;

            var behavior = property.GetCustomAttribute<FieldBehaviorAttribute>() ?? new FieldBehaviorAttribute();

            IncludeInInsert = behavior.IncludeInInsert;
            IncludeInUpdate = behavior.IncludeInUpdate;
            IncludeInSelect = behavior.IncludeInSelect; 
            PropertyInfo = new object();
        }

        public object? GetValue(object obj) => Property.GetValue(obj);
        public void SetValue(object obj, object? value) => Property.SetValue(obj, value);
    }
}
