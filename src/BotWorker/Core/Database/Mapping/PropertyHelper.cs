using System.Collections.Concurrent;
using System.Reflection;
using System.Text.Json.Serialization;

namespace BotWorker.Core.Database.Mapping
{
    public enum FieldControlType
    {
        Text,
        TextArea,
        Date,
        Dropdown,
        Checkbox,
        Number
    }

    public class PropertyExtended
    {
        public string Name { get; set; } = "";
        public string DisplayName { get; set; } = "";
        public bool IsKey { get; set; }
        public bool IncludeInList { get; set; }
        public bool IncludeInEdit { get; set; }
        public FieldControlType ControlType { get; set; } = FieldControlType.Text;
        public Dictionary<string, string>? DropdownOptions { get; set; } = null; // Key-Value 选项
    }

    public static class PropertyHelper
    {
        private static readonly ConcurrentDictionary<Type, List<PropertyMeta>> _cache = new();

        public static List<PropertyMeta> GetAll(Type type)
        {
            return _cache.GetOrAdd(type, t =>
                t.GetProperties(BindingFlags.Public | BindingFlags.Instance)
                 .Where(p => p.CanRead && p.CanWrite &&
                             p.GetCustomAttribute<JsonIgnoreAttribute>() == null)
                 .Select(p => new PropertyMeta(p))
                 .ToList()
            );
        }

        public static List<PropertyMeta> GetAll<T>() => GetAll(typeof(T));

        public static List<PropertyExtended> GetExtendedProperties(Type type)
        {
            var metas = GetAll(type);

            var extended = metas.Select(meta =>
            {
                var propInfo = meta.PropertyInfo;

                return new PropertyExtended
                {
                };
            }).ToList();

            return extended;
        }
    }
}
